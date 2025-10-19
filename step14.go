// step14.go
//
// 教学目标：演示矿工如何将一笔普通交易和一笔Coinbase奖励交易打包成一个有效的区块。
// 这完美复现了 Python `step9` 的核心逻辑，证明了Go核心库已具备构建完整账本的能力。

package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// RunStep14 是第14步的入口函数，由 main.go 调用。
func RunStep14() {
	fmt.Println("--- Step 14: 演示矿工打包Coinbase与普通交易以构建区块 ---")

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
	fmt.Printf("  - Bob's Address:   %s\n\n", bobAddress)

	// --- 2. Alice 创建一笔普通的UTXO交易 ---
	// a) 模拟Alice拥有的一个UTXO
	prevTxID, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000010")
	utxoToSpend := TxInput{
		Txid:      prevTxID,
		Vout:      0,
		Signature: nil,
		PubKey:    crypto.CompressPubkey(alicePubKey),
	}

	// b) 构建交易，支付30给Bob，找零70给自己
	vout := []TxOutput{
		{Value: 30, ScriptPubKey: bobAddress},
		{Value: 70, ScriptPubKey: aliceAddress},
	}
	aliceTx := &Transaction{
		Vin:  []TxInput{utxoToSpend},
		Vout: vout,
	}

	// c) 签名交易
	txHash := aliceTx.Hash()
	signature, err := Sign(alicePrivKey, txHash)
	if err != nil {
		log.Fatalf("Alice签名失败: %v", err)
	}
	aliceTx.Vin[0].Signature = signature
	aliceTx.ID = aliceTx.Hash()
	fmt.Printf("[2] Alice 创建了一笔普通交易 (ID: %s...)\n", hex.EncodeToString(aliceTx.ID)[:10])

	// --- 3. 矿工创建Coinbase交易 ---
	// 矿工调用 `NewCoinbaseTX` 为自己创建一笔奖励交易
	coinbaseTx := NewCoinbaseTX(minerAddress, "Mined by GoLang Gemini")
	fmt.Printf("[3] 矿工创建了Coinbase奖励交易 (ID: %s...)\n\n", hex.EncodeToString(coinbaseTx.ID)[:10])

	// --- 4. 矿工打包区块 ---
	// 关键规则：Coinbase交易必须是列表中的第一个！
	transactionsForBlock := []*Transaction{coinbaseTx, aliceTx}

	// 使用一个虚拟的前区块哈希来创建新区块
	prevBlockHash, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000020")
	newBlock := NewBlock(transactionsForBlock, prevBlockHash)

	fmt.Println("[4] 矿工已成功将两笔交易打包进新区块:")
	fmt.Printf("  - 区块哈希: %s\n", hex.EncodeToString(newBlock.Hash))
	fmt.Printf("  - 交易数量: %d\n", len(newBlock.Transactions))
	fmt.Printf("  - 第一笔交易是否为Coinbase: %v\n", newBlock.Transactions[0].IsCoinbase())
	fmt.Printf("  - 第二笔交易是否为Coinbase: %v\n", newBlock.Transactions[1].IsCoinbase())
}