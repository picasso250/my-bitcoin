// lib.go
//
// 这是我们用Go语言实现的“区块链核心工具箱”，它将整合所有与
// 加密、哈希、数据结构相关的基础功能。
// 它的目标是取代之前 step1 到 step9 的所有Python脚本的功能。

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/ripemd160"
)

// --- 常量 ---
const difficulty = 4 // 定义工作量证明的难度 (需要4个前导零)

// --- 核心数据结构 (源自 Python steps 4-9) ---

// TxInput 代表一个交易输入
type TxInput struct {
	Txid      []byte // 引用的是哪个交易的ID (哈希)
	Vout      int    // 该交易的第几个输出
	Signature []byte // 对当前交易的签名
	PubKey    []byte // 发起者的公钥 (非哈希)
}

// TxOutput 代表一个交易输出
type TxOutput struct {
	Value        int    // 金额 (单位: 聪)
	ScriptPubKey string // 锁定脚本，在我们的简化版里就是接收者地址
}

// Transaction 定义了一个符合UTXO模型的交易结构
type Transaction struct {
	ID   []byte     // 交易的ID (哈希)
	Vin  []TxInput  // 输入列表
	Vout []TxOutput // 输出列表
}

// Block 定义了区块的数据结构
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction // 区块包含的交易列表
	PrevBlockHash []byte         // 前一个区块的哈希
	Hash          []byte         // 当前区块的哈希
	Nonce         int
}

// --- 核心功能函数 ---

// NewKeyPair 生成一个符合比特币 secp256k1 曲线标准的密钥对。
func NewKeyPair() (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// PublicKeyToAddress 将一个 ECDSA 公钥转换为教学版的十六进制地址。
// 这个流程严格复刻了 step3_public_key_to_address.py 的逻辑。
func PublicKeyToAddress(pubKey *ecdsa.PublicKey) string {
	pubKeyBytes := crypto.CompressPubkey(pubKey)
	sha256Hasher := sha256.New()
	sha256Hasher.Write(pubKeyBytes)
	sha256Hash := sha256Hasher.Sum(nil)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash)
	pubKeyHash := ripemd160Hasher.Sum(nil)

	versionedHash := append([]byte{0x00}, pubKeyHash...)
	checksumHash1 := sha256.Sum256(versionedHash)
	checksumHash2 := sha256.Sum256(checksumHash1[:])
	checksum := checksumHash2[:4]
	finalAddressBytes := append(versionedHash, checksum...)

	return hex.EncodeToString(finalAddressBytes)
}

// Sign 使用给定的私钥对数据哈希进行签名。
func Sign(privateKey *ecdsa.PrivateKey, dataHash []byte) ([]byte, error) {
	return crypto.Sign(dataHash, privateKey)
}

// Verify 验证给定的签名是否有效。
func Verify(publicKey *ecdsa.PublicKey, dataHash []byte, signature []byte) bool {
	pubKeyBytes := crypto.FromECDSAPub(publicKey)
	if len(signature) == 0 || len(pubKeyBytes) == 0 || len(dataHash) == 0 {
		return false
	}
	sigWithoutRecoveryID := signature[:len(signature)-1] // 去掉 V
	return crypto.VerifySignature(pubKeyBytes, dataHash, sigWithoutRecoveryID)
}

// HashTransactions 计算并返回交易列表的哈希值，用于构建区块。
// 这是一个简化的默克尔树根实现。
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

// SetHash 计算并设置区块的哈希。
// 这是工作量证明的核心部分。
func (b *Block) SetHash() {
	timestamp := []byte(time.Unix(b.Timestamp, 0).String())
	// 将 Nonce (整数) 转换为其十进制表示的字符串，再转换为字节数组
	nonce := []byte(strconv.Itoa(b.Nonce))
	headers := bytes.Join(
		[][]byte{
			b.PrevBlockHash,
			b.HashTransactions(),
			timestamp,
			nonce, // 使用转换后的 nonce
		},
		[]byte{},
	)
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// [NEW] MineBlock 实现了工作量证明算法。
// 它会不断调整 Nonce 并重新计算哈希，直到找到满足难度要求的哈希值。
func (b *Block) MineBlock() {
	target := strings.Repeat("0", difficulty)
	fmt.Printf("\n[5] 开始挖矿... 目标: 找到一个以 '%s' 开头的哈希。\n", target)

	for {
		b.SetHash()
		hashHex := hex.EncodeToString(b.Hash)

		// 提供一个简单的进度反馈
		if b.Nonce%200000 == 0 && b.Nonce > 0 {
			fmt.Printf("  - 尝试 Nonce: %d, 哈希: %s...\n", b.Nonce, hashHex[:12])
		}

		if strings.HasPrefix(hashHex, target) {
			fmt.Println("🎉 挖矿成功!")
			break // 找到了!
		}
		b.Nonce++
	}
}

// NewBlock 创建一个新的区块。
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	// 注意：我们不再在创建时计算哈希，因为哈希现在由挖矿过程决定。
	return block
}

// NewGenesisBlock 创建创世区块。
func NewGenesisBlock() *Block {
	// 创世区块通常包含一个特殊的Coinbase交易
	// 为简化，我们创建一个没有交易的创世区块
	block := NewBlock([]*Transaction{}, []byte{})
	block.SetHash() // 创世区块不需要挖矿，直接设置哈希
	return block
}

// Hash 计算交易的哈希值（作为交易ID）。
// 注意：这是一个简化的实现，真实的比特币实现会更复杂。
func (tx *Transaction) Hash() []byte {
	txCopy := *tx
	txCopy.ID = []byte{}

	encoded, err := json.Marshal(txCopy)
	if err != nil {
		// 在实际应用中应处理这个错误
		return nil
	}
	hash := sha256.Sum256(encoded)
	return hash[:]
}

// TrimmedCopy 创建一个交易的深拷贝，其中所有输入的签名都被剥离。
// 这对于创建用于签名或验证的交易哈希至关重要。
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, vin := range tx.Vin {
		// 只复制ID和Vout，剥离签名，但保留公钥用于后续验证
		inputs = append(inputs, TxInput{
			Txid:      vin.Txid,
			Vout:      vin.Vout,
			Signature: nil,
			PubKey:    vin.PubKey, // 保留公钥
		})
	}

	outputs = append(outputs, tx.Vout...)

	return Transaction{ID: nil, Vin: inputs, Vout: outputs}
}

// Verify 验证交易中所有输入的签名是否都有效。
// 这是对整个交易合法性的核心检查。
func (tx *Transaction) Verify() bool {
	if tx.IsCoinbase() {
		return true // Coinbase交易无需签名验证
	}
	if len(tx.Vin) == 0 {
		return true // 无输入的交易（非Coinbase），暂定为有效
	}

	// 制作一个不包含任何签名的交易副本，用于计算待验证的哈希
	txCopy := tx.TrimmedCopy()

	// 遍历原始交易中的每一个输入，用副本计算哈希并验证签名
	for _, vin := range tx.Vin {
		// 1. 计算用于验证的哈希
		// 这个哈希精确地模拟了当初签名时的交易状态（即，Signature字段为空）
		dataToVerify := txCopy.Hash()

		// 2. 从字节恢复公钥对象
		pubKey, err := crypto.DecompressPubkey(vin.PubKey)
		if err != nil {
			return false // 如果公钥格式不正确，验证失败
		}

		// 3. 调用底层的Verify函数进行密码学验证
		if !Verify(pubKey, dataToVerify, vin.Signature) {
			// 任何一个签名验证失败，则整个交易无效
			return false
		}
	}

	// 所有输入都验证成功
	return true
}

// IsCoinbase 检查当前交易是否为 Coinbase 交易
func (tx *Transaction) IsCoinbase() bool {
	// 满足三个条件：1.只有一个输入 2.该输入的Txid为空 3.该输入的Vout为-1
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// NewCoinbaseTX 创建一笔新的 Coinbase 交易
// to: 接收奖励的矿工地址
// data: 矿工想写入的任意数据，如果为空则生成默认信息
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// Coinbase交易的输入，遵循“约定”：
	// Txid 为空，Vout 为 -1，Signature可以存放任意数据
	txin := TxInput{
		Txid:      []byte{},
		Vout:      -1,
		Signature: []byte(data),
		PubKey:    nil,
	}

	// Coinbase交易的输出，将区块奖励发送给矿工
	// 注意：这里的Value应该是区块奖励+交易费，为简化，我们先只考虑区块奖励
	txout := TxOutput{
		Value:        50, // 硬编码的区块奖励
		ScriptPubKey: to,
	}

	tx := Transaction{
		Vin:  []TxInput{txin},
		Vout: []TxOutput{txout},
	}
	tx.ID = tx.Hash() // 设置交易ID

	return &tx
}