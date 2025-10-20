package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	"my-blockchain/gocoin/wallet"
)

// TxInput 代表一个交易输入
type TxInput struct {
	Txid   []byte // 引用的是哪个交易的ID (哈希)
	Vout   int    // 该交易的第几个输出
	PubKey []byte // 发起者的公钥 (在解锁脚本中)
}

// TxOutput 代表一个交易输出
type TxOutput struct {
	Value      int    // 金额 (单位: 聪)
	PubKeyHash []byte // 锁定脚本，这里简化为公钥哈希
}

// Transaction 定义了一个符合UTXO模型的交易结构
type Transaction struct {
	ID   []byte     // 交易的ID (哈希)
	Vin  []TxInput  // 输入列表
	Vout []TxOutput // 输出列表
}

// SetID 计算并设置交易的ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// IsCoinbase 检查当前交易是否为 Coinbase 交易
func (tx *Transaction) IsCoinbase() bool {
	// 满足两个条件：1.只有一个输入 2.该输入的Txid为空
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0
}

// NewCoinbaseTX 创建一笔新的 Coinbase 交易
// toAddress: 接收奖励的矿工地址
// data: 矿工想写入的任意数据，如果为空则生成默认信息
func NewCoinbaseTX(toAddress, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", toAddress)
	}

	// Coinbase交易的输入，遵循“约定”：
	// Txid 为空，PubKey可以存放任意数据
	txin := TxInput{
		Txid:   []byte{},
		Vout:   -1, // Coinbase输入的Vout通常为-1
		PubKey: []byte(data),
	}

	// Coinbase交易的输出，将区块奖励发送给矿工
	// 先将地址解码为公Key哈希
	pubKeyHash, err := wallet.DecodeAddress(toAddress)
	if err != nil {
		log.Panic(err)
	}
	txout := TxOutput{
		Value:      50, // 硬编码的区块奖励
		PubKeyHash: pubKeyHash,
	}

	tx := Transaction{
		Vin:  []TxInput{txin},
		Vout: []TxOutput{txout},
	}
	tx.SetID() // 设置交易ID

	return &tx
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       PubKey:    %s", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

// UsesKey checks if the input uses a specific key
// For now, this is a simplified check.
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}