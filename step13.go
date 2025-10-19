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

	// [FIXED] 为Bob也创建一个合法的密钥对和地址，不再使用硬编码的假地址
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
	// 假设Alice拥有一个来自之前某笔交易的输出
	prevTxID, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	utxoToSpend := TxInput{
		Txid:      prevTxID,
		Vout:      0, // 这个UTXO是该交易的第一个输出
		Signature: nil, // 签名此时为空，稍后填充
		PubKey:    crypto.CompressPubkey(alicePubKey), // 公钥需要填充，用于验签
	}

	// --- 3. 构建交易 ---
	// 目标：Alice想用一个虚拟的UTXO，给Bob转30聪，并找零70聪给自己。

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

	// c) 计算待签名的交易哈希 (核心点1)
	// 这个哈希是基于不含签名的交易内容生成的。这是我们要签名的“合同”。
	txHash := tx.Hash()
	fmt.Printf("\n[2] 构建的交易 (待签名):\n")
	fmt.Printf("  - 待签名的交易哈希: %s\n", hex.EncodeToString(txHash))


	// --- 4. 签名交易 ---
	signature, err := Sign(alicePrivKey, txHash)
	if err != nil {
		log.Fatalf("签名失败: %v", err)
	}
	// 将签名填充到交易中
	tx.Vin[0].Signature = signature

	// --- 5. 设置最终交易ID ---
	// 签名完成后，整个交易的内容才算最终确定，此时我们计算最终的ID
	// 注意：这个ID与用于签名的txHash是不同的，因为它现在包含了签名。
	tx.ID = tx.Hash()

	fmt.Printf("\n[3] 签名完成后的交易:\n")
	fmt.Printf("  - 最终交易ID: %s\n", hex.EncodeToString(tx.ID))
	fmt.Printf("  - 输入签名: %s...\n", hex.EncodeToString(tx.Vin[0].Signature)[:20])


	// --- 6. 验证交易 (网络节点的操作) ---
	// 节点需要：交易哈希, 签名, 公钥
	
	// [FIXED] 关键修复点！
	// 用于验证的数据，必须是当初用于签名的数据。
	// 当初签名的是 txHash (不含签名的交易哈希)，所以这里必须用它来验证。
	// 错误的做法是: dataToVerify := tx.Hash()，因为那将是包含了签名的交易哈希。
	dataToVerify := txHash
	
	isValid := Verify(alicePubKey, dataToVerify, tx.Vin[0].Signature)

	fmt.Println("\n[4] 网络节点验证结果:")
	if isValid {
		fmt.Println("  - 验证成功! 👍 签名有效。")
	} else {
		fmt.Println("  - 验证失败! 💀 签名无效。")
	}
}