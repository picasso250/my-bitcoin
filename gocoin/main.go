package main

import (
	"my-blockchain/gocoin/blockchain"
)

func main() {
	// For the CLI, we can use a default address for now
	// This will be improved when we implement wallet management
	minerAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	bc := blockchain.NewBlockchain(minerAddress)
	defer bc.DB().Close()

	cli := CLI{bc}
	cli.Run()
}