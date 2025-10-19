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
	fmt.Println("--- Step 13: 演示Go语言版UTXO交易创建与签名 ---")

	// --- 1. 场景设置: 创建参与方 ---
	// Alice (发送者) 和 Bob (接收者)
	alicePrivKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("无法创建Alice的密钥: %v", err)
	}
	alicePubKey := &alicePrivKey.PublicKey
	aliceAddress := PublicKeyToAddress(alicePubKey)

	_, err = NewKeyPair() // Just to create a different key for Bob
	if err != nil {
		log.Fatalf("无法创建Bob的密钥: %v", err)
	}
	bobAddress := "1GaR4Mr3o8d3n2AkjJk53B5g3h3s4g5j6k" // 简化演示，Bob只有一个地址字符串

	fmt.Println("[1] 参与方身份:")
	fmt.Printf("  - Alice's Address: %s\n", aliceAddress)
	fmt.Printf("  - Bob's Address:   %s\n", bobAddress)

	// --- 2. 模拟一个Alice拥有的UTXO ---
	// 假设Alice拥有一个来自之前某笔交易的输出
	prevTxID, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	utxoToSpend := TxInput{
		Txid:      prevTxID,
		Vout:      0, // 这个UTXO是该交易的第一个输出
		Signature: nil, // 签名此时为空，稍后填充
		PubKey:    crypto.CompressPubkey(alicePubKey), // 公钥需要填充，用于验签
	}

	// --- 3. 构建交易 ---
	// 目标：Alice想用她那100聪的UTXO，给Bob转30聪，并找零70聪给自己。

	// a) 构建输出 (vout)
	vout := []TxOutput{
		{Value: 30, ScriptPubKey: bobAddress},    // 给Bob的
		{Value: 70, ScriptPubKey: aliceAddress}, // 找零给Alice
	}

	// b) 组装交易
	tx := Transaction{
		Vin:  []TxInput{utxoToSpend},
		Vout: vout,
	}

	// c) 计算待签名的交易哈希
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
	// 签名完成后，整个交易的内容才算最终确定，此时我们计算最终的ID
	tx.ID = tx.Hash()

	fmt.Printf("\n[3] 签名完成后的交易:\n")
	fmt.Printf("  - 最终交易ID: %s\n", hex.EncodeToString(tx.ID))
	fmt.Printf("  - 输入签名: %s...\n", hex.EncodeToString(tx.Vin[0].Signature)[:20])


	// --- 6. 验证交易 (网络节点的操作) ---
	// 节点需要：交易哈希, 签名, 公钥
	// 注意：在真实场景中，节点需要从输入中提取公钥字节，然后反序列化为公钥对象。
	// 这里为简化演示，我们直接使用之前生成的公钥对象。
	dataToVerify := tx.Hash() // 节点会独立计算这个哈希
	isValid := Verify(alicePubKey, dataToVerify, tx.Vin[0].Signature)

	fmt.Println("\n[4] 网络节点验证结果:")
	if isValid {
		fmt.Println("  - 验证成功! 👍 签名有效。")
	} else {
		fmt.Println("  - 验证失败! 💀 签名无效。")
	}
}