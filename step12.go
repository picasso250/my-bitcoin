// step12.go
//
// 教学目标：演示如何使用我们功能更丰富的 `lib.go` 工具库。
//
// 我们将演示：
// 1. 调用库函数生成密钥对和地址 (与之前相同)。
// 2. 使用库中定义的数据结构 `Transaction` 和 `Block`，并展示它们。

package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

// RunStep12 是第12步的入口函数，由 main.go 调用。
func RunStep12() {
	fmt.Println("--- Step 12: 演示Go语言版区块链核心库与数据结构 ---")

	// --- 1. 演示加密功能 (与之前版本类似) ---
	privateKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}
	publicKey := &privateKey.PublicKey
	address := PublicKeyToAddress(publicKey)

	fmt.Println("\n[1] 成功生成密钥和地址:")
	fmt.Printf("  - 地址 (Hex): %s\n", address)

	// --- 2. 演示核心数据结构 ---

	// a) 创建一个空的交易 (此时没有ID, 输入和输出)
	emptyTx := &Transaction{}
	fmt.Println("\n[2] 创建一个空的交易结构:")
	fmt.Printf("  - 交易ID (初始): %s\n", hex.EncodeToString(emptyTx.ID))
	fmt.Printf("  - 输入数量: %d\n", len(emptyTx.Vin))
	fmt.Printf("  - 输出数量: %d\n", len(emptyTx.Vout))

	// b) 创建一个创世区块 (没有交易，也没有前一个区块哈希)
	genesisBlock := NewGenesisBlock()
	fmt.Println("\n[3] 创建一个创世区块:")
	fmt.Printf("  - 区块时间戳: %d\n", genesisBlock.Timestamp)
	fmt.Printf("  - 前一区块哈希: %s\n", hex.EncodeToString(genesisBlock.PrevBlockHash))
	fmt.Printf("  - 当前区块哈希: %s\n", hex.EncodeToString(genesisBlock.Hash))
	fmt.Printf("  - 包含交易数: %d\n", len(genesisBlock.Transactions))


	fmt.Println("\n[4] 核心概念:")
	fmt.Println("-> 我们的 `lib.go` 不再仅仅是函数的集合，它现在定义了区块链的骨架。")
	fmt.Println("-> `Transaction` 和 `Block` 结构体是我们构建完整区块链账本的基石。")
}