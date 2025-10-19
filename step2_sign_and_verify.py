# step2_sign_and_verify.py
#
# 教学目标：理解数字签名的过程，这是比特币交易合法性的基石。
#
# 想象一下你在写一张支票。
# - 支票上的内容（金额、收款人）就是 "交易数据"。
# - 你的亲笔签名就是 "数字签名"。
#
# 任何人都可以看到支票的内容，但只有你的签名能证明这张支票是你开的。
# 在比特币中，过程类似：
# 1. 你用你的私钥对一笔交易进行签名。
# 2. 网络中的其他人用你的公钥来验证这个签名是否真实有效。
#
# 前置准备:
# pip install ecdsa

import ecdsa
import hashlib

# --- 准备工作：生成密钥对 (同第一步) ---
curve = ecdsa.SECP256k1
private_key = ecdsa.SigningKey.generate(curve=curve)
public_key = private_key.get_verifying_key()

# --- 场景模拟 ---
# 假设我们要发起一笔交易，其核心内容可以简化为一个字符串。
# 在真实的比特币中，这里会是更复杂的结构化数据。
transaction_data = "Alice sends 1 BTC to Bob"
print(f"原始交易数据: '{transaction_data}'")

# --- 签名过程 (由私钥持有者完成) ---

# 1. 对交易数据进行哈希运算
#    我们不是对原始数据签名，而是对其哈希值进行签名。
#    这样做效率更高，且能保证数据完整性。比特币使用双SHA256哈希。
#    为简化教学，我们这里用一次SHA256。
#    注意：必须先将字符串编码为字节（bytes）。
message_bytes = transaction_data.encode('utf-8')
message_hash = hashlib.sha256(message_bytes).digest()

print(f"\n[1] 交易数据哈希 (SHA256): {message_hash.hex()}")

# 2. 使用私钥签名哈希值
#    这个动作只有私钥的拥有者才能完成。
#    签名本身是一段看起来随机的数据，但它与私钥和交易数据精确关联。
signature = private_key.sign(message_hash)

print(f"\n[2] 生成的数字签名: {signature.hex()}")

# --- 验签过程 (由网络中的任何节点完成) ---

# 现在，假设一个网络节点收到了三样东西：
# 1. 原始交易数据 (transaction_data)
# 2. 数字签名 (signature)
# 3. 发起人的公钥 (public_key)

# 节点需要独立验证这笔交易是否真的由公钥对应的人发起的。

print("\n--- 网络节点开始验证 ---")

try:
    # 3. 节点重复哈希过程
    #    节点用收到的原始数据，以完全相同的方式计算出哈希值。
    #    如果原始数据在传输中被篡改，这里的哈希值就会不同，验证将失败。
    received_message_bytes = transaction_data.encode('utf-8')
    received_message_hash = hashlib.sha256(received_message_bytes).digest()

    # 4. 使用公钥验证签名
    #    `verify` 函数会用公钥、签名和哈希值进行数学运算。
    #    如果签名确实是由对应的私钥针对该哈希值生成的，函数会成功返回True。
    #    否则，它会抛出一个 ecdsa.BadSignatureError 异常。
    is_valid = public_key.verify(signature, received_message_hash)

    if is_valid:
        print("[3] 验证成功! 👍")
        print("    结论：签名有效，这笔交易确实由公钥所有者发起。")

except ecdsa.BadSignatureError:
    print("[3] 验证失败! 💀")
    print("    结论：签名无效！交易可能是伪造的，或数据已被篡改。")
except Exception as e:
    print(f"发生错误: {e}")


print("\n核心概念:")
print("-> 签名是使用【私钥】对【交易数据哈希】进行的操作。")
print("-> 验证是使用【公钥】对【签名】和【交易数据哈希】进行的操作。")
print("-> 这个过程确保了只有你才能花费你的比特币，且交易内容无法被篡改。")