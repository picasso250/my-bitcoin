# step10_simulated_network_consensus.py
#
# 教学目标：在进入真实P2P网络前，用一个“上帝视角”的模拟器，
#           完整地、清晰地跑通一个去中心化网络达成共识的全过程。
#
# 核心逻辑：
# 1. 交易广播 -> 全网节点的mempool中出现待确认交易。
# 2. 矿工挖矿 -> 某个节点从自己的mempool中打包交易，创建新区块。
# 3. 区块广播 -> 新区块被广播至全网。
# 4. 验证与接受 -> 其他节点验证该区块的合法性。
# 5. 账本更新与共识达成 -> 所有诚实节点都接受新区块，更新自己的区块链和UTXO池，
#                       并清理mempool，最终达到一致的状态。

import hashlib
import json
from time import time
import copy # 用于深度复制对象

# --- 全局常量和工具函数 ---
DIFFICULTY_PREFIX = '000'
BLOCK_REWARD = 50

def calculate_hash(data_dict, sort_keys=True):
    """计算字典数据的SHA256哈希值"""
    payload = json.dumps(data_dict, sort_keys=sort_keys, separators=(',', ':'))
    return hashlib.sha256(payload.encode('utf-8')).hexdigest()

class Network:
    """
    模拟一个理想化的P2P网络。它知道所有节点，并负责将消息广播给它们。
    这是一个中心化的模拟器，用于简化教学，后续将被真实的libp2p替代。
    """
    def __init__(self):
        self.nodes = []

    def add_node(self, node):
        self.nodes.append(node)
        node.network = self # 让节点能访问网络

    def broadcast_transaction(self, sending_node, tx):
        print(f"[Network] 节点 {sending_node.id} 正在广播交易 {tx['id'][:5]}...")
        for node in self.nodes:
            if node is not sending_node:
                node.receive_transaction(copy.deepcopy(tx))

    def broadcast_block(self, sending_node, block):
        print(f"[Network] 节点 {sending_node.id} 正在广播区块 {block['hash'][:5]}...")
        for node in self.nodes:
            if node is not sending_node:
                node.receive_block(copy.deepcopy(block))

class Node:
    """模拟一个区块链网络中的独立节点"""
    def __init__(self, node_id):
        self.id = node_id
        self.blockchain = []
        self.utxo_pool = {}
        self.mempool = {}
        self.network = None # 将在添加到网络时被设置

    def add_genesis_block(self, genesis_block):
        """所有节点都从同一个创世区块开始"""
        self.blockchain.append(genesis_block)
        # 创世区块的Coinbase输出是第一个UTXO
        coinbase_tx = genesis_block['transactions'][0]
        tx_id = coinbase_tx['id']
        self.utxo_pool[f"{tx_id}_0"] = coinbase_tx['vout'][0]

    def receive_transaction(self, tx):
        """接收来自网络的广播交易"""
        # 简化验证：只检查交易是否已在mempool中
        if tx['id'] not in self.mempool:
            print(f"  - [节点 {self.id}] 收到新交易 {tx['id'][:5]}，已放入mempool。")
            self.mempool[tx['id']] = tx
    
    def receive_block(self, block):
        """接收来自网络的广播区块"""
        print(f"  - [节点 {self.id}] 收到新区块 {block['hash'][:5]}，开始验证...")
        
        # 1. 验证工作量证明
        if not block['hash'].startswith(DIFFICULTY_PREFIX):
            print(f"    - [节点 {self.id}] 验证失败: PoW无效。")
            return

        # 2. 验证前一区块哈希
        last_block_hash = self.blockchain[-1]['hash']
        if block['previous_hash'] != last_block_hash:
            print(f"    - [节点 {self.id}] 验证失败: 区块链不连续。")
            return
            
        print(f"    - [节点 {self.id}] 验证成功! 准备更新账本...")
        self.blockchain.append(block)
        self._update_utxo_and_mempool(block)

    def _update_utxo_and_mempool(self, block):
        """根据新区块中的交易更新自己的UTXO池和Mempool"""
        for tx in block['transactions']:
            # 1. 如果是Coinbase交易，直接创建新的UTXO
            if 'COINBASE' in tx['vin'][0]:
                 self.utxo_pool[f"{tx['id']}_0"] = tx['vout'][0]
                 continue

            # 2. 对于普通交易
            # a) 销毁花费的UTXO
            for vin in tx['vin']:
                utxo_key = f"{vin['txid']}_{vin['vout']}"
                if utxo_key in self.utxo_pool:
                    del self.utxo_pool[utxo_key]

            # b) 创建新的UTXO
            for i, vout in enumerate(tx['vout']):
                self.utxo_pool[f"{tx['id']}_{i}"] = vout

            # c) 从mempool中移除已被确认的交易
            if tx['id'] in self.mempool:
                del self.mempool[tx['id']]
        
        print(f"    - [节点 {self.id}] 账本更新完毕。UTXO池大小: {len(self.utxo_pool)}, Mempool大小: {len(self.mempool)}")
        
    def mine_block(self):
        """矿工节点的核心工作"""
        if not self.mempool:
            print(f"[节点 {self.id}] Mempool为空，无需挖矿。")
            return

        # 1. 从mempool中选择交易 (简化：选择第一个)
        tx_to_mine = list(self.mempool.values())[0]

        # 2. 创建Coinbase交易
        last_block = self.blockchain[-1]
        coinbase_tx = {
            'id': calculate_hash({'vin': 'COINBASE', 'nonce': last_block['index']+1}),
            'vin': [{'COINBASE': True}],
            'vout': [{'to_address': self.id, 'amount': BLOCK_REWARD}] # 简化，矿工费未计入
        }

        # 3. 组装候选区块
        candidate_block = {
            'index': last_block['index'] + 1,
            'transactions': [coinbase_tx, tx_to_mine],
            'previous_hash': last_block['hash'],
            'timestamp': time(),
            'nonce': 0
        }

        # 4. 工作量证明 (挖矿)
        print(f"[节点 {self.id}] 开始挖矿... 目标: 找到以 '{DIFFICULTY_PREFIX}' 开头的哈希。")
        nonce = 0
        while True:
            candidate_block['nonce'] = nonce
            block_hash = calculate_hash(candidate_block, sort_keys=False) # nonce变化，无需排序
            if block_hash.startswith(DIFFICULTY_PREFIX):
                print(f"  - [节点 {self.id}] 挖矿成功! Nonce: {nonce}")
                final_block = candidate_block
                final_block['hash'] = block_hash
                
                # 挖矿成功后，立即更新自己的账本并广播
                self.blockchain.append(final_block)
                self._update_utxo_and_mempool(final_block)
                self.network.broadcast_block(self, final_block)
                return
            nonce += 1
            
    def display_state(self):
        """打印节点的当前状态"""
        print(f"--- 节点 {self.id} 状态 ---")
        print(f"  - 区块链高度: {len(self.blockchain)}")
        print(f"  - Mempool中的交易数: {len(self.mempool)}")
        print(f"  - UTXO池中的条目数: {len(self.utxo_pool)}")
        print("-" * (20 + len(self.id)))

# --- 主仿真流程 ---

# 1. 初始化网络和节点
network = Network()
alice_node = Node("Alice")
bob_node = Node("Bob")
miner_node = Node("Miner")

network.add_node(alice_node)
network.add_node(bob_node)
network.add_node(miner_node)

# 2. 创建并分发创世区块
genesis_coinbase = {
    'id': calculate_hash({'vin': 'COINBASE', 'nonce': 0}),
    'vin': [{'COINBASE': True}],
    'vout': [{'to_address': 'Alice', 'amount': 100}] # 创世时给Alice一些钱
}
genesis_block = {
    'index': 0,
    'transactions': [genesis_coinbase],
    'previous_hash': '0'*64,
    'timestamp': 0,
    'nonce': 0
}
genesis_block['hash'] = calculate_hash(genesis_block, sort_keys=False)

for node in network.nodes:
    node.add_genesis_block(copy.deepcopy(genesis_block))

print("=== 初始状态 ===")
for node in network.nodes:
    node.display_state()
print("\n" + "="*40 + "\n")


# 3. 第1幕: Alice创建并广播交易
print("=== 第1幕: Alice向Bob转账并广播交易 ===")
# a) Alice找到自己的UTXO
alice_utxo_txid = genesis_coinbase['id']
# b) Alice构建交易
tx_draft = {
    'vin': [{'txid': alice_utxo_txid, 'vout': 0}],
    'vout': [
        {'to_address': 'Bob', 'amount': 20},
        {'to_address': 'Alice', 'amount': 80} # 找零
    ]
}
alice_tx = {
    'id': calculate_hash(tx_draft),
    **tx_draft
}
# c) Alice广播交易 (注意: 简化了签名过程)
alice_node.mempool[alice_tx['id']] = alice_tx # 自己的mempool也要放一份
alice_node.network.broadcast_transaction(alice_node, alice_tx)

print("\n--- 广播后，挖矿前 ---")
for node in network.nodes:
    node.display_state()
print("\n" + "="*40 + "\n")


# 4. 第2幕: 矿工挖矿并广播区块
print("=== 第2幕: 矿工打包交易并挖出新区块 ===")
miner_node.mine_block()


# 5. 第3幕: 最终共识
print("\n=== 第3幕: 最终共识达成 ===")
print("所有节点都接收并验证了新区块，更新了各自的账本。")
print("Alice的交易从Mempool中被清除，Bob收到了钱。")
for node in network.nodes:
    node.display_state()