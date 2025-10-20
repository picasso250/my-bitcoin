package main

import (
	"encoding/hex"
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
  chansimu - 教学演示：Miner ↔ Thunder 互转

Usage:
  go run . node  <subcmd>
  go run . miner [--coinbase addr]
  go run . chansimu
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
	case "chansimu":
		runChansimu()
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

/*
	----------------------------------------------------------
	  3b 教学版：Miner ↔ Thunder 互转（Channel 广播）

----------------------------------------------------------
*/
func runChansimu() {
	// 1. 矿工钱包（已存在）
	minerWallet := LoadOrCreateDefaultWallet()
	minerAddr := minerWallet.GetAddress()
	minerPubHash, _ := DecodeAddress(minerAddr)

	// 2. Thunder 钱包（运行时生成，不落盘）
	thunderPriv, thunderPub := newKeyPair()
	thunderAddr := addressFromPubKey(thunderPub)
	thunderPubHash, _ := DecodeAddress(thunderAddr)
	fmt.Println("Thunder address:", thunderAddr)

	// 3. 区块链与 UTXO 集
	bc := NewBlockchain(minerAddr)
	defer bc.DB().Close()
	utxo := UTXOSet{bc}

	// 创建独立的交易池
	txPool := NewTxPool()

	// 4. 先给矿工挖一个空块，拿到 50 币 Coinbase
	fmt.Println("Genesis empty block for Miner coinbase...")
	bc.MineBlock([]*Transaction{})

	txCh := make(chan *Transaction) // 无缓冲广播通道

	// 5. 矿工 goroutine：每 2s 挖矿 + 条件转账
	go func() {
		for {
			// 5.1 收通道交易入池
			for {
				select {
				case tx := <-txCh:
					if bc.VerifyTransaction(tx) {
						txPool.Add(tx)
					}
				default:
					goto checkSend
				}
			}
		checkSend:
			// 5.2 如果自己有 ≥3 币，给 Thunder 转 2 币（含 1 手续费）
			minerBal := sumUTXO(utxo.FindUTXO(minerPubHash))
			if minerBal >= 3 {
				tx, _ := NewUTXOTransaction(minerWallet, thunderAddr, 2, &utxo)
				txCh <- tx
			}
			// 5.3 挖矿
			hashes := txPool.GetAllHashes()
			var txs []*Transaction
			for _, h := range hashes {
				txs = append(txs, txPool.Get(h))
			}
			bc.MineBlock(txs)
			height := getHeight(bc)
			fmt.Printf("Mined block #%d  Miner:%d  Thunder:%d\n",
				height, sumUTXO(utxo.FindUTXO(minerPubHash)), sumUTXO(utxo.FindUTXO(thunderPubHash)))
			time.Sleep(2 * time.Second)
		}
	}()

	// 6. Thunder goroutine：每 5s 条件转账
	go func() {
		for {
			time.Sleep(5 * time.Second)
			thunderBal := sumUTXO(utxo.FindUTXO(thunderPubHash))
			if thunderBal >= 3 {
				// 构造交易：给矿工转 2 币，1 币手续费
				tx, _ := NewUTXOTransaction(&Wallet{
					PrivateKey: thunderPriv,
					PublicKey:  thunderPub,
				}, minerAddr, 2, &utxo)
				txCh <- tx
				fmt.Println("Thunder: send 2 to Miner")
			}
		}
	}()

	fmt.Println("Chansimu started, Ctrl-C to stop")
	select {}
}

/* ---------- 辅助：生成 Thunder 地址 ---------- */

func addressFromPubKey(pubKey []byte) string {
	pubKeyHash := HashPubKey(pubKey)
	versioned := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versioned)
	full := append(versioned, checksum...)
	return hex.EncodeToString(full)
}

/* ---------- 辅助：求 UTXO 总额 ---------- */
func sumUTXO(outs []TxOutput) int {
	sum := 0
	for _, o := range outs {
		sum += o.Value
	}
	return sum
}

func getHeight(bc *Blockchain) int {
	height := 0
	it := bc.Iterator()
	for {
		blk := it.Next()
		if blk == nil || len(blk.PrevBlockHash) == 0 {
			break
		}
		height++
	}
	return height
}
