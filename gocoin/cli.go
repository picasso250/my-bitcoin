package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"my-blockchain/gocoin/blockchain"
)

// CLI 负责处理命令行参数
type CLI struct {
	bc *blockchain.Blockchain
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA    - Add a block to the blockchain")
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
	// In the future, this will be replaced by a proper transaction creation mechanism
	tx := &blockchain.Transaction{
		Vin:  []blockchain.TxInput{{Txid: []byte{}, Vout: -1, PubKey: []byte(data)}},
		Vout: []blockchain.TxOutput{{Value: 50, PubKeyHash: []byte("reward")}},
	}
	tx.SetID()
	cli.bc.MineBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()

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

// Run parses command line arguments and runs commands
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

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
}