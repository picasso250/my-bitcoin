package blockchain

import (
	"encoding/hex"
	"encoding/json"

	"my-blockchain/gocoin/p2p" // 引入刚搬走的类型
)

// 以下方法仍挂在区块链自己的类型上，但内部使用 p2p 包里的视图结构
func (b *Block) MarshalJSON() ([]byte, error) {
	jb := p2p.JSONBlock{
		Hash:          hex.EncodeToString(b.Hash),
		PrevBlockHash: hex.EncodeToString(b.PrevBlockHash),
		Timestamp:     b.Timestamp,
		Nonce:         b.Nonce,
		Transactions:  make([]p2p.JSONTx, len(b.Transactions)),
	}
	for i, tx := range b.Transactions {
		jb.Transactions[i] = txToJSON(tx)
	}
	return json.Marshal(jb)
}

func (b *Block) UnmarshalJSON(data []byte) error {
	var jb p2p.JSONBlock
	if err := json.Unmarshal(data, &jb); err != nil {
		return err
	}
	b.Hash, _ = hex.DecodeString(jb.Hash)
	b.PrevBlockHash, _ = hex.DecodeString(jb.PrevBlockHash)
	b.Timestamp = jb.Timestamp
	b.Nonce = jb.Nonce
	b.Transactions = make([]*Transaction, len(jb.Transactions))
	for i, jtx := range jb.Transactions {
		b.Transactions[i] = jsonToTx(jtx)
	}
	return nil
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(txToJSON(tx))
}

func (tx *Transaction) UnmarshalJSON(data []byte) error {
	var jtx p2p.JSONTx
	if err := json.Unmarshal(data, &jtx); err != nil {
		return err
	}
	*tx = *jsonToTx(jtx)
	return nil
}

// ---------- 辅助函数 ----------
func txToJSON(tx *Transaction) p2p.JSONTx {
	jtx := p2p.JSONTx{
		ID:   hex.EncodeToString(tx.ID),
		Vin:  make([]p2p.JSONTxIn, len(tx.Vin)),
		Vout: make([]p2p.JSONTxOut, len(tx.Vout)),
	}
	for i, in := range tx.Vin {
		jtx.Vin[i] = p2p.JSONTxIn{
			Txid:      hex.EncodeToString(in.Txid),
			Vout:      in.Vout,
			Signature: hex.EncodeToString(in.Signature),
			PubKey:    hex.EncodeToString(in.PubKey),
		}
	}
	for i, out := range tx.Vout {
		jtx.Vout[i] = p2p.JSONTxOut{
			Value:      out.Value,
			PubKeyHash: hex.EncodeToString(out.PubKeyHash),
		}
	}
	return jtx
}

func jsonToTx(jtx p2p.JSONTx) *Transaction {
	tx := &Transaction{
		ID:   make([]byte, 32),
		Vin:  make([]TxInput, len(jtx.Vin)),
		Vout: make([]TxOutput, len(jtx.Vout)),
	}
	copy(tx.ID, jtx.ID)
	for i, jin := range jtx.Vin {
		tx.Vin[i] = TxInput{
			Txid:      make([]byte, 32),
			Vout:      jin.Vout,
			Signature: make([]byte, len(jin.Signature)/2),
			PubKey:    make([]byte, len(jin.PubKey)/2),
		}
		copy(tx.Vin[i].Txid, jin.Txid)
		copy(tx.Vin[i].Signature, jin.Signature)
		copy(tx.Vin[i].PubKey, jin.PubKey)
	}
	for i, jout := range jtx.Vout {
		tx.Vout[i] = TxOutput{
			Value:      jout.Value,
			PubKeyHash: make([]byte, len(jout.PubKeyHash)/2),
		}
		copy(tx.Vout[i].PubKeyHash, jout.PubKeyHash)
	}
	return tx
}
