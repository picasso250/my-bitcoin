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
	bc *blockchain.Blockchain
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet                 - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS  - Get balance of ADDRESS")
	fmt.Println("  addblock -data BLOCK_DATA    - Add a block to the blockchain (deprecated)")
	fmt.Println("  printchain                   - Print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// addBlock a temporary method to add blocks for testing
func (cli *CLI) addBlock(data string) {
	// For now, we will create a simple transaction to include in the block
	// This is a placeholder and doesn't create valid spendable outputs
	tx := &blockchain.Transaction{
		Vin:  []blockchain.TxInput{{Txid: []byte{}, Vout: -1, PubKey: []byte(data)}},
		Vout: []blockchain.TxOutput{{Value: 1, PubKeyHash: []byte("unspendable")}},
	}
	tx.SetID()
	cli.bc.MineBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success! (Note: addblock creates non-standard transactions)")
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
	wallets, err := wallet.NewWallets()
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}
	address := wallets.CreateWallet()
	wallets.SaveToFile()

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

	addBlockData := addBlockCmd.String("data", "", "Block data")
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
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
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