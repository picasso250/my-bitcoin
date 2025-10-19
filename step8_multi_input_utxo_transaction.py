# step8_multi_input_utxo_transaction.py
#
# 教学目标：实现一个完整的、更真实的UTXO交易——多输入，多输出。
#
# 核心洞见：
# 1. 钱包的真相：用户的“钱”不是存在一个地方，而是分散在多个由不同私钥控制的UTXO里。
#    钱包软件的核心工作之一就是管理这些“钥匙”。
# 2. 凑钱支付：当单个UTXO金额不足时，必须将多个UTXO作为一笔交易的输入，
#    这种“凑钱”能力是比特币实用性的基础。
# 3. 多重授权：每一个被花费的UTXO，都必须由其对应的唯一私钥进行签名授权。
#    一笔交易可以包含多个来源不同、签名不同的输入。

import ecdsa
import hashlib
import json
from bitcoin_lib import generate_address

# --- 准备工作: 函数与场景设置 ---

def create_and_sign_multi_input_transaction(transaction_draft, private_keys):
    """
    接收一个构建好的、含多个输入的交易草稿和对应的私钥列表，完成所有签名。

    Args:
        transaction_draft (dict): 包含vin和vout的交易草稿。
        private_keys (list): ecdsa.SigningKey对象的列表，顺序必须与vin中的输入一一对应。
    
    Returns:
        (dict, str): 包含所有签名的完整交易，以及该交易的ID。
    """
    # 1. 确定性序列化与哈希
    # 这是所有签名共享的同一个“合同”
    tx_payload = json.dumps(transaction_draft, sort_keys=True, separators=(',', ':'))
    tx_hash_to_sign = hashlib.sha256(tx_payload.encode('utf-8')).digest()

    # 2. 循环遍历所有输入，并用对应的私钥签名
    for i, vin_item in enumerate(transaction_draft['vin']):
        private_key = private_keys[i]
        public_key = private_key.get_verifying_key()
        
        signature = private_key.sign(tx_hash_to_sign)
        
        # 将签名和公钥填充到对应的输入中
        vin_item['scriptSig'] = {
            'signature': signature.hex(),
            'publicKey': public_key.to_string("compressed").hex()
        }
        
    # 3. 计算最终的交易ID (对包含签名的完整交易进行双哈希)
    final_tx_payload = json.dumps(transaction_draft, sort_keys=True, separators=(',', ':'))
    txid = hashlib.sha256(hashlib.sha256(final_tx_payload.encode('utf-8')).digest()).hexdigest()

    return transaction_draft, txid


# --- 场景设置 ---
curve = ecdsa.SECP256k1

# Alice拥有一个“钱包”，里面有两把不同的钥匙（私钥）
alice_private_key_1 = ecdsa.SigningKey.generate(curve=curve)
alice_public_key_1 = alice_private_key_1.get_verifying_key()
alice_address_1 = generate_address(alice_public_key_1.to_string("compressed"))

alice_private_key_2 = ecdsa.SigningKey.generate(curve=curve)
alice_public_key_2 = alice_private_key_2.get_verifying_key()
alice_address_2 = generate_address(alice_public_key_2.to_string("compressed"))

# Bob是收款方
bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_public_key = bob_private_key.get_verifying_key()
bob_address = generate_address(bob_public_key.to_string("compressed"))


# --- 1. 模拟全局UTXO池 ---
# Alice的“零钱”分散在由不同地址锁定的UTXO中
utxo_pool = {}

# 第一个UTXO，由alice_address_1锁定
txid_1 = '0'*63 + 'a'
utxo_key_1 = f"{txid_1}_0"
utxo_pool[utxo_key_1] = { 'address': alice_address_1, 'amount': 50 }

# 第二个UTXO，由alice_address_2锁定
txid_2 = '0'*63 + 'b'
utxo_key_2 = f"{txid_2}_1" # 假设这是另一笔交易的第二个输出
utxo_pool[utxo_key_2] = { 'address': alice_address_2, 'amount': 80 }

print("--- 1. 交易前的UTXO池 (Alice拥有两笔零钱) ---")
print(json.dumps(utxo_pool, indent=2), "\n")


# --- 2. Alice手动构建“凑钱”交易 ---
# 目标：Alice想给Bob支付120聪。单个UTXO金额不足，必须同时使用两个。
total_input_amount = 50 + 80

# a) 构建输入 (vin)
#    必须严格按照顺序，这决定了之后签名的顺序
vin = [
    { 'txid': txid_1, 'vout': 0, 'scriptSig': None },
    { 'txid': txid_2, 'vout': 1, 'scriptSig': None }
]

# b) 构建输出 (vout)
#    总输入130，支付120，找零10
vout = [
    { 'to_address': bob_address, 'amount': 120 },
    { 'to_address': alice_address_1, 'amount': 10 } # 找零可以回到任意一个自己的地址
]

# c) 组装交易草稿
transaction_draft = { 'vin': vin, 'vout': vout }

print("--- 2. Alice构建的多输入交易草稿 (尚未签名) ---")
print(json.dumps(transaction_draft, indent=2), "\n")


# --- 3. 对交易进行多重签名 ---
# 关键：私钥列表的顺序必须和vin列表的顺序严格对应
keys_for_signing = [alice_private_key_1, alice_private_key_2]
signed_transaction, new_txid = create_and_sign_multi_input_transaction(transaction_draft, keys_for_signing)

print("--- 3. 包含多个签名的完整交易 ---")
print(f"新交易ID (TXID): {new_txid}")
print(json.dumps(signed_transaction, indent=2), "\n")


# --- 4. 模拟交易被确认，更新UTXO池 ---
# a) 销毁两个被花费的UTXO
del utxo_pool[utxo_key_1]
del utxo_pool[utxo_key_2]

# b) 创建两个新的UTXO
utxo_pool[f"{new_txid}_0"] = signed_transaction['vout'][0]
utxo_pool[f"{new_txid}_1"] = signed_transaction['vout'][1]

print("--- 4. 交易后的UTXO池 ---")
print(json.dumps(utxo_pool, indent=2), "\n")

print("核心概念:")
print("-> 凑钱支付: 多个独立的UTXO被合并为一笔交易的输入，以满足支付需求。")
print("-> 多重授权: 每个输入都由其对应的私钥独立签名，共同授权了同一笔交易。")
print("-> UTXO池的演变: 两个旧的UTXO被销毁，两个新的UTXO被创造，总金额保持守恒（暂不考虑矿工费）。")