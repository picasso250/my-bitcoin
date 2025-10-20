package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain 结构体代表了整条链
type Blockchain struct {
	tip []byte   // 存储最后一个区块的哈希
	db  *bbolt.DB // 数据库连接
}

// DB 返回对数据库的引用。这是一个公开的 Getter 方法。
func (bc *Blockchain) DB() *bbolt.DB {
	return bc.db
}

// MineBlock 打包交易并创建一个新区块 (通过工作量证明)
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

	// 2. 创建新区块实例
	newBlock := NewBlock(transactions, lastHash)

	// 3. 创建 PoW 实例并执行挖矿
	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Run()

	// 4. 将计算出的 Nonce 和 Hash 设置回新区块
	newBlock.Nonce = nonce
	newBlock.Hash = hash

	// 5. 将挖出的新区块存入数据库
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

	// 6. 更新 UTXO 集
	utxoSet := UTXOSet{bc}
	utxoSet.Update(newBlock)
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

	// 在启动时重建 UTXO 索引
	fmt.Println("Reindexing UTXO set...")
	utxoSet := UTXOSet{&bc}
	utxoSet.Reindex()
	fmt.Println("Reindexing finished.")

	return &bc
}

// Iterator 返回一个 BlockchainIterator 实例
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

// FindAllUTXO finds all unspent transaction outputs and returns a map
// where keys are transaction IDs and values are slices of output indices
func (bc *Blockchain) FindAllUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}