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

// CLI responsible for processing command line arguments
type CLI struct {
	bc      *blockchain.Blockchain
	wallets *wallet.Wallets
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet                 - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS  - Get balance of ADDRESS")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Mines a new block with the transaction")
	fmt.Println("  printchain                   - Print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		fmt.Printf("%s\n", block.String())
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		

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

func (cli *CLI) send(from, to string, amount int) {
	fromWallet := cli.wallets.GetWallet(from)
	utxoSet := blockchain.UTXOSet{Blockchain: cli.bc}

	tx, err := blockchain.NewUTXOTransaction(&fromWallet, to, amount, &utxoSet)
	if err != nil {
		fmt.Printf("Failed to create transaction: %s\n", err)
		return
	}
	
	// For simplicity in this version, the sender of the transaction also mines the block.
	// A coinbase transaction is added to reward the miner.
	cbTx := blockchain.NewCoinbaseTX(from, "")
	txs := []*blockchain.Transaction{cbTx, tx}

	cli.bc.MineBlock(txs)
	fmt.Println("Success! Transaction sent.")
}


// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")


	switch os.Args[1] {
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
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
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

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}