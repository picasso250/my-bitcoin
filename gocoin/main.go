package main

import (
	"fmt"

	"my-blockchain/gocoin/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()
	defer bc.DB().Close() // 确保在程序结束时关闭数据库连接

	fmt.Println("成功创建或加载区块链!")

	// 示例：添加一个区块 (将在下一步通过命令行实现)
	// bc.AddBlock("Send 1 BTC to Ivan")
	// bc.AddBlock("Send 2 more BTC to Ivan")

	// 我们将在下一步实现一个迭代器来打印所有区块
}