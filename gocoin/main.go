package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	listenAddr = flag.String("listen", "", "P2P listen address, e.g. :7000")
	seedAddr   = flag.String("seed", "", "Optional seed peer, e.g. 127.0.0.1:7001")
)

func main() {
	flag.Parse()

	// wallets & blockchain setup (unchanged)
	wallets, err := NewWallets()
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	if len(addresses) == 0 {
		addr := wallets.CreateWallet()
		wallets.SaveToFile()
		fmt.Println("Created first wallet:", addr)
		addresses = append(addresses, addr)
	}
	bc := NewBlockchain(addresses[0])
	defer bc.DB().Close()

	// P2P layer
	var node *Node
	if *listenAddr != "" {
		node, err = NewNode(*listenAddr, bc)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("P2P listening on", *listenAddr)
	}
	if *seedAddr != "" && node != nil {
		if err := node.Connect(*seedAddr); err != nil {
			fmt.Println("Seed connect error:", err)
		} else {
			fmt.Println("Connected to seed", *seedAddr)
		}
	}

	cli := CLI{bc: bc, wallets: wallets, node: node}
	cli.Run()
}
