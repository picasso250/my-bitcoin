package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
const difficulty = 4 // 将挖矿难度定义移到这里，作为链的属性

// Blockchain 结构体代表了整条链
type Blockchain struct {
	tip []byte   // 存储最后一个区块的哈希
	db  *bbolt.DB // 数据库连接 (保持私有)
}

// DB 返回对数据库的引用。这是一个公开的 Getter 方法。
func (bc *Blockchain) DB() *bbolt.DB {
	return bc.db
}

// MineBlock 查找、打包交易并创建一个新区块 (工作量证明)
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	// 1. 获取最后一个区块的哈希作为新区块的前向哈希
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// 2. 创建新区块
	newBlock := NewBlock(transactions, lastHash)

	// 3. 工作量证明 (PoW)
	target := strings.Repeat("0", difficulty)
	fmt.Printf("Mining the block...\n")
	for {
		newBlock.SetHash()
		hashHex := hex.EncodeToString(newBlock.Hash)

		if strings.HasPrefix(hashHex, target) {
			fmt.Printf("Mined! hash: %s, nonce: %d\n", hashHex, newBlock.Nonce)
			break
		}
		newBlock.Nonce++
	}

	// 4. 将挖出的新区块存入数据库
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

// NewBlockchain 创建一个新的区块链数据库，如果不存在，则创建创世区块
func NewBlockchain(minerAddress string) *Blockchain {
	var tip []byte
	db, err := bbolt.Open(dbFile, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
			cbtx := NewCoinbaseTX(minerAddress, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			fmt.Println("Existing blockchain found.")
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