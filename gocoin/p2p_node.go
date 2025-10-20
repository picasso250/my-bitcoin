package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	protocolVersion = 1
	maxPeers        = 25
)

// Node is the TCP P2P server
type Node struct {
	listener net.Listener
	chain    *Blockchain
	txPool   *TxPool

	peerMu sync.RWMutex
	peers  map[string]*Peer // remoteAddr -> Peer
}

// NewNode creates and starts a listening node
func NewNode(listenAddr string, chain *Blockchain) (*Node, error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	n := &Node{
		listener: ln,
		chain:    chain,
		txPool:   NewTxPool(),
		peers:    make(map[string]*Peer),
	}
	go n.acceptLoop()
	return n, nil
}

// Connect to a seed peer (blocking)
func (n *Node) Connect(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	go n.handleConn(conn, true)
	return nil
}

// BroadcastTx to all peers (non-blocking best-effort)
func (n *Node) BroadcastTx(tx *Transaction) {
	n.txPool.Add(tx)
	inv := InvPayload{Items: []InvItem{{Type: "tx", Hash: tx.ID}}}
	n.broadcast(inv)
}

// broadcast helper
func (n *Node) broadcast(payload interface{}) {
	msg := MustWrap(payload)
	n.peerMu.RLock()
	defer n.peerMu.RUnlock()
	for _, p := range n.peers {
		_ = p.Send(msg) // ignore error
	}
}

// acceptLoop runs in goroutine
func (n *Node) acceptLoop() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			return
		}
		go n.handleConn(conn, false)
	}
}

// handleConn runs in its own goroutine
func (n *Node) handleConn(conn net.Conn, outbound bool) {
	remoteAddr := conn.RemoteAddr().String()
	p := NewPeer(conn, n)
	n.peerMu.Lock()
	if len(n.peers) >= maxPeers {
		n.peerMu.Unlock()
		conn.Close()
		return
	}
	n.peers[remoteAddr] = p
	n.peerMu.Unlock()

	defer func() {
		n.peerMu.Lock()
		delete(n.peers, remoteAddr)
		n.peerMu.Unlock()
		conn.Close()
	}()

	if err := p.Run(); err != nil {
		fmt.Println("peer error:", err)
	}
}

// MustWrap creates a Message from any payload (panic on error)
func MustWrap(v interface{}) Message {
	var typ string
	switch v.(type) {
	case VersionPayload:
		typ = "version"
	case VerackPayload:
		typ = "verack"
	case GetBlocksPayload:
		typ = "getblocks"
	case InvPayload:
		typ = "inv"
	case GetDataPayload:
		typ = "getdata"
	case BlockPayload:
		typ = "block"
	case TxPayload:
		typ = "tx"
	case GetMempoolPayload:
		typ = "getmempool"
	case MempoolPayload:
		typ = "mempool"
	default:
		panic("unknown message type")
	}
	raw, _ := json.Marshal(v)
	return Message{Type: typ, Payload: raw}
}
