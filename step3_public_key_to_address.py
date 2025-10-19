# step3_public_key_to_address.py
#
# 教学目标：实现从公钥到地址的派生，并理解每一步设计的核心意图。
# 我们将采用“AI-Ready”注释风格，只点出关键概念，鼓励您向AI提问以深入探索。
#
# 前置准备:
# pip install ecdsa

import ecdsa
import hashlib

# --- 准备工作: 获取公钥 ---
# 整个流程的起点是一个公钥。
# 我们使用“压缩公钥”格式，这是比特币标准，能节省空间。
curve = ecdsa.SECP256k1
private_key = ecdsa.SigningKey.generate(curve=curve)
public_key_bytes = private_key.get_verifying_key().to_string("compressed")

print(f"--- 0. 起点: 压缩公钥 ---\n{public_key_bytes.hex()}\n")


# --- 步骤 1: 双哈希 - SHA256 -> RIPEMD160 ---
# 目的: 增强安全性。
# - 哈希公钥可以防御潜在的“量子计算攻击”。
# - 使用两种不同的哈希算法是一种“纵深防御 (Defense in Depth)”策略。
# - RIPEMD160也能将哈希长度缩短至20字节，使地址更短。
sha256_hash = hashlib.sha256(public_key_bytes).digest()
ripemd160 = hashlib.new('ripemd160')
ripemd160.update(sha256_hash)
public_key_hash = ripemd160.digest()

print(f"--- 1. 公钥哈希 ---\n{public_key_hash.hex()}\n")


# --- 步骤 2: 添加版本字节 ---
# 目的: 赋予地址“类型”和“功能”。
# - `0x00` 代表主网的 P2PKH 地址 (Pay-to-Public-Key-Hash)。
#
# 重要概念: 协议中的数据是字节流 (Bytes)，而非文本 (String)。
# `b'\x00' + ...` 是在字节序列前添加一个值为0的字节。
# 这才是数据在底层的真实结构，后续的哈希计算也必须基于这个字节序列。
version_byte = b'\x00'
versioned_hash = version_byte + public_key_hash

print(f"--- 2. 添加版本字节 `00` ---\n{versioned_hash.hex()}\n")


# --- 步骤 3: 计算并拼接校验和 ---
# 目的: 数据完整性检查 (Data Integrity Check)，防止用户输入错误。
# 这不是为了加密安全，而是为了防止资金因地址抄写错误而永久丢失。
checksum_hash_1 = hashlib.sha256(versioned_hash).digest()
checksum_hash_2 = hashlib.sha256(checksum_hash_1).digest()
checksum = checksum_hash_2[:4] # 取双哈希结果的前4个字节

final_address_bytes = versioned_hash + checksum

print(f"--- 3. 拼接4字节校验和 ---\n{checksum.hex()} (校验和部分)\n{final_address_bytes.hex()} (完整地址字节)\n")


# --- 步骤 4: 编码 (教学版简化) ---
# 传统比特币使用 Base58Check 编码，优化“手动抄写”场景。
# 我们跳过这一步，直接使用十六进制，因为它更适合编程和调试，
# 能让你清晰地看到 [版本字节 | 公钥哈希 | 校验和] 的数据结构。
simplified_address_hex = final_address_bytes.hex()

print("="*40)
print("     最终教学版地址 (十六进制表示)")
print("="*40)
print(simplified_address_hex)
print("\n[版本: 1字节 | 公钥哈希: 20字节 | 校验和: 4字节]")