# step4_build_real_transaction.py
#
# 教学目标：将所有部件组装起来，创建、签名并验证一笔模拟的比特币交易。
# 这是对比特币运作模式最核心的模拟。
#
# 前置准备:
# pip install ecdsa

import ecdsa
import hashlib
import json

# --- 辅助函数: 从公钥生成地址 (封装Step 3的逻辑) ---
# 为了代码整洁，我们将第三步的逻辑封装成一个可重用的函数。
def generate_address(public_key_bytes):
    """从压缩公钥字节生成简化的十六进制地址。"""
    # 1. 双哈希
    sha256_hash = hashlib.sha256(public_key_bytes).digest()
    ripemd160 = hashlib.new('ripemd160')
    ripemd160.update(sha256_hash)
    public_key_hash = ripemd160.digest()
    # 2. 加版本字节
    version_byte = b'\x00'
    versioned_hash = version_byte + public_key_hash
    # 3. 加校验和
    checksum_hash_1 = hashlib.sha256(versioned_hash).digest()
    checksum_hash_2 = hashlib.sha256(checksum_hash_1).digest()
    checksum = checksum_hash_2[:4]
    final_address_bytes = versioned_hash + checksum
    # 4. 编码
    return final_address_bytes.hex()

# --- 1. 场景设置: 创建参与方 ---
# 我们需要一个发送者(Alice)和一个接收者(Bob)。
# 他们各自拥有自己的密钥对和地址。
curve = ecdsa.SECP256k1

# Alice的身份
alice_private_key = ecdsa.SigningKey.generate(curve=curve)
alice_public_key = alice_private_key.get_verifying_key()
alice_address = generate_address(alice_public_key.to_string("compressed"))

# Bob的身份
bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_public_key = bob_private_key.get_verifying_key()
bob_address = generate_address(bob_public_key.to_string("compressed"))

print("--- 1. 参与方身份 ---")
print(f"Alice's Address: {alice_address}")
print(f"Bob's Address:   {bob_address}\n")


# --- 2. 构建交易结构 ---
# 一笔交易的核心是“输入”和“输出”的列表。
# - Input:  声明你要花费哪一笔钱 (引用一个UTXO - 未花费交易输出)。
#           为简化，我们只声明花费的钱来自Alice的地址。
# - Output: 声明钱将流向何处。
transaction = {
    "inputs": [
        {"from_address": alice_address, "amount": 1.0}
    ],
    "outputs": [
        {"to_address": bob_address, "amount": 0.9},
        {"to_address": alice_address, "amount": 0.1} # 找零 (Change)
    ]
}
print("--- 2. 构建的交易 (JSON格式) ---")
print(json.dumps(transaction, indent=2), "\n")


# --- 3. 为签名准备数据 ---
# 我们必须对交易数据进行签名，以证明Alice同意这笔花费。
# 为保证所有节点计算出的哈希完全一致，我们需要对数据进行“确定性序列化”。
# - `sort_keys=True`: 保证字典键的顺序。
# - `separators=(',', ':')`: 去掉所有不必要的空格。
tx_string = json.dumps(transaction, sort_keys=True, separators=(',', ':'))
tx_bytes = tx_string.encode('utf-8')
tx_hash = hashlib.sha256(tx_bytes).digest()

print("--- 3. 待签名的交易哈希 ---")
print(f"序列化字符串: {tx_string}")
print(f"交易哈希: {tx_hash.hex()}\n")


# --- 4. 签名交易 (Alice的操作) ---
# 这是关键一步：只有Alice能用她的私钥完成签名。
# 这个签名与交易数据和她的身份牢固地绑定在一起。
signature = alice_private_key.sign(tx_hash)

print("--- 4. Alice签名交易 ---")
print(f"生成的签名: {signature.hex()}\n")


# --- 5. 验证交易 (网络中任何节点的操作) ---
# 一个节点收到了交易、签名和Alice的公钥。它需要进行验证。
# 验证过程不涉及任何私钥。
print("--- 5. 网络节点验证交易 ---")
try:
    # 节点独立地重复第3步，计算出它认为的交易哈希。
    # 这样可以防止交易数据在传输中被篡改。
    node_tx_hash = hashlib.sha256(tx_bytes).digest()

    # 使用Alice的公钥验证签名是否与交易哈希匹配。
    # 如果`verify`成功，说明签名有效。如果失败，它会抛出异常。
    is_valid = alice_public_key.verify(signature, node_tx_hash)
    print("验证结果: 成功! 👍")
    print("结论: 签名有效，可以确认这笔交易是Alice本人发起的。")

except ecdsa.BadSignatureError:
    print("验证结果: 失败! 💀")
    print("结论: 签名无效！这是一个欺诈性或已损坏的交易。")