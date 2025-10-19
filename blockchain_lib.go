// blockchain_lib.go
//
// 这是我们用Go语言实现的“区块链核心工具箱”，它将整合所有与
// 加密、哈希、数据结构相关的基础功能。
// 它的目标是取代之前 step1 到 step9 的所有Python脚本的功能。

package main

import (
	"crypto/ecdsa" // 导入标准库的ecdsa，因为ethereum的类型是基于它的
	"crypto/sha256"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto" // 导入 go-ethereum 的主 crypto 包
	"golang.org/x/crypto/ripemd160"
)

// Transaction 定义了一个简化的交易结构 (未来会扩展成UTXO模型)
type Transaction struct {
	From   string
	To     string
	Amount int
	// Signature []byte // 签名等字段将在后续步骤加入
}

// Block 定义了区块的数据结构
type Block struct {
	Timestamp     int64
	Transactions  []Transaction
	PreviousHash  string
	Nonce         int
	// Hash          string // 哈希值将在计算后动态获得
}

// NewKeyPair 生成一个符合比特币 secp256k1 曲线标准的密钥对。
//
// 修正：我们直接使用 go-ethereum/crypto 包提供的 GenerateKey 函数。
// 这个函数是专门为 secp256k1 设计的，可以避免混合使用标准库可能引发的兼容性问题。
func NewKeyPair() (*ecdsa.PrivateKey, error) {
	// crypto.GenerateKey() 内部已经封装了所有细节，包括使用安全的随机源。
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// PublicKeyToAddress 将一个 ECDSA 公钥转换为教学版的十六进制地址。
// 这个流程严格复刻了 step3_public_key_to_address.py 的逻辑。
func PublicKeyToAddress(pubKey *ecdsa.PublicKey) string {
	// --- 0. 获取压缩公钥字节 ---
	// go-ethereum 的 crypto.CompressPubkey 实现了这个功能
	pubKeyBytes := crypto.CompressPubkey(pubKey)

	// --- 1. 双哈希: SHA256 -> RIPEMD160 ---
	// a) SHA256 哈希
	sha256Hasher := sha256.New()
	sha256Hasher.Write(pubKeyBytes)
	sha256Hash := sha256Hasher.Sum(nil)

	// b) RIPEMD160 哈希
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash)
	pubKeyHash := ripemd160Hasher.Sum(nil)

	// --- 2. 添加版本字节 (0x00) ---
	versionedHash := append([]byte{0x00}, pubKeyHash...)

	// --- 3. 计算并拼接校验和 ---
	// a) 第一次 SHA256
	checksumHash1 := sha256.Sum256(versionedHash)
	// b) 第二次 SHA256
	checksumHash2 := sha256.Sum256(checksumHash1[:])
	// c) 取前4个字节作为校验和
	checksum := checksumHash2[:4]
	// d) 拼接
	finalAddressBytes := append(versionedHash, checksum...)

	// --- 4. 编码为十六进制字符串 ---
	address := hex.EncodeToString(finalAddressBytes)

	return address
}