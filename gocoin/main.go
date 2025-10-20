package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const usage = `
GoCoin 单钱包 CLI
  wallet  - 已移除，自动使用 wallet.dat
  node    - 消费者：同步、查余额、转账
  miner   - 全节点+矿工

Usage:
  go run . node  <subcmd>
  go run . miner [--coinbase addr]
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "node":
		runNodeCmd(os.Args[2:])
	case "miner":
		runMinerCmd(os.Args[2:])
	default:
		fmt.Print(usage)
		os.Exit(1)
	}
}

/* ---------- node ---------- */
func runNodeCmd(args []string) {
	fs := flag.NewFlagSet("node", flag.ExitOnError)
	_ = fs.Parse(args)

	wallet := LoadOrCreateDefaultWallet()
	addr := wallet.GetAddress()

	openChain := func() *Blockchain {
		return NewBlockchain(addr)
	}

	sub := fs.Arg(0)
	switch sub {
	case "balance":
		chain := openChain()
		defer chain.DB().Close()
		pubHash, _ := DecodeAddress(addr)
		utxo := UTXOSet{chain}
		sum := 0
		for _, u := range utxo.FindUTXO(pubHash) {
			sum += u.Value
		}
		fmt.Println(sum)
	case "send":
		if fs.NArg() < 3 {
			log.Fatal("usage: node send --to ADDR --amount N")
		}
		sendFs := flag.NewFlagSet("send", flag.ExitOnError)
		to := sendFs.String("to", "", "destination address")
		amt := sendFs.Int("amount", 0, "amount to send")
		_ = sendFs.Parse(fs.Args()[1:])
		if *to == "" || *amt <= 0 {
			log.Fatal("-to and -amount required")
		}
		chain := openChain()
		defer chain.DB().Close()
		utxo := UTXOSet{chain}
		tx, err := NewUTXOTransaction(wallet, *to, *amt, &utxo)
		if err != nil {
			log.Panic(err)
		}
		chain.MineBlock([]*Transaction{tx})
		fmt.Println("tx mined")
	default:
		log.Fatal("unknown subcmd: balance | send")
	}
}

/* ---------- miner ---------- */
func runMinerCmd(args []string) {
	fs := flag.NewFlagSet("miner", flag.ExitOnError)
	coinbase := fs.String("coinbase", "", "Coinbase reward address (optional)")
	_ = fs.Parse(args)

	wallet := LoadOrCreateDefaultWallet()
	coinbaseAddr := wallet.GetAddress()
	if *coinbase != "" {
		coinbaseAddr = *coinbase
	}

	bc := NewBlockchain(coinbaseAddr)
	defer bc.DB().Close()

	currentHeight := 0
	bci := bc.Iterator()
	for {
		blk := bci.Next()
		if blk == nil {
			break
		}
		currentHeight++
		if len(blk.PrevBlockHash) == 0 {
			break
		}
	}

	fmt.Println("Miner started, Ctrl-C to stop")
	for {
		bc.MineBlock([]*Transaction{})
		currentHeight++
		bci := bc.Iterator()
		latestBlock := bci.Next()
		if latestBlock != nil {
			fmt.Printf("Mined block #%d  %x\n", currentHeight, latestBlock.Hash)
		}
		time.Sleep(2 * time.Second)
	}
}