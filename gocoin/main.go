package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const usage = `
GoCoin 三角色 CLI
  wallet  - 纯离线钱包操作
  node    - 消费者：同步、查余额、转账
  miner   - 全节点+矿工

Usage:
  go run . wallet <subcmd>
  go run . node  [--listen addr] [--seed addr] [--wallet addr] <subcmd>
  go run . miner [--listen addr] [--seed addr] --coinbase addr
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "wallet":
		runWalletCmd(os.Args[2:])
	case "node":
		runNodeCmd(os.Args[2:])
	case "miner":
		runMinerCmd(os.Args[2:])
	default:
		fmt.Print(usage)
		os.Exit(1)
	}
}

/* ---------- wallet ---------- */
func runWalletCmd(args []string) {
	fs := flag.NewFlagSet("wallet", flag.ExitOnError)
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}
	switch args[0] {
	case "create":
		if _, err := os.Stat(walletFile); err == nil {
			log.Fatalf("钱包已存在（%s），如需新建请先备份并删除原文件", walletFile)
		}
		wallets, _ := NewWallets()
		addr := wallets.CreateWallet()
		wallets.SaveToFile()
		fmt.Println(addr)
	case "list":
		wallets, err := NewWallets()
		if err != nil && !os.IsNotExist(err) {
			log.Panic(err)
		}
		for _, a := range wallets.GetAddresses() {
			fmt.Println(a)
		}
	default:
		fs.Usage()
	}
}

/* ---------- node ---------- */
func runNodeCmd(args []string) {
	fs := flag.NewFlagSet("node", flag.ExitOnError)
	walletStr := fs.String("wallet", "", "Wallet address to use")
	_ = fs.Parse(args)

	if *walletStr == "" {
		log.Fatal("-wallet required")
	}
	wallets, err := NewWallets()
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}
	if wallets.GetWallet(*walletStr).PrivateKey.D == nil {
		log.Fatal("wallet not found, run 'wallet create' first")
	}
	// 懒加载链
	var bc *Blockchain
	openChain := func() *Blockchain {
		if bc == nil {
			bc = NewBlockchain(*walletStr)
		}
		return bc
	}

	sub := fs.Arg(0)
	switch sub {
	case "balance":
		chain := openChain()
		defer chain.DB().Close()
		pubHash, _ := DecodeAddress(*walletStr)
		utxo := UTXOSet{chain}
		sum := 0
		for _, u := range utxo.FindUTXO(pubHash) {
			sum += u.Value
		}
		fmt.Println(sum)
	case "send":
		if fs.NArg() < 3 {
			log.Fatal("usage: node ... send --to ADDR --amount N")
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
		w := wallets.GetWallet(*walletStr)
		utxo := UTXOSet{chain}
		tx, err := NewUTXOTransaction(&w, *to, *amt, &utxo)
		if err != nil {
			log.Panic(err)
		}
		// 简化：直接本地挖矿；后续可改广播
		chain.MineBlock([]*Transaction{tx})
		fmt.Println("tx mined")
	default:
		log.Fatal("unknown subcmd: balance | send")
	}
}

/* ---------- miner ---------- */
func runMinerCmd(args []string) {
	fs := flag.NewFlagSet("miner", flag.ExitOnError)
	coinbase := fs.String("coinbase", "", "Coinbase reward address")
	_ = fs.Parse(args)

	// 简化矿工功能：直接使用第一个钱包地址作为 coinbase
	wallets, err := NewWallets()
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}
	
	// 获取第一个钱包地址作为 coinbase
	addresses := wallets.GetAddresses()
	if len(addresses) == 0 {
		log.Fatal("没有找到钱包，请先运行 'go run . wallet create' 创建钱包")
	}
	
	coinbaseAddr := addresses[0]
	if *coinbase != "" {
		coinbaseAddr = *coinbase // 如果指定了 coinbase，使用指定的地址
	}
	
	fmt.Printf("使用钱包地址 %s 作为 coinbase\n", coinbaseAddr)
	
	bc := NewBlockchain(coinbaseAddr)
	defer bc.DB().Close()

	// 计算当前区块链高度
	currentHeight := 0
	bci := bc.Iterator()
	for {
		block := bci.Next()
		if block == nil {
			break
		}
		currentHeight++
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	fmt.Println("矿工启动，按 Ctrl-C 停止")
	for {
		// 仅打包 coinbase 交易
		bc.MineBlock([]*Transaction{})
		currentHeight++
		// 获取最新挖出的区块哈希
		bci := bc.Iterator()
		latestBlock := bci.Next()
		if latestBlock != nil {
			fmt.Printf("已挖出区块 #%d  %x\n", currentHeight, latestBlock.Hash)
		}
		time.Sleep(2 * time.Second) // 避免占满 CPU
	}
}
