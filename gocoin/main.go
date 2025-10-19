package main

import (
	"my-blockchain/gocoin/blockchain"
)

func main() {
	// 这个地址现在只是一个占位符，代表矿工的身份
	// 在后续阶段，我们可以从钱包文件中选择一个地址来使用
	minerAddress := "miner-reward-address"

	bc := blockchain.NewBlockchain(minerAddress)
	defer bc.DB().Close()

	cli := CLI{bc}
	cli.Run()
}