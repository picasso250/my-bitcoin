package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"my-blockchain/gocoin/blockchain"
	"my-blockchain/gocoin/wallet"
)

// CLI 负责处理命令行参数
type CLI struct {
	bc      *blockchain.Blockchain
	wallets *wallet.Wallets // Add wallets to CLI
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet                 - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS  - Get balance of ADDRESS")
	fmt.Println("  addblock                     - Mines a new block with a coinbase reward")
	fmt.Println("  printchain                   - Print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// addBlock now correctly creates a coinbase transaction and mines a block.
func (cli *CLI) addBlock() {
	// For simplicity, we'll use the first address in the wallet as the miner address.
	// This mirrors the logic in main.go
	addresses := cli.wallets.GetAddresses()
	if len(addresses) == 0 {
		fmt.Println("No addresses found. Please create a wallet first.")
		os.Exit(1)
	}
	minerAddress := addresses[0]

	fmt.Printf("Mining a new block, reward to: %s\n", minerAddress)

	// Create the coinbase transaction
	coinbaseTx := blockchain.NewCoinbaseTX(minerAddress, "")

	// Mine the block with the coinbase transaction
	cli.bc.MineBlock([]*blockchain.Transaction{coinbaseTx})

	fmt.Println("Success! New block mined.")
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createWallet() {
	address := cli.wallets.CreateWallet()
	cli.wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) getBalance(address string) {
	pubKeyHash, err := wallet.DecodeAddress(address)
	if err != nil {
		fmt.Println("ERROR: Invalid address")
		log.Panic(err)
	}

	utxoSet := blockchain.UTXOSet{Blockchain: cli.bc}
	utxos := utxoSet.FindUTXO(pubKeyHash)

	balance := 0
	for _, out := range utxos {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// Run parses command line arguments and runs commands
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		cli.addBlock()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
}