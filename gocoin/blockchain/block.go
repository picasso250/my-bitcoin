package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"strings"
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
	// 创世区块也需要通过挖矿来获得有效的哈希
	block := NewBlock([]*Transaction{coinbase}, []byte{})
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

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

// String returns a human-readable representation of a block
func (b Block) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("============ Block %x ============", b.Hash))
	lines = append(lines, fmt.Sprintf("Timestamp:     %s", time.Unix(b.Timestamp, 0).Format(time.RFC1123)))
	lines = append(lines, fmt.Sprintf("PrevBlockHash: %x", b.PrevBlockHash))
	lines = append(lines, fmt.Sprintf("Nonce:         %d", b.Nonce))
	lines = append(lines, "Transactions:")

	for _, tx := range b.Transactions {
		lines = append(lines, tx.String())
	}

	return strings.Join(lines, "\n")
}