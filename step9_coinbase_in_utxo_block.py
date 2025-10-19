# step9_coinbase_in_utxo_block.py
#
# 教学目标：完整演示Coinbase交易在UTXO模型中的作用。
#
# 核心洞见：
# 1. 新币发行：Coinbase交易是系统中唯一“凭空”创造价值的机制，它将区块奖励注入UTXO池。
# 2. 激励机制：矿工费（交易输入总额 - 输出总额）是矿工打包普通交易的核心动力。
# 3. 特殊结构：Coinbase交易没有有效的输入(vin)和签名，它的合法性由区块的PoW哈希保障。

import ecdsa
import hashlib
import json
from time import time

# --- 0. 导入并复用我们的工具箱 ---
# 为了让此脚本能独立运行，我们将之前在lib中的函数直接复制过来。
def generate_address(public_key_bytes):
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

def create_candidate_block(transactions, previous_hash='0'*64):
    block = {
        'timestamp': time(),
        'transactions': transactions,
        'previous_hash': previous_hash,
        'nonce': 0
    }
    return block

def mine_block(block, difficulty_prefix):
    print(f"矿工开始工作... 目标: 找到一个以 '{difficulty_prefix}' 开头的哈希。")
    nonce = 0
    while True:
        block['nonce'] = nonce
        # 注意：为了让Coinbase交易的txid稳定，我们必须对交易列表也进行确定性序列化
        # 但为简化教学，此处我们依然直接序列化整个block字典
        block_string = json.dumps(block, sort_keys=True, separators=(',', ':'))
        block_hash = hashlib.sha256(block_string.encode('utf-8')).hexdigest()

        if nonce % 20000 == 0:
            print(f"  - 尝试 Nonce: {nonce}, 哈希: {block_hash[:10]}...")

        if block_hash.startswith(difficulty_prefix):
            print(f"\n🎉 挖矿成功! 🎉")
            print(f"  - 最终 Nonce: {nonce}")
            return block, block_hash
        nonce += 1

# --- 1. 定义常量和场景设置 ---
DIFFICULTY_PREFIX = '0000'
BLOCK_REWARD = 50

curve = ecdsa.SECP256k1

# 创建矿工、Alice和Bob的身份
miner_private_key = ecdsa.SigningKey.generate(curve=curve)
miner_public_key = miner_private_key.get_verifying_key()
miner_address = generate_address(miner_public_key.to_string('compressed'))

alice_private_key = ecdsa.SigningKey.generate(curve=curve)
alice_public_key = alice_private_key.get_verifying_key()
alice_address = generate_address(alice_public_key.to_string("compressed"))

bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_public_key = bob_private_key.get_verifying_key()
bob_address = generate_address(bob_public_key.to_string("compressed"))

# --- 2. 初始化全局UTXO池 ---
# 假设Alice之前已拥有一个100聪的UTXO
utxo_pool = {}
initial_txid = '0'*63 + 'f'
utxo_to_spend_key = f"{initial_txid}_0"
utxo_pool[utxo_to_spend_key] = {'address': alice_address, 'amount': 100}

print("--- 1. 挖矿前的UTXO池 ---")
print(json.dumps(utxo_pool, indent=2), "\n")


# --- 3. Alice创建一笔带矿工费的交易 ---
# 目标: Alice用100聪的UTXO，给Bob转60聪，找零35聪。
# 100 (输入) - 60 (支付) - 35 (找零) = 5 (矿工费)
vin = [{'txid': initial_txid, 'vout': 0, 'scriptSig': None}]
vout = [
    {'to_address': bob_address, 'amount': 60},
    {'to_address': alice_address, 'amount': 35}
]
alice_tx_draft = {'vin': vin, 'vout': vout}

# 对交易草稿进行签名
tx_payload = json.dumps(alice_tx_draft, sort_keys=True, separators=(',', ':'))
tx_hash = hashlib.sha256(tx_payload.encode('utf-8')).digest()
signature = alice_private_key.sign(tx_hash)
alice_tx_draft['vin'][0]['scriptSig'] = {
    'signature': signature.hex(),
    'publicKey': alice_public_key.to_string("compressed").hex()
}
alice_signed_tx = alice_tx_draft
alice_tx_id = hashlib.sha256(hashlib.sha256(json.dumps(alice_signed_tx, sort_keys=True, separators=(',', ':')).encode('utf-8')).digest()).hexdigest()

print(f"--- 2. Alice创建了一笔交易 (TXID: {alice_tx_id[:10]}...) ---")
print("       - 输入: 100聪")
print("       - 输出: 60聪给Bob, 35聪找零给自己")
print("       - 隐含矿工费: 5聪\n")


# --- 4. 矿工组装候选区块 ---

# a) 计算总矿工费
total_fee = utxo_pool[utxo_to_spend_key]['amount'] - sum(o['amount'] for o in alice_signed_tx['vout'])

# b) 创建Coinbase交易
coinbase_tx = {
    "vin": [{"txid": "COINBASE", "vout": -1, "scriptSig": "Mined by Gemini" }],
    "vout": [{"to_address": miner_address, "amount": BLOCK_REWARD + total_fee}]
}
coinbase_tx_id = hashlib.sha256(hashlib.sha256(json.dumps(coinbase_tx, sort_keys=True, separators=(',', ':')).encode('utf-8')).digest()).hexdigest()

# c) 将Coinbase交易放在首位，组装区块
transactions_for_block = [coinbase_tx, alice_signed_tx]
candidate_block = create_candidate_block(transactions_for_block)

print(f"--- 3. 矿工组装候选区块 ---")
print(f"       - 包含Coinbase交易 (奖励: {BLOCK_REWARD + total_fee}聪)")
print(f"       - 包含Alice的交易\n")


# --- 5. 矿工进行挖矿 ---
final_block, final_hash = mine_block(candidate_block, DIFFICULTY_PREFIX)


# --- 6. 模拟区块被确认，更新UTXO池 ---
print("\n--- 4. 区块被确认，开始更新UTXO池 ---")

# a) 处理Alice的交易：销毁旧UTXO，创建新UTXO
del utxo_pool[utxo_to_spend_key]
print(f"   - [销毁] Alice花费的UTXO: {utxo_to_spend_key}")

utxo_pool[f"{alice_tx_id}_0"] = alice_signed_tx['vout'][0]
print(f"   - [创建] 给Bob的新UTXO (60聪)")
utxo_pool[f"{alice_tx_id}_1"] = alice_signed_tx['vout'][1]
print(f"   - [创建] 给Alice的找零UTXO (35聪)")

# b) 处理Coinbase交易：凭空创建新UTXO
utxo_pool[f"{coinbase_tx_id}_0"] = coinbase_tx['vout'][0]
print(f"   - [创建] 给矿工的Coinbase奖励UTXO ({BLOCK_REWARD + total_fee}聪)\n")


# --- 7. 最终结果 ---
print("--- 5. 挖矿后的最终UTXO池 ---")
print(json.dumps(utxo_pool, indent=2), "\n")

print("核心概念:")
print("-> 价值守恒与创造: 普通交易(Alice)遵守UTXO的价值守恒，而Coinbase交易则为系统注入了新的价值。")
print("-> 矿工的双重收益: 矿工通过挖矿获得了固定的区块奖励(50)和用户支付的手续费(5)，总计55个聪。")