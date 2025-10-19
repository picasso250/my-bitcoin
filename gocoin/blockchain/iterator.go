package blockchain

import (
	"log"

	"go.etcd.io/bbolt"
)

// BlockchainIterator 用于遍历区块链中的区块
type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

// Next 返回链中的下一个区块（从后向前遍历）
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		if encodedBlock == nil {
			// This can happen if the key is not found, handle gracefully.
			log.Printf("Warning: Block with hash %x not found in the bucket.", i.currentHash)
			return nil
		}
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// Move the iterator to the previous block hash
	if block != nil {
		i.currentHash = block.PrevBlockHash
	}

	return block
}