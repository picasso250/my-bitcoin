# step4_build_real_transaction.py
#
# 教学目标：演示如何将一个结构化的“转账意图”进行数字签名，确保其不可篡改。
# 这是 Step 2 的升级版：我们签名的不再是简单字符串，而是一个包含地址和金额的JSON对象。
# 注意：本步骤为教学目的进行了简化，尚未引入完整的UTXO模型。
#
# 前置准备:
# pip install ecdsa

import ecdsa
import hashlib
import json

# --- 0. 导入我们的工具箱 ---
# 我们将复用的函数放在了 bitcoin_lib.py 中，使本文件更聚焦于交易本身。
from bitcoin_lib import generate_address


# --- 1. 场景设置: 创建参与方 ---
# Alice (发送者) 和 Bob (接收者)
curve = ecdsa.SECP256k1

alice_private_key = ecdsa.SigningKey.generate(curve=curve)
alice_public_key = alice_private_key.get_verifying_key()
alice_address = generate_address(alice_public_key.to_string("compressed"))

bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_public_key = bob_private_key.get_verifying_key()
bob_address = generate_address(bob_public_key.to_string("compressed"))

print("--- 1. 参与方身份 ---")
print(f"Alice's Address: {alice_address}")
print(f"Bob's Address:   {bob_address}\n")


# --- 2. 构建简化的交易意图 (使用整数!) ---
# 核心教学点：绝不使用浮点数处理金融计算！
# 我们用整数来表示最小单位“聪 (Satoshi)”，确保计算的精确性。
# 1 BTC = 100,000,000 Satoshis.
satoshi_per_btc = 100_000_000

# 这是一个简化的交易结构，清晰地表达了“谁给谁转了多少钱”。
simple_transaction = {
    "from": alice_address,
    "to": bob_address,
    "amount": int(0.9 * satoshi_per_btc)
}
print("--- 2. 构建的简化交易 (单位: 聪) ---")
print(json.dumps(simple_transaction, indent=2), "\n")


# --- 3. 确定性序列化与哈希 ---
# 核心教学点：如何将交易数据转换成唯一的、可供签名的“指纹”。
# 这个过程必须是“确定性的”，即任何人在任何机器上执行，都必须得到完全相同的结果。
# 我们使用 `json.dumps` 并配合 `sort_keys=True` 和 `separators` 来实现。
tx_string_for_signing = json.dumps(simple_transaction, sort_keys=True, separators=(',', ':'))
tx_bytes = tx_string_for_signing.encode('utf-8')
tx_hash = hashlib.sha256(tx_bytes).digest()

print("--- 3. 待签名的交易哈希 ---")
print(f"确定性序列化字符串: {tx_string_for_signing}")
print(f"交易哈希: {tx_hash.hex()}\n")


# --- 4. 签名交易 (Alice的操作) ---
# Alice 使用她的私钥对这个独一无二的交易哈希进行签名。
signature = alice_private_key.sign(tx_hash)

print("--- 4. Alice签名交易 ---")
print(f"生成的签名: {signature.hex()}\n")


# --- 5. 验证交易 (网络节点的操作) ---
# 节点使用 Alice 的公钥来验证签名是否有效。
print("--- 5. 网络节点验证交易 ---")
try:
    # 节点必须使用完全相同的序列化方法来独立计算哈希值。
    node_tx_hash = hashlib.sha256(tx_string_for_signing.encode('utf-8')).digest()

    # 使用公钥验证签名。如果成功，一切正常；如果失败，则抛出异常。
    # 这里的公钥 (alice_public_key) 在真实网络中需要从签名或交易数据中恢复出来，
    # 但为了教学清晰，我们假设节点已经知道了Alice的公钥。
    is_valid = alice_public_key.verify(signature, node_tx_hash)
    print("验证结果: 成功! 👍")
    print("结论: 签名与交易数据匹配，证明这笔转账意图确实由Alice发起且未被篡改。")

except ecdsa.BadSignatureError:
    print("验证结果: 失败! 💀")
    print("结论: 签名无效！")