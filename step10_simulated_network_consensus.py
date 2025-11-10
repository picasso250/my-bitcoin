# step10_simulated_network_consensus.py
#
# 教学目标：在进入真实P2P网络前，用一个“上帝视角”的模拟器，
#           完整地、清晰地跑通一个去中心化网络达成共识的全过程。
#
# 最终重构版：
# - 节点拥有真实的密钥和地址。
# - 交易经过严格的UTXO选择、找零计算和数字签名。
# - 节点之间广播的是带签名的合法交易。
# - 节点会严格验证收到的交易（UTXO存在性、签名有效性、价值守恒）。
# - 矿工会正确计算并收取所有交易的矿工费。
# - 所有节点在接收到新区块后，都会独立验证并更新自己的账本，最终达成共识。

import hashlib
import json
from time import time
import copy
import ecdsa

# --- 0. 核心依赖与函数 (从 bitcoin_lib.py 搬入，使其可独立运行) ---

def generate_address(public_key_bytes):
    """从一个压缩公钥字节生成简化的十六进制P2PKH地址。"""
    sha256_hash = hashlib.sha256(public_key_bytes).digest()
    ripemd160 = hashlib.new('ripemd160')
    ripemd160.update(sha256_hash)
    public_key_hash = ripemd160.digest()
    version_byte = b'\x00'
    versioned_hash = version_byte + public_key_hash
    checksum_hash_1 = hashlib.sha256(versioned_hash).digest()
    checksum_hash_2 = hashlib.sha256(checksum_hash_1).digest()
    checksum = checksum_hash_2[:4]
    final_address_bytes = versioned_hash + checksum
    return final_address_bytes.hex()

def calculate_tx_id(tx_data):
    """计算交易ID (双SHA256)"""
    # 注意：为了ID计算的稳定性，必须使用确定性序列化
    tx_string = json.dumps(tx_data, sort_keys=True, separators=(',', ':'))
    return hashlib.sha256(hashlib.sha256(tx_string.encode('utf-8')).digest()).hexdigest()

def calculate_block_hash(block_header_data):
    """计算区块头的哈希"""
    header_string = json.dumps(block_header_data, sort_keys=True, separators=(',', ':'))
    return hashlib.sha256(hashlib.sha256(header_string.encode('utf-8')).digest()).hexdigest()


# --- 全局常量 ---
DIFFICULTY_PREFIX = '000'
BLOCK_REWARD = 50
CURVE = ecdsa.SECP256k1

# --- 模拟网络层 ---
class Network:
    """模拟一个理想化的P2P网络，负责消息广播。"""
    def __init__(self):
        self.nodes = []

    def add_node(self, node):
        self.nodes.append(node)
        node.network = self

    def broadcast_transaction(self, sending_node, tx):
        print(f"[Network] 节点 {sending_node.id} 正在广播交易 {tx['id'][:10]}...")
        for node in self.nodes:
            if node is not sending_node:
                node.receive_transaction(copy.deepcopy(tx))

    def broadcast_block(self, sending_node, block):
        block_hash = block['header']['hash']
        print(f"[Network] 节点 {sending_node.id} 正在广播区块 {block_hash[:10]}...")
        for node in self.nodes:
            if node is not sending_node:
                node.receive_block(copy.deepcopy(block))

# --- 节点/钱包核心类 ---
class Node:
    """
    模拟一个“钱包+节点”的混合体。
    它有自己的身份（密钥），也维护着一份独立的区块链账本。
    """
    def __init__(self, node_id):
        self.id = node_id
        self.private_key = ecdsa.SigningKey.generate(curve=CURVE)
        self.public_key = self.private_key.get_verifying_key()
        self.address = generate_address(self.public_key.to_string('compressed'))
        
        self.blockchain = []
        self.utxo_pool = {}
        self.mempool = {}
        self.network = None
        print(f"  - 已创建节点: {self.id}, 地址: {self.address[:10]}...")

    def add_genesis_block(self, genesis_block):
        """所有节点从同一个创世区块开始"""
        self.blockchain.append(genesis_block)
        coinbase_tx = genesis_block['transactions'][0]
        tx_id = coinbase_tx['id']
        self.utxo_pool[f"{tx_id}_0"] = coinbase_tx['vout'][0]

    def create_transaction(self, recipient_address, amount, fee):
        """扫描UTXO池，构建一笔交易草稿。"""
        # 1. 筛选属于自己的UTXO并计算总余额
        my_utxos = []
        current_balance = 0
        for utxo_key, utxo_data in self.utxo_pool.items():
            if utxo_data['to_address'] == self.address:
                my_utxos.append((utxo_key, utxo_data))
                current_balance += utxo_data['amount']

        # 2. 检查余额是否足够
        total_to_spend = amount + fee
        if current_balance < total_to_spend:
            print(f"[节点 {self.id}] 错误: 余额不足 ({current_balance})，需要 {total_to_spend}")
            return None

        # 3. 选择UTXO来构建输入 (vin)
        vin = []
        amount_collected = 0
        for utxo_key, utxo_data in my_utxos:
            txid, vout_index = utxo_key.split('_')
            vin.append({'txid': txid, 'vout': int(vout_index), 'scriptSig': None})
            amount_collected += utxo_data['amount']
            if amount_collected >= total_to_spend:
                break
        
        # 4. 构建输出 (vout)
        vout = [{'to_address': recipient_address, 'amount': amount}]
        change = amount_collected - total_to_spend
        if change > 0:
            vout.append({'to_address': self.address, 'amount': change})

        return {'vin': vin, 'vout': vout}

    def sign_transaction(self, tx_draft):
        """对交易草稿进行签名，生成最终交易。"""
        # 1. 为每个输入签名
        tx_to_sign_string = json.dumps(tx_draft, sort_keys=True, separators=(',', ':'))
        tx_hash = hashlib.sha256(tx_to_sign_string.encode('utf-8')).digest()
        signature = self.private_key.sign(tx_hash)

        # 2. 将签名和公钥放入scriptSig
        # (简化模型：所有输入共享同一个签名)
        pubkey_hex = self.public_key.to_string('compressed').hex()
        for vin_item in tx_draft['vin']:
            vin_item['scriptSig'] = {
                'signature': signature.hex(),
                'publicKey': pubkey_hex
            }
        
        # 3. 计算最终交易ID
        signed_tx = tx_draft
        signed_tx['id'] = calculate_tx_id(signed_tx)
        return signed_tx

    def validate_transaction(self, tx):
        """严格验证一笔交易的合法性。"""
        # a. 交易哈希和ID是否正确
        tx_copy = copy.deepcopy(tx)
        tx_id = tx_copy.pop('id')
        if calculate_tx_id(tx_copy) != tx_id:
            print(f"    - [节点 {self.id}] 验证失败: 交易ID {tx_id[:10]}... 无效。")
            return False

        total_in = 0
        total_out = 0
        
        # b. 验证每个输入
        tx_draft_for_sig_verify = {'vin': copy.deepcopy(tx['vin']), 'vout': tx['vout']}
        for i, vin in enumerate(tx['vin']):
            # b.1 输入的UTXO是否存在于该节点的utxo_pool中
            utxo_key = f"{vin['txid']}_{vin['vout']}"
            if utxo_key not in self.utxo_pool:
                print(f"    - [节点 {self.id}] 验证失败: 输入的UTXO {utxo_key} 不存在或已被花费。")
                return False
            
            utxo_spent = self.utxo_pool[utxo_key]
            total_in += utxo_spent['amount']

            # b.2 每个输入的数字签名是否有效
            script_sig = vin.pop('scriptSig')
            tx_draft_for_sig_verify['vin'][i]['scriptSig'] = None # 准备用于验证的草稿
            
            # 从scriptSig中获取公钥和签名
            signature = bytes.fromhex(script_sig['signature'])
            pubkey_bytes = bytes.fromhex(script_sig['publicKey'])
            vk = ecdsa.VerifyingKey.from_string(pubkey_bytes, curve=CURVE, hashfunc=hashlib.sha256)
            
            # 验证公钥是否能生成UTXO的地址
            if generate_address(pubkey_bytes) != utxo_spent['to_address']:
                print(f"    - [节点 {self.id}] 验证失败: 公钥与UTXO地址不匹配。")
                return False

            # 验证签名
            tx_to_verify_string = json.dumps(tx_draft_for_sig_verify, sort_keys=True, separators=(',', ':'))
            tx_hash_to_verify = hashlib.sha256(tx_to_verify_string.encode('utf-8')).digest()
            try:
                if not vk.verify(signature, tx_hash_to_verify):
                    print(f"    - [节点 {self.id}] 验证失败: 数字签名无效。")
                    return False
            except ecdsa.BadSignatureError:
                print(f"    - [节点 {self.id}] 验证失败: 数字签名格式错误。")
                return False
            
            vin['scriptSig'] = script_sig # 恢复原样

        # c. 输入总金额是否大于等于输出总金额
        for vout in tx['vout']:
            total_out += vout['amount']
        
        if total_in < total_out:
            print(f"    - [节点 {self.id}] 验证失败: 输入({total_in}) < 输出({total_out})。")
            return False
            
        return True

    def receive_transaction(self, tx):
        """接收并验证来自网络的广播交易"""
        tx_id = tx['id']
        if tx_id in self.mempool:
            return
        
        print(f"  - [节点 {self.id}] 收到新交易 {tx_id[:10]}..., 开始验证。")
        if self.validate_transaction(tx):
            print(f"    - [节点 {self.id}] 交易 {tx_id[:10]}... 验证通过，已放入mempool。")
            self.mempool[tx_id] = tx
    
    def receive_block(self, block):
        """接收并验证来自网络的广播区块"""
        header = block['header']
        block_hash = header['hash']
        print(f"  - [节点 {self.id}] 收到新区块 {block_hash[:10]}..., 开始验证...")
        
        # 1. 验证工作量证明
        if not block_hash.startswith(DIFFICULTY_PREFIX):
            print(f"    - [节点 {self.id}] 验证失败: PoW无效。")
            return
        if calculate_block_hash({k:v for k,v in header.items() if k != 'hash'}) != block_hash:
            print(f"    - [节点 {self.id}] 验证失败: 区块哈希值不正确。")
            return

        # 2. 验证前一区块哈希
        last_block_hash = self.blockchain[-1]['header']['hash']
        if header['previous_hash'] != last_block_hash:
            print(f"    - [节点 {self.id}] 验证失败: 区块链不连续。")
            return
        
        # 3. 验证区块内的所有交易
        temp_utxo_pool = copy.deepcopy(self.utxo_pool)
        for i, tx in enumerate(block['transactions']):
            if i == 0: # Coinbase交易跳过签名验证
                continue
            if not self.validate_transaction(tx):
                print(f"    - [节点 {self.id}] 验证失败: 区块中包含无效交易 {tx['id'][:10]}...")
                return

        print(f"    - [节点 {self.id}] 区块验证成功! 准备更新账本...")
        self.blockchain.append(block)
        self._update_utxo_and_mempool(block)

    def _update_utxo_and_mempool(self, block):
        """根据新区块中的交易更新自己的UTXO池和Mempool"""
        for tx in block['transactions']:
            tx_id = tx['id']
            # a) 销毁花费的UTXO
            if 'COINBASE' not in tx['vin'][0].get('txid', ''):
                for vin in tx['vin']:
                    utxo_key = f"{vin['txid']}_{vin['vout']}"
                    if utxo_key in self.utxo_pool:
                        del self.utxo_pool[utxo_key]

            # b) 创建新的UTXO
            for i, vout in enumerate(tx['vout']):
                self.utxo_pool[f"{tx_id}_{i}"] = vout

            # c) 从mempool中移除已被确认的交易
            if tx_id in self.mempool:
                del self.mempool[tx_id]
        
        print(f"    - [节点 {self.id}] 账本更新完毕。UTXO池大小: {len(self.utxo_pool)}, Mempool大小: {len(self.mempool)}")
        
    def mine_block(self):
        """矿工节点的核心工作：打包交易、计算矿工费、挖矿、广播。"""
        if not self.mempool:
            print(f"[节点 {self.id}] Mempool为空，无需挖矿。")
            return

        # 1. 从mempool中选择交易并计算总矿工费
        txs_to_mine = list(self.mempool.values())
        total_fees = 0
        for tx in txs_to_mine:
            total_in = sum(self.utxo_pool[f"{v['txid']}_{v['vout']}"]['amount'] for v in tx['vin'])
            total_out = sum(vout['amount'] for vout in tx['vout'])
            total_fees += total_in - total_out

        # 2. 创建Coinbase交易
        last_block_header = self.blockchain[-1]['header']
        coinbase_tx_data = {
            "vin": [{"txid": "COINBASE", "vout": -1, "scriptSig": f"Mined by {self.id}"}],
            "vout": [{"to_address": self.address, "amount": BLOCK_REWARD + total_fees}]
        }
        coinbase_tx = {**coinbase_tx_data, 'id': calculate_tx_id(coinbase_tx_data)}
        
        # 3. 组装候选区块
        block_transactions = [coinbase_tx] + txs_to_mine
        candidate_header = {
            'index': last_block_header['index'] + 1,
            'previous_hash': last_block_header['hash'],
            'timestamp': time(),
            'nonce': 0
        }

        # 4. 工作量证明 (挖矿)
        print(f"[节点 {self.id}] 开始挖矿... 目标: '{DIFFICULTY_PREFIX}'。打包 {len(txs_to_mine)} 笔交易，总矿工费: {total_fees}。")
        nonce = 0
        while True:
            candidate_header['nonce'] = nonce
            block_hash = calculate_block_hash(candidate_header)
            if block_hash.startswith(DIFFICULTY_PREFIX):
                print(f"  - [节点 {self.id}] 挖矿成功! Nonce: {nonce}")
                final_header = {**candidate_header, 'hash': block_hash}
                final_block = {'header': final_header, 'transactions': block_transactions}
                
                self.blockchain.append(final_block)
                self._update_utxo_and_mempool(final_block)
                self.network.broadcast_block(self, final_block)
                return
            nonce += 1
            
    def display_state(self):
        """打印节点的当前状态"""
        print(f"--- 节点 {self.id} 状态 (地址: {self.address[:10]}...) ---")
        print(f"  - 区块链高度: {len(self.blockchain)}")
        print(f"  - Mempool中的交易数: {len(self.mempool)}")
        print(f"  - UTXO池中的条目数: {len(self.utxo_pool)}")
        my_balance = sum(u['amount'] for u in self.utxo_pool.values() if u['to_address'] == self.address)
        print(f"  - 节点余额: {my_balance} 聪")
        print("-" * (20 + len(self.id)))

# --- 主仿真流程 ---

# 1. 初始化网络和节点
print("=== 初始化网络和节点 ===")
network = Network()
alice_node = Node("Alice")
bob_node = Node("Bob")
miner_node = Node("Miner")

network.add_node(alice_node)
network.add_node(bob_node)
network.add_node(miner_node)
print("\n" + "="*50 + "\n")

# 2. 创建并分发创世区块
print("=== 创建并分发创世区块 ===")
genesis_coinbase_data = {
    "vin": [{"txid": "COINBASE", "vout": -1, "scriptSig": "Genesis"}],
    "vout": [{"to_address": alice_node.address, "amount": 100}] # 创世时给Alice一些钱
}
genesis_coinbase = {**genesis_coinbase_data, 'id': calculate_tx_id(genesis_coinbase_data)}

genesis_header_data = {
    'index': 0,
    'previous_hash': '0'*64,
    'timestamp': 0,
    'nonce': 0
}
genesis_header = {**genesis_header_data, 'hash': calculate_block_hash(genesis_header_data)}
genesis_block = {'header': genesis_header, 'transactions': [genesis_coinbase]}

for node in network.nodes:
    node.add_genesis_block(copy.deepcopy(genesis_block))

print("--- 初始状态 ---")
for node in network.nodes:
    node.display_state()
print("\n" + "="*50 + "\n")


# 3. 第1幕: Alice创建、签名并广播交易
print("=== 第1幕: Alice向Bob转账(60聪)，支付矿工费(5聪) ===")
# a) Alice创建交易草稿
tx_draft = alice_node.create_transaction(bob_node.address, amount=60, fee=5)
if tx_draft:
    # b) Alice签名交易
    alice_signed_tx = alice_node.sign_transaction(tx_draft)
    print(f"[节点 Alice] 已创建并签名交易 {alice_signed_tx['id'][:10]}...")
    
    # c) Alice将交易放入自己的mempool并广播
    alice_node.mempool[alice_signed_tx['id']] = alice_signed_tx
    alice_node.network.broadcast_transaction(alice_node, alice_signed_tx)

print("\n--- 广播后，挖矿前 ---")
for node in network.nodes:
    node.display_state()
print("\n" + "="*50 + "\n")


# 4. 第2幕: 矿工挖矿并广播区块
print("=== 第2幕: 矿工打包交易并挖出新区块 ===")
miner_node.mine_block()


# 5. 第3幕: 最终共识
print("\n=== 第3幕: 最终共识达成 ===")
print("所有节点都接收并验证了新区块，更新了各自的账本。")
print("Alice的交易从Mempool中被清除，Bob收到了钱，矿工获得了奖励+矿工费。")
for node in network.nodes:
    node.display_state()

print("\n--- 最终UTXO池对比 ---")
print("Alice UTXO Pool:", json.dumps(alice_node.utxo_pool, indent=2, sort_keys=True))
print("Bob UTXO Pool:  ", json.dumps(bob_node.utxo_pool, indent=2, sort_keys=True))
print("Miner UTXO Pool:", json.dumps(miner_node.utxo_pool, indent=2, sort_keys=True))
print("\n观察：所有诚实节点的UTXO池在经历了一轮交易和挖矿后，再次达到完全一致的状态。共识达成！")
