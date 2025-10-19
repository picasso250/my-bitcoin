# step6_mine_a_block.py
#
# 教学目标：引入“挖矿”和“区块奖励” (重构版)。
#
# 核心重构点：
# 1. 矿工现在有了自己的身份（地址）。
# 2. 我们在挖矿前，手动创建一笔特殊的 "Coinbase 交易"。
#    这笔交易将区块奖励（一笔凭空产生的币）发送给矿工。
# 3. Coinbase 交易必须是区块中交易列表的第一个元素。

import hashlib
import json
from time import time
import ecdsa
# 从我们的工具箱导入三个核心函数
from bitcoin_lib import generate_address, create_signed_simple_transaction, create_candidate_block

# --- 1. 定义常量 ---
DIFFICULTY_PREFIX = '0000'
BLOCK_REWARD = 50 # 设定挖出一个区块的奖励为50个币


def mine_block(block):
    """
    接收一个候选区块，通过不断尝试Nonce来寻找满足难度目标的哈希值。
    (此函数与上一版完全相同)
    """
    print("矿工开始工作... 目标: 找到一个以 '{}' 开头的哈希。".format(DIFFICULTY_PREFIX))
    
    nonce = 0
    while True:
        block['nonce'] = nonce
        block_string = json.dumps(block, sort_keys=True, separators=(',', ':'))
        block_hash = hashlib.sha256(block_string.encode('utf-8')).hexdigest()

        if nonce % 10000 == 0:
            print(f"  - 尝试 Nonce: {nonce}, 哈希: {block_hash[:10]}...")

        if block_hash.startswith(DIFFICULTY_PREFIX):
            print(f"\n🎉 挖矿成功! 🎉")
            print(f"  - 最终 Nonce: {nonce}")
            return block, block_hash
        
        nonce += 1

# --- 2. 准备工作: 创建多个身份 ---
curve = ecdsa.SECP256k1

# 普通用户
alice_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_address = generate_address(bob_private_key.get_verifying_key().to_string("compressed"))

# 矿工
miner_private_key = ecdsa.SigningKey.generate(curve=curve)
miner_public_key = miner_private_key.get_verifying_key()
miner_address = generate_address(miner_public_key.to_string('compressed'))
print(f"--- 准备工作 ---")
print(f"矿工地址: {miner_address}\n")


# --- 3. 创建普通交易和Coinbase交易 ---

# 一笔由Alice发给Bob的普通交易
user_transaction = create_signed_simple_transaction(alice_private_key, bob_address, 100)

# 一笔特殊的Coinbase交易，用于奖励矿工
# 它没有 "from"，因为它代表了新创造的币
# 它没有签名，因为它的合法性由PoW共识来保障
coinbase_transaction = {
    "from": "COINBASE",
    "to": miner_address,
    "amount": BLOCK_REWARD
}

# 将Coinbase交易放在列表的最前面，这是比特币协议的规则
transactions_for_block = [coinbase_transaction, user_transaction]


# --- 4. 调用函数库组装“候选区块” ---
candidate_block = create_candidate_block(
    transactions_for_block,
    '0000000000000000000000000000000000000000000000000000000000000000'
)

print("--- 1. 组装好的候选区块 (包含Coinbase交易) ---")
print(json.dumps(candidate_block, indent=2), "\n")


# --- 5. 开始挖矿 ---
final_block, final_hash = mine_block(candidate_block)


# --- 6. 结果展示 ---
print("\n--- 2. 挖矿完成后的最终区块 ---")
print(json.dumps(final_block, indent=2), "\n")

print("--- 3. 最终区块的有效哈希 ---")
print(f"区块哈希: {final_hash}\n")

print("核心概念:")
print("-> Coinbase交易: 每个区块的第一笔特殊交易，用于将固定的区块奖励和交易费发给矿工。这是矿工提供算力保护网络的核心激励。")