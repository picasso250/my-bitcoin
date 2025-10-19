// step12.go
//
// 教学目标：演示如何使用我们新创建的 `blockchain_lib.go` 工具库。
//
// 这个文件本身不包含复杂的逻辑，它的主要作用是：
// 1. 调用库函数来生成密钥对。
// 2. 调用库函数来派生地址。
// 3. 打印结果，验证工具库的功能与我们之前Python脚本的输出一致。

package main

import (
	"encoding/hex" // <-- 修正：添加缺失的包导入
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// RunStep12 是第12步的入口函数，由 main.go 调用。
func RunStep12() {
	fmt.Println("--- Step 12: 演示Go语言版区块链核心库 ---")

	// 1. 使用库函数生成一个新的密钥对
	privateKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}

	// 从私钥中获取公钥
	publicKey := &privateKey.PublicKey

	// 2. 将密钥转换为十六进制以便打印
	// crypto.FromECDSA 将私钥对象转换为字节切片
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// crypto.FromECDSAPub 将公钥对象转换为非压缩字节切片
	// 我们在这里手动获取压缩公 गोयल用于打印，与地址生成过程保持一致
	publicKeyBytesCompressed := crypto.CompressPubkey(publicKey)
	publicKeyHexCompressed := hex.EncodeToString(publicKeyBytesCompressed)

	fmt.Println("\n[1] 生成的密钥对 (secp256k1):")
	fmt.Printf("  - 私钥 (Hex): %s\n", privateKeyHex)
	fmt.Printf("  - 公钥 (Compressed Hex): %s\n", publicKeyHexCompressed)

	// 3. 使用库函数从公钥派生地址
	address := PublicKeyToAddress(publicKey)

	fmt.Println("\n[2] 从公钥派生的教学版地址:")
	fmt.Printf("  - 地址 (Hex): %s\n", address)
	fmt.Println("  - 结构: [版本: 1字节 | 公钥哈希: 20字节 | 校验和: 4字节]")

	fmt.Println("\n[3] 核心概念:")
	fmt.Println("-> 成功将Python的核心加密逻辑迁移到了可复用的Go函数中。")
	fmt.Println("-> `blockchain_lib.go` 成为了我们未来构建交易、区块和完整节点的基础。")
}
