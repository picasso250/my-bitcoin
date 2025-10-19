// step13.go
//
// 教学目标：完整复现 step7 的核心逻辑 —— 创建并签名一笔单输入、双输出的UTXO交易。
//
// 这标志着我们已经将Python原型中最核心的交易概念成功迁移到了Go语言的工具库中。

package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// RunStep13 是第13步的入口函数，由 main.go 调用。
func RunStep13() {
	fmt.Println("--- Step 13: 演示Go语言版UTXO交易创建与签名 (架构优化版) ---")

	// --- 1. 场景设置: 创建参与方 ---
	// Alice (发送者) 和 Bob (接收者)
	alicePrivKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("无法创建Alice的密钥: %v", err)
	}
	alicePubKey := &alicePrivKey.PublicKey
	aliceAddress := PublicKeyToAddress(alicePubKey)

	bobPrivKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("无法创建Bob的密钥: %v", err)
	}
	bobPubKey := &bobPrivKey.PublicKey
	bobAddress := PublicKeyToAddress(bobPubKey)

	fmt.Println("[1] 参与方身份:")
	fmt.Printf("  - Alice's Address: %s\n", aliceAddress)
	fmt.Printf("  - Bob's Address:   %s\n", bobAddress)

	// --- 2. 模拟一个Alice拥有的UTXO ---
	prevTxID, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	utxoToSpend := TxInput{
		Txid:      prevTxID,
		Vout:      0,
		Signature: nil,
		PubKey:    crypto.CompressPubkey(alicePubKey),
	}

	// --- 3. 构建交易 ---
	vout := []TxOutput{
		{Value: 30, ScriptPubKey: bobAddress},
		{Value: 70, ScriptPubKey: aliceAddress},
	}
	tx := Transaction{
		Vin:  []TxInput{utxoToSpend},
		Vout: vout,
	}

	// 计算待签名的交易哈希 (基于不含签名的交易内容)
	txHash := tx.Hash()
	fmt.Printf("\n[2] 构建的交易 (待签名):\n")
	fmt.Printf("  - 待签名的交易哈希: %s\n", hex.EncodeToString(txHash))


	// --- 4. 签名交易 ---
	signature, err := Sign(alicePrivKey, txHash)
	if err != nil {
		log.Fatalf("签名失败: %v", err)
	}
	tx.Vin[0].Signature = signature

	// --- 5. 设置最终交易ID ---
	tx.ID = tx.Hash()
	fmt.Printf("\n[3] 签名完成后的交易:\n")
	fmt.Printf("  - 最终交易ID: %s\n", hex.EncodeToString(tx.ID))
	fmt.Printf("  - 输入签名: %s...\n", hex.EncodeToString(tx.Vin[0].Signature)[:20])


	// --- 6. 验证交易 (网络节点的操作) ---
	// [NEW] 架构优化：不再需要外部逻辑来手动验证。
	// 交易对象现在自己“知道”如何验证自己。
	// 节点只需调用这一个方法即可。
	fmt.Println("\n[4] 网络节点调用 tx.Verify() 进行验证:")
	if tx.Verify() {
		fmt.Println("  - 验证成功! 👍 交易有效。")
	} else {
		fmt.Println("  - 验证失败! 💀 交易无效。")
	}
}