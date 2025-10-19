package blockchain

import (
	"log"
	"time"

	"go.etcd.io/bbolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

// Blockchain 结构体代表了整条链
type Blockchain struct {
	tip []byte   // 存储最后一个区块的哈希
	db  *bbolt.DB // 数据库连接 (保持私有)
}

// DB 返回对数据库的引用。这是一个公开的 Getter 方法。
func (bc *Blockchain) DB() *bbolt.DB {
	return bc.db
}

// NewBlockchain 创建一个新的区块链数据库，如果不存在，则创建创世区块
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bbolt.Open(dbFile, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			// 如果 bucket 不存在，说明是第一次创建
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			// 将创世区块存入数据库
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			// 将创世区块的哈希存为 "l" (last)
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			// 如果 bucket 已存在，读取 "l" 获取最后一个区块的哈希
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

// AddBlock 向区块链中添加一个新的区块
// (我们先定义好，下一步再实现命令行来调用它)
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	// 1. 获取最后一个区块的哈希
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// 2. 创建新区块
	newBlock := NewBlock(data, lastHash)

	// 3. 将新区块存入数据库
	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		
		bc.tip = newBlock.Hash // 更新内存中的 tip
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}