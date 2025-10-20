package gocoin

import (
	"sync"
)

// TxPool is a thread-safe in-memory mempool
type TxPool struct {
	mu  sync.RWMutex
	txs map[string]*Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{txs: make(map[string]*Transaction)}
}

func (tp *TxPool) Add(tx *Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.txs[string(tx.ID)] = tx
}

func (tp *TxPool) Get(id []byte) *Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.txs[string(id)]
}

func (tp *TxPool) Has(id []byte) bool {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	_, ok := tp.txs[string(id)]
	return ok
}

func (tp *TxPool) GetAllHashes() [][]byte {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	out := make([][]byte, 0, len(tp.txs))
	for id := range tp.txs {
		out = append(out, []byte(id))
	}
	return out
}
