package main

import (
	"fmt"

	"my-blockchain/gocoin/blockchain"
)

func main() {
	// 定义一个矿工地址，用于接收创世块奖励和未来的挖矿奖励
	minerAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	// 1. 创建或加载区块链
	// 如果是第一次运行，它会自动生成一个包含Coinbase交易的创世区块
	bc := blockchain.NewBlockchain(minerAddress)
	defer bc.DB().Close() // 确保在程序结束时关闭数据库连接

	fmt.Println("--- Blockchain Ready ---")

	// 2. 准备一些待打包到新区块的交易
	// 在真实的系统中，这些交易将从内存池(mempool)中获取
	// 这里为了演示，我们手动创建一笔交易
	tx1 := &blockchain.Transaction{
		// 注意: 为简化，我们暂时不引用真实的UTXO。
		// 这笔交易在密码学上是无效的，但足以演示打包和挖矿。
		Vin: []blockchain.TxInput{{Txid: []byte{}, Vout: -1, PubKey: []byte("Alice's Money")}},
		Vout: []blockchain.TxOutput{
			{Value: 10, PubKeyHash: []byte("Bob")},
			{Value: 5, PubKeyHash: []byte("Alice")}, // 找零
		},
	}
	tx1.SetID()

	// 矿工需要在交易列表的开头加上自己的Coinbase交易
	coinbaseTx := blockchain.NewCoinbaseTX(minerAddress, "Mined by Gemini")
	transactions := []*blockchain.Transaction{coinbaseTx, tx1}

	// 3. 调用 MineBlock，它会打包交易、挖矿，然后将新区块添加到链上
	fmt.Println("\nAttempting to mine a new block with new transactions...")
	bc.MineBlock(transactions)
	fmt.Println("Success! New block has been added to the blockchain.")

	// 下一步: 我们将实现一个命令行工具来打印链的内容。
}