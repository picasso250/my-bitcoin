package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"strconv"
	"time"
)

// Block 定义了区块的数据结构
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction // 区块包含的交易列表
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// HashTransactions 计算并返回交易列表的哈希值（简化的默克尔树根）
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

// SetHash 计算并设置区块的哈希
// 这是工作量证明的核心计算部分
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	nonce := []byte(strconv.Itoa(b.Nonce))
	headers := bytes.Join(
		[][]byte{
			b.PrevBlockHash,
			b.HashTransactions(),
			timestamp,
			nonce,
		},
		[]byte{},
	)
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// NewBlock 创建一个新的区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	// 注意：哈希现在由挖矿过程决定，创建时不再计算
	return block
}

// NewGenesisBlock 创建创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	// 创世区块直接创建，不经过挖矿
	block := NewBlock([]*Transaction{coinbase}, []byte{})
	block.SetHash()
	return block
}

// Serialize 使用 gob 编码将区块序列化为字节切片
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock 使用 gob 解码将字节切片反序列化为一个区块指针
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}