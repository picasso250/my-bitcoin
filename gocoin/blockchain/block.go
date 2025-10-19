package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

// Block 定义了区块的数据结构
// 注意：为了清晰，我们暂时移除了 Transaction 相关的字段，先聚焦于链本身
type Block struct {
	Timestamp     int64
	Data          []byte // 在这个简化版本中，我们将交易数据抽象为字节切片
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// SetHash 计算并设置区块的哈希
func (b *Block) SetHash() {
	timestamp := []byte(time.Unix(b.Timestamp, 0).String())
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// NewBlock 创建一个新的区块
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	block.SetHash() // 注意：这里我们暂时不进行挖矿，直接设置哈希
	return block
}

// NewGenesisBlock 创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
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