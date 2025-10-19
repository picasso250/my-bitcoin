// main.go
// 项目统一入口

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	args := os.Args[2:] // 传递给子命令的参数

	switch command {
	case "step11":
		RunStep11(args)
	case "step12":
		RunStep12() // 演示Go版本的核心库与数据结构
	case "step13":
		RunStep13() // 演示Go版本的UTXO交易创建与签名
	case "step14":
		RunStep14() // 演示矿工打包Coinbase与普通交易
	default:
		fmt.Printf("错误: 未知的命令 '%s'\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("用法: go run . <command> [arguments]")
	fmt.Println("\n可用命令:")
	fmt.Println("  step11    - 运行P2P Hello World示例")
	fmt.Println("  step12    - 演示Go版本的核心库与数据结构")
	fmt.Println("  step13    - 演示Go版本的UTXO交易创建与签名")
	fmt.Println("  step14    - 演示矿工打包Coinbase与普通交易")
	// 在这里添加未来步骤的说明
}