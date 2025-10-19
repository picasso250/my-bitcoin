// step11_p2p_hello_world.go
//
// 教学目标：使用 Go 语言实现一个最基础的 P2P “Hello, World” 应用。
//
// 核心逻辑：
// 1. 程序可以作为 “listener” 或 “dialer” 启动。
// 2. Listener 节点会启动并监听一个网络地址，打印自己的 P2P 地址。
// 3. Dialer 节点会使用 Listener 的 P2P 地址去连接它。
// 4. 连接成功后，Dialer 发送一条消息，Listener 收到后打印并回复一条消息。
//
// Go 语言的优势：
// - 强类型语言，编译时检查错误，更稳定。
// - 原生支持并发 (goroutine)，处理网络流非常自然。
// - go-libp2p 是官方参考实现，功能最全，社区最活跃。

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"
	// 修正: 导入路径已更新
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// PROTOCOL_ID 是我们自定义协议的唯一标识符。
// 网络中的节点通过这个ID来找到并使用我们定义的通信协议。
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
	// 创建一个新的 libp2p 主机 (Host)，它会自动监听一个可用的 TCP 端口。
	// `libp2p.ListenAddrStrings` 告诉主机要在哪个网络地址上进行监听。
	// "/ip4/127.0.0.1/tcp/0" 表示在本地 IPv4 地址上监听一个由操作系统自动选择的 TCP 端口。
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	if err != nil {
		log.Fatalf("创建主机失败: %s", err)
	}
	defer host.Close()

	// 为我们的聊天协议设置流处理器
	host.SetStreamHandler(PROTOCOL_ID, streamHandler)

	// 打印 Listener 节点的完整 P2P 地址。
	// 这个地址包含了节点的 Peer ID 和它正在监听的网络地址。
	// Dialer 将使用这个地址来连接我们。
	log.Println("--- Listener 节点已启动 ---")
	log.Println("请将以下任意一个地址复制给 Dialer 节点:")
	for _, addr := range host.Addrs() {
		fmt.Printf("- %s/p2p/%s\n", addr, host.ID())
	}
	log.Println("\n正在等待连接...")

	// 等待程序被终止 (例如，通过 Ctrl+C)
	<-ctx.Done()
}

// runDialer 启动 Dialer 节点，并连接到 Listener。
func runDialer(ctx context.Context, targetAddress string) {
	// Dialer 节点不需要监听，所以我们用 `libp2p.NoListenAddrs()` 创建它。
	host, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		log.Fatalf("创建主机失败: %s", err)
	}
	defer host.Close()

	log.Println("--- Dialer 节点已启动 ---")
	log.Printf("准备连接到: %s", targetAddress)

	// 1. 将目标地址字符串解析成 `multiaddr` 对象
	maddr, err := multiaddr.NewMultiaddr(targetAddress)
	if err != nil {
		log.Fatalf("无效的目标地址: %s", err)
	}

	// 2. 从 `multiaddr` 中提取 Peer ID 和网络地址
	addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatalf("解析地址信息失败: %s", err)
	}

	// 3. 连接到目标节点
	if err := host.Connect(ctx, *addrInfo); err != nil {
		log.Fatalf("连接失败: %s", err)
	}
	log.Printf("[Dialer] 已成功连接到 Peer ID: %s", addrInfo.ID.String())

	// 4. 打开一个新的流，并指定使用我们的聊天协议
	stream, err := host.NewStream(ctx, addrInfo.ID, PROTOCOL_ID)
	if err != nil {
		log.Fatalf("打开新流失败: %s", err)
	}
	defer stream.Close()

	// 5. 发送消息
	message := "Hello P2P world!\n"
	_, err = stream.Write([]byte(message))
	if err != nil {
		log.Printf("发送消息失败: %s", err)
		return
	}
	log.Printf("[Dialer] 已发送消息: '%s'", message)

	// 6. 读取回复
	reader := bufio.NewReader(stream)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("读取回复失败: %s", err)
		return
	}
	log.Printf("[Dialer] 收到回复: '%s'", response)

	log.Println("[Dialer] 任务完成。")
}

func main() {
	// 使用 Context 来优雅地处理程序的中断
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  go run . listener")
		fmt.Println("  go run . dialer <listener_multiaddress>")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "listener":
		runListener(ctx)
	case "dialer":
		if len(os.Args) < 3 {
			log.Fatal("错误: Dialer模式需要一个目标地址。")
		}
		targetAddr := os.Args[2]
		runDialer(ctx, targetAddr)
	default:
		log.Fatalf("错误: 未知的模式 '%s'", mode)
	}
}