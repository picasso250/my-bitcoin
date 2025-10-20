package main

import (
	"fmt"
	"log"
	"my-blockchain/gocoin/blockchain"
	"my-blockchain/gocoin/wallet"
	"os"
)

func main() {
	// Create or load wallets
	wallets, err := wallet.NewWallets()
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}

	// If no wallets exist, create one to be the miner's address
	addresses := wallets.GetAddresses()
	if len(addresses) == 0 {
		newAddress := wallets.CreateWallet()
		wallets.SaveToFile()
		fmt.Printf("No wallets found. Created a new one. Your address: %s\n", newAddress)
		addresses = append(addresses, newAddress)
	}

	// Get the first address as the miner address
	minerAddress := addresses[0]
	fmt.Printf("Using miner address: %s\n\n", minerAddress)

	bc := blockchain.NewBlockchain(minerAddress)
	defer bc.DB().Close()

	cli := CLI{bc}
	cli.Run()
}