package gocoin

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"go.etcd.io/bbolt"
)

// Peer represents one TCP connection
type Peer struct {
	conn   net.Conn
	node   *Node
	reader *bufio.Reader
	writer *bufio.Writer

	// handshake state
	gotVersion     bool
	gotVerack      bool
	sendVersionAck bool
}

func NewPeer(conn net.Conn, node *Node) *Peer {
	return &Peer{
		conn:   conn,
		node:   node,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

// Run blocks until disconnect
func (p *Peer) Run() error {
	// send our version immediately
	height := int32(0)
	p.node.chain.DB().View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		heightBytes := b.Get([]byte("l"))
		if len(heightBytes) == 32 {
			// crude height: count back to genesis
			iter := p.node.chain.Iterator()
			for {
				blk := iter.Next()
				if blk == nil {
					break
				}
				height++
				if len(blk.PrevBlockHash) == 0 {
					break
				}
			}
		}
		return nil
	})
	vp := VersionPayload{
		Version:   protocolVersion,
		Agent:     "gocoin/0.4",
		Height:    height,
		Services:  []string{"full"},
		Timestamp: time.Now().Unix(),
	}
	if err := p.Send(MustWrap(vp)); err != nil {
		return err
	}

	for {
		var msg Message
		if err := Decode(p.reader, &msg); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := p.handle(msg); err != nil {
			return err
		}
	}
}

// Send a message
func (p *Peer) Send(msg Message) error {
	if err := Encode(p.writer, msg); err != nil {
		return err
	}
	return p.writer.Flush()
}

func (p *Peer) handle(msg Message) error {
	switch msg.Type {
	case "version":
		var vp VersionPayload
		if err := json.Unmarshal(msg.Payload, &vp); err != nil {
			return err
		}
		p.gotVersion = true
		// send verack
		return p.Send(MustWrap(VerackPayload{}))
	case "verack":
		p.gotVerack = true
	case "getblocks":
		var gp GetBlocksPayload
		if err := json.Unmarshal(msg.Payload, &gp); err != nil {
			return err
		}
		return p.handleGetBlocks(gp)
	case "inv":
		var inv InvPayload
		if err := json.Unmarshal(msg.Payload, &inv); err != nil {
			return err
		}
		return p.handleInv(inv)
	case "getdata":
		var gd GetDataPayload
		if err := json.Unmarshal(msg.Payload, &gd); err != nil {
			return err
		}
		return p.handleGetData(gd)
	case "block":
		var bp BlockPayload
		if err := json.Unmarshal(msg.Payload, &bp); err != nil {
			return err
		}
		return p.handleBlock(bp)
	case "tx":
		var tp TxPayload
		if err := json.Unmarshal(msg.Payload, &tp); err != nil {
			return err
		}
		return p.handleTx(tp)
	case "getmempool":
		return p.handleGetMempool()
	case "mempool":
		var mp MempoolPayload
		if err := json.Unmarshal(msg.Payload, &mp); err != nil {
			return err
		}
		return p.handleMempool(mp)
	default:
		return fmt.Errorf("unknown message type %s", msg.Type)
	}
	return nil
}

func (p *Peer) handleGetBlocks(gp GetBlocksPayload) error {
	// crude: send all block hashes after genesis
	var hashes [][]byte
	p.node.chain.DB().View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if len(k) == 32 {
				hashes = append(hashes, k)
			}
		}
		return nil
	})
	inv := InvPayload{}
	for _, h := range hashes {
		inv.Items = append(inv.Items, InvItem{Type: "block", Hash: h})
	}
	return p.Send(MustWrap(inv))
}

func (p *Peer) handleInv(inv InvPayload) error {
	// request everything we don't have
	var gd GetDataPayload
	for _, it := range inv.Items {
		switch it.Type {
		case "block":
			if !p.hasBlock(it.Hash) {
				gd.Items = append(gd.Items, it)
			}
		case "tx":
			if !p.node.txPool.Has(it.Hash) {
				gd.Items = append(gd.Items, it)
			}
		}
	}
	if len(gd.Items) > 0 {
		return p.Send(MustWrap(gd))
	}
	return nil
}

func (p *Peer) handleGetData(gd GetDataPayload) error {
	for _, it := range gd.Items {
		switch it.Type {
		case "block":
			var blk *Block
			p.node.chain.DB().View(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(blocksBucket))
				blkBytes := b.Get(it.Hash)
				if blkBytes != nil {
					blk = DeserializeBlock(blkBytes)
				}
				return nil
			})
			if blk != nil {
				jb := blkToJSON(blk)
				if err := p.Send(MustWrap(BlockPayload{Block: jb})); err != nil {
					return err
				}
			}
		case "tx":
			tx := p.node.txPool.Get(it.Hash)
			if tx != nil {
				jtx := txToJSON(tx)
				if err := p.Send(MustWrap(TxPayload{Tx: jtx})); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *Peer) handleBlock(bp BlockPayload) error {
	blk := jsonToBlock(bp.Block)
	// very simple: if it extends our tip, accept
	tip := p.node.chain.Iterator().Next()
	if tip != nil && bytes.Equal(blk.PrevBlockHash, tip.Hash) {
		p.node.chain.MineBlock(blk.Transactions) // re-use existing logic
	}
	return nil
}

func (p *Peer) handleTx(tp TxPayload) error {
	tx := jsonToTx(tp.Tx)
	if p.node.chain.VerifyTransaction(tx) {
		p.node.txPool.Add(tx)
		// relay
		inv := InvPayload{Items: []InvItem{{Type: "tx", Hash: tx.ID}}}
		p.node.broadcast(inv)
	}
	return nil
}

func (p *Peer) handleGetMempool() error {
	hashes := p.node.txPool.GetAllHashes()
	inv := InvPayload{}
	for _, h := range hashes {
		inv.Items = append(inv.Items, InvItem{Type: "tx", Hash: h})
	}
	return p.Send(MustWrap(inv))
}

func (p *Peer) handleMempool(mp MempoolPayload) error {
	// for now just ignore, could batch request
	return nil
}

func (p *Peer) hasBlock(hash []byte) bool {
	var exists bool
	p.node.chain.DB().View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b.Get(hash) != nil {
			exists = true
		}
		return nil
	})
	return exists
}

// helper to convert Block -> JSONBlock
func blkToJSON(b *Block) JSONBlock {
	jb := JSONBlock{
		Hash:          hex.EncodeToString(b.Hash),
		PrevBlockHash: hex.EncodeToString(b.PrevBlockHash),
		Timestamp:     b.Timestamp,
		Nonce:         b.Nonce,
		Transactions:  make([]JSONTx, len(b.Transactions)),
	}
	for i, tx := range b.Transactions {
		jb.Transactions[i] = txToJSON(tx)
	}
	return jb
}

func jsonToBlock(jb JSONBlock) *Block {
	blk := &Block{
		Hash:          make([]byte, 32),
		PrevBlockHash: make([]byte, 32),
		Timestamp:     jb.Timestamp,
		Nonce:         jb.Nonce,
		Transactions:  make([]*Transaction, len(jb.Transactions)),
	}
	copy(blk.Hash, jb.Hash)
	copy(blk.PrevBlockHash, jb.PrevBlockHash)
	for i, jtx := range jb.Transactions {
		blk.Transactions[i] = jsonToTx(jtx)
	}
	return blk
}
