# step1_generate_keys.py
#
# 教学目标：演示如何生成比特币兼容的密钥对。
#
# 在比特币系统中，你的“账户”实际上就是一个由私钥和公钥组成的密钥对。
# - 私钥 (Private Key): 一个绝对机密的数字，你有权用它来花费你的比特币。
# - 公钥 (Public Key): 由私钥通过椭圆曲线算法推算得出，可以公开。地址是公钥的另一种表现形式。
#
# 我们将使用 `ecdsa` 这个库来处理椭圆曲线数字签名算法，
# 它是比特币安全的核心。
#
# 前置准备:
# pip install ecdsa

import ecdsa
import hashlib
import os

# 1. 选择比特币使用的椭圆曲线 (secp256k1)
# SECP256k1 是比特币协议选定的特定曲线，所有比特币密钥都必须在这条曲线上生成。
curve = ecdsa.SECP256k1

# 2. 生成私钥
# 私钥本质上是一个非常大的随机数。
# SigningKey.generate() 会创建一个安全的、密码学意义上随机的私钥。
# 我们可以将其看作是“从一堆沙子中捡起特定一粒”的难度。
private_key = ecdsa.SigningKey.generate(curve=curve)

# 3. 从私钥派生公钥
# 这个过程是单向的，从私钥可以轻松计算出公钥，但反之则不可能。
# 这就是为什么你可以安全地把公钥（或地址）告诉别人。
public_key = private_key.get_verifying_key()

# 4. 打印密钥
# 为了方便查看，我们将密钥的二进制格式转换为十六进制字符串。
# .to_string() 返回的是原始的字节（bytes），.hex() 将其转换为可读的十六进制。
private_key_hex = private_key.to_string().hex()
public_key_hex = public_key.to_string("compressed").hex() # 通常使用压缩公钥格式

print("--- 教学版比特币密钥生成 ---")
print(f"所用曲线: {curve.name}")
print("\n[1] 生成的私钥:")
print(f"  - 十六进制: {private_key_hex}")
print(f"  - 长度: {len(private_key.to_string())} 字节 (256 位)")

print("\n[2] 派生的公钥 (压缩格式):")
print(f"  - 十六进制: {public_key_hex}")
print(f"  - 长度: {len(public_key.to_string('compressed'))} 字节")

print("\n核心概念:")
print("-> 私钥是你的核心秘密，绝不能泄露。")
print("-> 公钥由私钥单向派生，用于验证你的签名。")
print("-> 下一步，我们将学习如何用私钥对一笔“交易”进行签名。")