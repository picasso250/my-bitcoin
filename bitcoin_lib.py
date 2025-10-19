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
import json

def create_signed_simple_transaction(sender_private_key, recipient_address, amount_satoshi):
    """
    创建并签名一个简化的、非UTXO模型的交易包。

    这个函数封装了 Step 4 和 Step 5 中的核心逻辑：
    1. 从发送方私钥派生出公钥和地址。
    2. 构建交易字典。
    3. 对交易进行确定性序列化和哈希。
    4. 使用私钥对哈希进行签名。
    5. 将所有必要信息打包成一个可验证的字典。

    Args:
        sender_private_key: ecdsa.SigningKey 对象，交易发送方的私钥。
        recipient_address (str): 接收方地址的十六进制字符串。
        amount_satoshi (int): 交易金额，以最小单位“聪”表示。

    Returns:
        dict: 一个包含交易负载、公钥和签名的字典，可以直接放入区块。
    """
    # 1. 从私钥派生公钥和地址
    sender_public_key = sender_private_key.get_verifying_key()
    sender_address = generate_address(sender_public_key.to_string("compressed"))

    # 2. 构建交易数据
    tx_data = {
        "from": sender_address,
        "to": recipient_address,
        "amount": amount_satoshi
    }

    # 3. 确定性序列化与哈希
    tx_payload = json.dumps(tx_data, sort_keys=True, separators=(',', ':'))
    tx_hash = hashlib.sha256(tx_payload.encode('utf-8')).digest()

    # 4. 签名
    signature = sender_private_key.sign(tx_hash)

    # 5. 打包成最终的交易结构
    signed_tx = {
        "transaction_payload": tx_payload,
        "public_key": sender_public_key.to_string('compressed').hex(),
        "signature": signature.hex()
    }

    return signed_tx
from time import time

def create_candidate_block(transactions, previous_hash='0'*64):
    """
    组装一个“候选区块”。

    这个函数封装了区块创建的通用逻辑，使其与挖矿过程解耦。
    它接收所有需要打包的数据，并返回一个结构化的、准备好被“挖矿”的区块。

    Args:
        transactions (list): 一个包含已签名交易字典的列表。
        previous_hash (str, optional): 前一个区块的哈希值. 默认为创世区块的哈希.

    Returns:
        dict: 一个包含 'timestamp', 'transactions', 'previous_hash', 和 'nonce' 的候选区块字典.
    """
    block = {
        'timestamp': time(),
        'transactions': transactions,
        'previous_hash': previous_hash,
        'nonce': 0  # Nonce总是从0开始尝试
    }
    return block