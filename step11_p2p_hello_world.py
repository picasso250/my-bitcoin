# step11_p2p_hello_world.py
#
# 教学目标：迈出从“模拟”到“真实”的第一步，实现一个最简单的P2P网络应用。
#
# 核心概念:
# 1. Peer ID: 每个运行的节点都有一个唯一的身份标识，就像是它的指纹。
# 2. Multiaddress: 一个统一的、自描述的地址格式，包含了节点的网络位置和身份。
#    例如: /ip4/127.0.0.1/tcp/12345/p2p/Qm...
# 3. Listener (监听者): 一个被动等待连接的节点。
# 4. Dialer (拨号者): 一个主动发起连接到Listener的节点。
# 5. Stream (流): 两个节点成功连接后，它们之间会打开一个双向通信通道，称为流。
#
# 我们将使用 `libp2p` 库，这是一个专业的P2P网络框架。
#
# 前置准备 (请在你的终端中运行):
# pip install "py-libp2p[yamux,mplex,noise]"

import asyncio
import sys

from libp2p import new_host
from libp2p.network.stream.net_stream_interface import INetStream
from libp2p.peer.peerinfo import info_from_p2p_addr
from libp2p.typing import TProtocol

# 定义一个自定义的协议ID。
# 就像网站有HTTP协议，我们的P2P应用也需要一个名字来识别服务。
PROTOCOL_ID = TProtocol("/my-chat/1.0.0")


async def stream_handler(stream: INetStream) -> None:
    """
    当Listener节点接收到一个新的连接时，这个函数会被自动调用。
    它负责处理来自Dialer的数据。
    """
    # 从流中读取数据 (Dialer发送的消息)
    message = await stream.read()
    message_str = message.decode('utf-8')
    
    sender_peer_id = stream.muxed_conn.peer_id.to_base58()
    print(f"\n[Listener] 收到来自 {sender_peer_id[:6]}... 的消息: '{message_str}'")
    
    # 准备并发送一个回复
    response = f"Hello back to you, {sender_peer_id[:6]}...!".encode('utf-8')
    await stream.write(response)
    print(f"[Listener] 已发送回复: '{response.decode('utf-8')}'")
    
    # 关闭流
    await stream.close()


async def run_listener():
    """启动并运行Listener节点。"""
    # 1. 创建一个libp2p主机 (节点)
    #    它会自动监听一个可用的TCP端口
    listen_host = new_host()
    
    # 2. 设置流处理器
    #    告诉节点，当收到基于我们自定义协议的连接时，应该调用 `stream_handler` 函数
    listen_host.set_stream_handler(PROTOCOL_ID, stream_handler)
    
    # 3. 打印节点的地址，以便Dialer知道如何连接
    print("--- Listener 节点已启动 ---")
    print("请将以下任意一个地址复制给Dialer节点:")
    for addr in listen_host.get_addrs():
        print(f"- {addr}/p2p/{listen_host.get_id().to_base58()}")
    print("\n正在等待连接...")

    # 保持节点持续运行
    while True:
        await asyncio.sleep(1)


async def run_dialer(target_address: str):
    """启动Dialer节点，并连接到Listener。"""
    # 1. 创建Dialer自己的主机
    dial_host = new_host()
    
    print(f"--- Dialer 节点已启动 ---")
    print(f"准备连接到: {target_address}")

    try:
        # 2. 从目标地址中解析出Peer ID和网络地址
        target_info = info_from_p2p_addr(target_address)
        
        # 3. 连接到目标节点
        await dial_host.connect(target_info)
        print(f"[Dialer] 已成功连接到 Peer ID: {target_info.peer_id.to_base58()[:6]}...")
        
        # 4. 在连接上打开一个新流，使用我们的自定义协议
        stream = await dial_host.new_stream(target_info.peer_id, [PROTOCOL_ID])
        
        # 5. 发送消息
        message = b'Hello P2P world!'
        await stream.write(message)
        print(f"[Dialer] 已发送消息: '{message.decode('utf-8')}'")
        
        # 6. 等待并读取回复
        response = await stream.read()
        print(f"[Dialer] 收到回复: '{response.decode('utf-8')}'")
        
        # 7. 关闭流和主机
        await stream.close()
        await dial_host.close()
        print("[Dialer] 任务完成，已关闭。")

    except Exception as e:
        print(f"[Dialer] 连接失败: {e}")
        await dial_host.close()


def main():
    """
    主函数，根据命令行参数决定是作为Listener还是Dialer运行。
    """
    if len(sys.argv) < 2:
        print("用法:")
        print("  python step11_p2p_hello_world.py listener")
        print("  python step11_p2p_hello_world.py dialer <listener_multiaddress>")
        return

    mode = sys.argv[1]

    try:
        if mode == 'listener':
            asyncio.run(run_listener())
        elif mode == 'dialer':
            if len(sys.argv) < 3:
                print("错误: Dialer模式需要一个目标地址。")
                return
            target_addr = sys.argv[2]
            asyncio.run(run_dialer(target_addr))
        else:
            print(f"错误: 未知的模式 '{mode}'")
    except KeyboardInterrupt:
        print("\n程序被用户中断。")


if __name__ == "__main__":
    main()