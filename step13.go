// step13.go
//
// 教学目标：演示如何使用 Go 语言实现数字签名与验证。
// 这与 step2_sign_and_verify.py 的功能完全对应。

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
)

// RunStep13 是第13步的入口函数，由 main.go 调用。
func RunStep13() {
	fmt.Println("--- Step 13: 演示Go语言版签名与验签 ---")

	// --- 准备工作：生成密钥对 (同第12步) ---
	privateKey, err := NewKeyPair()
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// --- 场景模拟 ---
	// 假设我们要对一笔交易的核心内容进行签名。
	transactionData := "Alice sends 1 BTC to Bob"
	fmt.Printf("原始交易数据: '%s'\n", transactionData)

	// --- 签名过程 (由私钥持有者完成) ---

	// 1. 对交易数据进行哈希运算 (SHA256)
	//    我们签名的对象是数据的哈希值，而不是原始数据本身。
	messageBytes := []byte(transactionData)
	messageHash := sha256.Sum256(messageBytes) // Sum256返回的是[32]byte数组
	fmt.Printf("\n[1] 交易数据哈希 (SHA256): %s\n", hex.EncodeToString(messageHash[:]))

	// 2. 使用私钥签名哈希值
	//    我们调用 `blockchain_lib.go` 中新增的 Sign 函数。
	signature, err := Sign(privateKey, messageHash[:]) // 需要将数组转换为切片
	if err != nil {
		log.Fatalf("签名失败: %v", err)
	}
	fmt.Printf("\n[2] 生成的数字签名 (Hex): %s\n", hex.EncodeToString(signature))

	// --- 验签过程 (由网络中的任何节点完成) ---

	fmt.Println("\n--- 网络节点开始验证 ---")

	// 3. 节点使用公钥验证签名
	//    我们调用 `blockchain_lib.go` 中新增的 Verify 函数。
	//    节点需要拥有：原始数据哈希、签名、发送者的公钥。
	isValid := Verify(publicKey, messageHash[:], signature)

	if isValid {
		fmt.Println("[3] 验证成功! 👍")
		fmt.Println("    结论：签名有效，这笔交易确实由公钥所有者发起。")
	} else {
		fmt.Println("[3] 验证失败! 💀")
		fmt.Println("    结论：签名无效！交易可能是伪造的，或数据已被篡改。")
	}

	fmt.Println("\n核心概念:")
	fmt.Println("-> 签名是使用【私钥】对【数据哈希】进行的操作。")
	fmt.Println("-> 验证是使用【公钥】对【签名】和【数据哈希】进行的操作。")
	fmt.Println("-> 我们成功地将签名和验签功能也封装到了 `blockchain_lib.go` 中。")
}