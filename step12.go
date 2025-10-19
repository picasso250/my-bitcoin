// step12.go
//
// 教学目标：演示矿工如何将一笔普通交易和一笔Coinbase奖励交易打包成一个有效的区块。
// (最终版) 这个版本将完整演示：
// 1. 多输入: 交易将花费多个UTXO。
// 2. 矿工费: 交易的总输入会大于总输出，差额部分作为给矿工的小费。
// 3. 矿工奖励升级: 矿工的Coinbase奖励 = 固定区块奖励 + 所有交易的矿工费总和。
// 4. 挖矿: 将打包好的区块通过工作量证明（PoW）挖出来，得到一个满足难度要求的哈希。

package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// RunStep12 是第12步的入口函数，由 main.go 调用。
func RunStep12() {
	fmt.Println("--- Step 12 (最终版): 演示从打包到挖矿的全过程 ---")

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

	// c) 签名交易
	txHash := aliceTx.Hash()
	signature, err := Sign(alicePrivKey, txHash)
	if err != nil {
		log.Fatalf("Alice签名失败: %v", err)
	}
	aliceTx.Vin[0].Signature = signature
	aliceTx.Vin[1].Signature = signature

	aliceTx.ID = aliceTx.Hash()
	fmt.Printf("[3] Alice 创建了交易 (ID: %s...)\n", hex.EncodeToString(aliceTx.ID)[:10])
	fmt.Printf("    - 隐含矿工费: %d\n", minerFee)

	// --- 3. 矿工创建Coinbase交易 ---
	coinbaseTx := NewCoinbaseTX(minerAddress, "Mined by GoLang Gemini")
	coinbaseTx.Vout[0].Value += minerFee // 在50的基础上加上矿工费
	coinbaseTx.ID = coinbaseTx.Hash()

	fmt.Printf("\n[4] 矿工创建了Coinbase交易 (ID: %s...)\n", hex.EncodeToString(coinbaseTx.ID)[:10])
	fmt.Printf("    - 总奖励 (50 + %d): %d\n", minerFee, coinbaseTx.Vout[0].Value)

	// --- 4. 矿工打包区块 ---
	transactionsForBlock := []*Transaction{coinbaseTx, aliceTx}
	prevBlockHash, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000020")
	newBlock := NewBlock(transactionsForBlock, prevBlockHash)

	// --- 5. 矿工进行挖矿 ---
	newBlock.MineBlock() // 这是新增的核心步骤！

	// --- 6. 最终结果展示 ---
	fmt.Println("\n[6] 区块诞生全过程完成:")
	fmt.Printf("  - 最终区块哈希: %s\n", hex.EncodeToString(newBlock.Hash))
	fmt.Printf("  - Nonce: %d\n", newBlock.Nonce)
	fmt.Printf("  - 交易数量: %d\n", len(newBlock.Transactions))
	fmt.Printf("  - 第一笔交易 (Coinbase) 的奖励金额: %d\n", newBlock.Transactions[0].Vout[0].Value)
}