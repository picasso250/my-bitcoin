# step3_public_key_to_address.py
#
# 教学目标：理解比特币地址的生成过程，揭示其背后的安全与健壮性设计。
#
# 地址不是随机生成的，而是从公钥经过一系列确定性的哈希和格式化步骤派生出来的。
# 这个过程我们刚才已详细讨论过，现在来亲手实现它。
#
# 前置准备:
# pip install ecdsa

import ecdsa
import hashlib

# --- 准备工作：生成密钥对，获取公钥 ---
curve = ecdsa.SECP256k1
private_key = ecdsa.SigningKey.generate(curve=curve)
public_key = private_key.get_verifying_key()

# 在比特币中，通常使用压缩公钥（以02或03开头）来生成地址，可以节省空间。
public_key_bytes = public_key.to_string("compressed")
print(f"--- 步骤 0: 准备压缩公钥 ---\n{public_key_bytes.hex()}\n")


# --- 步骤 1: 对公钥进行双哈希 (SHA-256 then RIPEMD-160) ---
# 这是为了“隐藏”公钥，防范未来可能的量子计算攻击。
sha256_hash = hashlib.sha256(public_key_bytes).digest()
print(f"--- 步骤 1a: SHA-256 哈希 ---\n{sha256_hash.hex()}\n")

# RIPEMD-160 是一个不同的哈希算法，提供“纵深防御”。
# 注意：Python标准库没有RIPEMD-160，但hashlib通常可以通过OpenSSL支持它。
ripemd160 = hashlib.new('ripemd160')
ripemd160.update(sha256_hash)
public_key_hash = ripemd160.digest()
print(f"--- 步骤 1b: RIPEMD-160 哈希 (得到公钥哈希) ---\n{public_key_hash.hex()}\n")


# --- 步骤 2: 添加版本字节 ---
# 版本字节 `0x00` 代表这是主网的P2PKH地址。
# 它像文件的扩展名，告诉网络如何处理这个地址。
version_byte = b'\x00'
versioned_hash = version_byte + public_key_hash
print(f"--- 步骤 2: 添加版本字节 `00` ---\n{versioned_hash.hex()}\n")


# --- 步骤 3: 计算并添加校验和 ---
# 校验和用于防止输入错误，是用户体验和资金安全的“安全带”。
# 它是对“版本化哈希”进行双SHA-256运算后的前4个字节。
checksum_hash_1 = hashlib.sha256(versioned_hash).digest()
checksum_hash_2 = hashlib.sha256(checksum_hash_1).digest()
checksum = checksum_hash_2[:4]
print(f"--- 步骤 3a: 计算校验和 (双SHA256的前4字节) ---\n{checksum.hex()}\n")

# 将校验和附加到版本化哈希的末尾
final_address_bytes = versioned_hash + checksum
print(f"--- 步骤 3b: 拼接成最终地址字节 ---\n{final_address_bytes.hex()}\n")


# --- 步骤 4: 编码 (教学版简化) ---
# 传统的比特币地址会进行Base58Check编码，使其更易于人工抄写。
# 根据我们的讨论，这是一个为罕见场景优化的复杂步骤。
# 为了聚焦核心，我们将其简化，直接使用十六进制表示。
# 这种格式在编程和调试时更清晰。
simplified_address_hex = final_address_bytes.hex()

print("="*40)
print("       最终教学版地址 (十六进制)")
print("="*40)
print(simplified_address_hex)
print("\n核心概念:")
print("-> 地址是公钥经过 SHA256 -> RIPEMD160 -> 加版本号 -> 加校验和 的产物。")
print("-> 每一步都有其明确的设计目的：安全、健壮或扩展性。")
print("-> 我们省略了Base58编码，以十六进制展示，更聚焦于协议核心。")