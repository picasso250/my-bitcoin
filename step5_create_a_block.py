# step5_create_a_block.py
#
# 教学目标：构建一个包含“已签名”交易的安全区块 (重构版)。
#
# 核心重构：我们不再手动处理每一笔交易的签名细节，
# 而是直接调用 `bitcoin_lib.py` 中的 `create_signed_simple_transaction` 函数。
# 这使得本文件的代码更聚焦于“区块构建”本身，提升了代码的可读性和可维护性。

import hashlib
import json
from time import time
import ecdsa
# 从我们的工具箱导入两个函数
from bitcoin_lib import generate_address, create_signed_simple_transaction

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


# --- 2. 使用函数库创建并签名多笔交易 ---
#
# 核心重构点：现在我们只需一行代码就能创建一笔完整的、已签名的交易。
# 所有复杂的签名逻辑都已封装在 `create_signed_simple_transaction` 函数中。
signed_tx1 = create_signed_simple_transaction(alice_private_key, bob_address, 10)
signed_tx2 = create_signed_simple_transaction(charlie_private_key, alice_address, 5)

transactions_to_pack = [signed_tx1, signed_tx2]

print("--- 1. 已签名的、可无歧义验证的交易包 (由函数库生成) ---")
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
