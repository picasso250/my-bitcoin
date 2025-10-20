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

// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.ID)
	}
	mTree := sha256.Sum256(bytes.Join(transactions, []byte{}))

	return mTree[:]
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	
	// Note: Hash and Nonce are set by the proof-of-work algorithm
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	block := NewBlock([]*Transaction{coinbase}, []byte{})
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
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