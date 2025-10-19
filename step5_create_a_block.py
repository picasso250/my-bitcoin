# step5_create_a_block.py
#
# 教学目标：构建一个包含“已签名”交易的安全区块 (修正版)。
#
# 核心修正：根据用户反馈，交易包中存储的不再是原始的交易字典，
# 而是“被签名的那个确定性JSON字符串本身”。
#
# 为什么这是一个关键修正？
# 1. 消除模糊性：验证者直接对包内的字符串哈希，无需猜测序列化方法，保证了验证的唯一性。
# 2. 忠于密码学：签名的对象和验证的对象必须是字节级别上完全相同的数据。

import hashlib
import json
from time import time
import ecdsa
from bitcoin_lib import generate_address

# --- 1. 准备工作: 创建多个参与方 ---
curve = ecdsa.SECP256k1

# Alice
alice_private_key = ecdsa.SigningKey.generate(curve=curve)
alice_public_key = alice_private_key.get_verifying_key()
alice_address = generate_address(alice_public_key.to_string("compressed"))

# Bob
bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_public_key = bob_private_key.get_verifying_key()
bob_address = generate_address(bob_public_key.to_string("compressed"))

# Charlie
charlie_private_key = ecdsa.SigningKey.generate(curve=curve)
charlie_public_key = charlie_private_key.get_verifying_key()
charlie_address = generate_address(charlie_public_key.to_string("compressed"))


# --- 2. 创建并签名多笔交易 ---
# 交易包现在包含三部分：被签名的负载(字符串)、签名、公钥。

# 交易 1: Alice -> Bob
tx1_data = {"from": alice_address, "to": bob_address, "amount": 10}
# 这是将被签名的“官方”数据，是一个字符串
tx1_payload = json.dumps(tx1_data, sort_keys=True, separators=(',', ':'))
tx1_hash = hashlib.sha256(tx1_payload.encode('utf-8')).digest()
tx1_signature = alice_private_key.sign(tx1_hash)

signed_tx1 = {
    "transaction_payload": tx1_payload, # 存储字符串，而非字典！
    "public_key": alice_public_key.to_string('compressed').hex(),
    "signature": tx1_signature.hex()
}

# 交易 2: Charlie -> Alice
tx2_data = {"from": charlie_address, "to": alice_address, "amount": 5}
# 这是将被签名的“官方”数据，是一个字符串
tx2_payload = json.dumps(tx2_data, sort_keys=True, separators=(',', ':'))
tx2_hash = hashlib.sha256(tx2_payload.encode('utf-8')).digest()
tx2_signature = charlie_private_key.sign(tx2_hash)

signed_tx2 = {
    "transaction_payload": tx2_payload, # 存储字符串，而非字典！
    "public_key": charlie_public_key.to_string('compressed').hex(),
    "signature": tx2_signature.hex()
}

transactions_to_pack = [signed_tx1, signed_tx2]

print("--- 1. 已签名的、可无歧义验证的交易包 ---")
print(json.dumps(transactions_to_pack, indent=2), "\n")


# --- 3. 定义并构建安全区块 ---
block = {
    'timestamp': time(),
    'transactions': transactions_to_pack,
    'previous_hash': '0000000000000000000000000000000000000000000000000000000000000000',
    'nonce': 0
}

print("--- 2. 构建的包含安全交易的区块 ---")
print(json.dumps(block, indent=2), "\n")


# --- 4. 计算区块的哈希值 ---
block_string = json.dumps(block, sort_keys=True, separators=(',', ':'))
block_hash = hashlib.sha256(block_string.encode('utf-8')).hexdigest()

print("--- 3. 计算出的区块哈希 ---")
print(f"区块哈希 (SHA256): {block_hash}\n")

print("="*40)
print("     核心架构原则")
print("="*40)
print("-> 区块不关心'交易负载'的内部结构，只把它当作一个需要验证的字符串。")
print("-> 验证确定性：通过存储被签名的原始字符串，彻底消除了验证过程中的任何模糊性。")