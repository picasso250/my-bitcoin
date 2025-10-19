# bitcoin_lib.py
#
# 我们的教学版比特币项目的“工具箱”。
# 存放可被多个脚本重复使用的核心函数。

import hashlib

def generate_address(public_key_bytes):
    """
    从一个压缩公钥字节生成简化的十六进制P2PKH地址。
    这是 Step 3 的核心逻辑封装。
    """
    # 1. 双哈希 (SHA256 -> RIPEMD160) - 增强安全性
    sha256_hash = hashlib.sha256(public_key_bytes).digest()
    ripemd160 = hashlib.new('ripemd160')
    ripemd160.update(sha256_hash)
    public_key_hash = ripemd160.digest()

    # 2. 添加版本字节 `0x00` - 标识地址类型
    version_byte = b'\x00'
    versioned_hash = version_byte + public_key_hash

    # 3. 计算并拼接4字节校验和 - 防止输入错误
    checksum_hash_1 = hashlib.sha256(versioned_hash).digest()
    checksum_hash_2 = hashlib.sha256(checksum_hash_1).digest()
    checksum = checksum_hash_2[:4]
    final_address_bytes = versioned_hash + checksum

    # 4. 以十六进制返回
    return final_address_bytes.hex()