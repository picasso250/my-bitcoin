// step14.go
//
// 教学目标：演示矿工如何将一笔普通交易和一笔Coinbase奖励交易打包成一个有效的区块。
// (增强版) 这个版本将演示一个更真实的场景：
// 1. 多输入: 交易将花费多个UTXO。
// 2. 矿工费: 交易的总输入会大于总输出，差额部分作为给矿工的小费。
// 3. 矿工奖励升级: 矿工的Coinbase奖励 = 固定区块奖励 + 所有交易的矿工费总和。

package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// RunStep14 是第14步的入口函数，由 main.go 调用。
func RunStep14() {
	fmt.Println("--- Step 14 (增强版): 演示包含多输入和矿工费的交易打包 ---")

	// --- 1. 场景设置: 创建矿工、Alice和Bob ---
	minerPrivKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("无法创建矿工的密钥: %v", err)
	}
	minerAddress := PublicKeyToAddress(&minerPrivKey.PublicKey)

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
	bobAddress := PublicKeyToAddress(&bobPrivKey.PublicKey)

	fmt.Println("[1] 参与方身份:")
	fmt.Printf("  - Miner's Address: %s\n", minerAddress)
	fmt.Printf("  - Alice's Address: %s\n", aliceAddress)
	fmt.Printf("  - Bob's Address:   %s\n", bobAddress)

	// --- 2. Alice 创建一笔多输入、带矿工费的交易 ---

	// a) 模拟Alice拥有的两个UTXO
	prevTxID1, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000001a")
	utxo1 := TxInput{
		Txid:      prevTxID1,
		Vout:      0, // 第一个UTXO，价值100
		Signature: nil,
		PubKey:    crypto.CompressPubkey(alicePubKey),
	}
	prevTxID2, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000002b")
	utxo2 := TxInput{
		Txid:      prevTxID2,
		Vout:      1, // 第二个UTXO，价值50
		Signature: nil,
		PubKey:    crypto.CompressPubkey(alicePubKey),
	}
	totalInputValue := 100 + 50
	fmt.Printf("\n[2] Alice 准备花钱，她拥有两个UTXO (100 + 50 = %d)\n", totalInputValue)

	// b) 构建交易，目标：支付120给Bob，留下10聪作为矿工费
	paymentToBob := 120
	changeToAlice := 20 // 150 - 120 = 30, 但只找零20, 剩下10是矿工费
	minerFee := totalInputValue - paymentToBob - changeToAlice

	vout := []TxOutput{
		{Value: paymentToBob, ScriptPubKey: bobAddress},
		{Value: changeToAlice, ScriptPubKey: aliceAddress},
	}
	aliceTx := &Transaction{
		Vin:  []TxInput{utxo1, utxo2}, // 两个输入
		Vout: vout,
	}

	// c) 签名交易 (为简化，我们用同一个密钥对两个输入签名)
	txHash := aliceTx.Hash()
	signature, err := Sign(alicePrivKey, txHash)
	if err != nil {
		log.Fatalf("Alice签名失败: %v", err)
	}
	// 将签名应用到所有输入上
	aliceTx.Vin[0].Signature = signature
	aliceTx.Vin[1].Signature = signature

	aliceTx.ID = aliceTx.Hash()
	fmt.Printf("[3] Alice 创建了交易 (ID: %s...)\n", hex.EncodeToString(aliceTx.ID)[:10])
	fmt.Printf("    - 输入总额: %d\n", totalInputValue)
	fmt.Printf("    - 支付给Bob: %d\n", paymentToBob)
	fmt.Printf("    - 找零给自己: %d\n", changeToAlice)
	fmt.Printf("    - 隐含矿工费: %d\n", minerFee)

	// --- 3. 矿工创建Coinbase交易 ---
	// 关键：矿工的奖励 = 固定奖励 + 他打包的所有交易的矿工费
	coinbaseTx := NewCoinbaseTX(minerAddress, "Mined by GoLang Gemini")
	coinbaseTx.Vout[0].Value += minerFee // 在50的基础上加上矿工费
	coinbaseTx.ID = coinbaseTx.Hash()    // 更新ID

	fmt.Printf("\n[4] 矿工创建了Coinbase交易 (ID: %s...)\n", hex.EncodeToString(coinbaseTx.ID)[:10])
	fmt.Printf("    - 固定奖励: 50\n")
	fmt.Printf("    - 收到矿工费: %d\n", minerFee)
	fmt.Printf("    - 总奖励: %d\n", coinbaseTx.Vout[0].Value)

	// --- 4. 矿工打包区块 ---
	transactionsForBlock := []*Transaction{coinbaseTx, aliceTx}
	prevBlockHash, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000020")
	newBlock := NewBlock(transactionsForBlock, prevBlockHash)

	fmt.Println("\n[5] 矿工已成功将两笔交易打包进新区块:")
	fmt.Printf("  - 区块哈希: %s\n", hex.EncodeToString(newBlock.Hash))
	fmt.Printf("  - 交易数量: %d\n", len(newBlock.Transactions))
	fmt.Printf("  - 第一笔交易 (Coinbase) 的奖励金额: %d\n", newBlock.Transactions[0].Vout[0].Value)
}