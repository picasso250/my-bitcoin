// step11.go

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// PROTOCOL_ID 是我们自定义协议的唯一标识符。
const PROTOCOL_ID = "/my-chat/1.0.0"

// streamHandler 当 Listener 节点接收到一个新的连接时，这个函数会被自动调用。
func streamHandler(stream network.Stream) {
	// 获取发送方 Peer ID
	senderPeerID := stream.Conn().RemotePeer().String()
	log.Printf("[Listener] 收到来自 %s 的新连接", senderPeerID)

	// 使用 bufio.Reader 来方便地从流中读取数据
	reader := bufio.NewReader(stream)
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("读取消息失败: %s", err)
		return
	}
	log.Printf("[Listener] 收到消息: '%s'", message)

	// 准备并发送回复
	response := fmt.Sprintf("Hello back to you, %s!\n", senderPeerID)
	_, err = stream.Write([]byte(response))
	if err != nil {
		log.Printf("发送回复失败: %s", err)
	} else {
		log.Printf("[Listener] 已发送回复: '%s'", response)
	}

	// 关闭流
	stream.Close()
}

// runListener 启动并运行 Listener 节点。
func runListener(ctx context.Context) {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		log.Fatalf("创建主机失败: %s", err)
	}
	defer host.Close()

	host.SetStreamHandler(PROTOCOL_ID, streamHandler)

	log.Println("--- Listener 节点已启动 ---")
	log.Println("请将以下任意一个地址复制给 Dialer 节点:")
	for _, addr := range host.Addrs() {
		fmt.Printf("- %s/p2p/%s\n", addr, host.ID())
	}
	log.Println("\n正在等待连接...")

	<-ctx.Done()
}

// runDialer 启动 Dialer 节点，并连接到 Listener。
func runDialer(ctx context.Context, targetAddress string) {
	host, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		log.Fatalf("创建主机失败: %s", err)
	}
	defer host.Close()

	log.Println("--- Dialer 节点已启动 ---")
	log.Printf("准备连接到: %s", targetAddress)

	maddr, err := multiaddr.NewMultiaddr(targetAddress)
	if err != nil {
		log.Fatalf("无效的目标地址: %s", err)
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatalf("解析地址信息失败: %s", err)
	}

	if err := host.Connect(ctx, *addrInfo); err != nil {
		log.Fatalf("连接失败: %s", err)
	}
	log.Printf("[Dialer] 已成功连接到 Peer ID: %s", addrInfo.ID.String())

	stream, err := host.NewStream(ctx, addrInfo.ID, PROTOCOL_ID)
	if err != nil {
		log.Fatalf("打开新流失败: %s", err)
	}
	defer stream.Close()

	message := "Hello P2P world!\n"
	_, err = stream.Write([]byte(message))
	if err != nil {
		log.Printf("发送消息失败: %s", err)
		return
	}
	log.Printf("[Dialer] 已发送消息: '%s'", message)

	reader := bufio.NewReader(stream)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("读取回复失败: %s", err)
		return
	}
	log.Printf("[Dialer] 收到回复: '%s'", response)

	log.Println("[Dialer] 任务完成。")
}

// RunStep11 是第11步的入口函数，由主程序 main.go 调用。
func RunStep11(args []string) {
	// 使用 Context 来优雅地处理程序的中断
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(args) < 1 {
		fmt.Println("用法: go run . step11 <mode> [args]")
		fmt.Println("  <mode>: listener 或 dialer")
		return
	}

	mode := args[0]

	switch mode {
	case "listener":
		runListener(ctx)
	case "dialer":
		if len(args) < 2 {
			log.Fatal("错误: Dialer模式需要一个目标地址。")
		}
		targetAddr := args[1]
		runDialer(ctx, targetAddr)
	default:
		log.Fatalf("错误: 未知的模式 '%s'", mode)
	}
}