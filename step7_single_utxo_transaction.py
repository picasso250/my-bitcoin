# step7_single_utxo_transaction.py
#
# 教学目标：引入比特币最核心、也最独特的概念——UTXO (Unspent Transaction Output)。
#
# 核心洞见：
# 1. 比特币没有“账户余额”这个概念。你的“钱”实际上是散落在区块链上、
#    像一张张“待报销发票”一样的UTXO。
# 2. 花钱的本质是：消耗一个或多个你拥有的旧UTXO，并创造出新的UTXO。
# 3. UTXO是不可分割的。就像你不能把一张100元的纸币撕开只用30元，
#    你必须把整张100元都花掉，然后收回70元的找零。
#
# 本脚本将演示这个“单输入，双输出（支付+找零）”的最简UTXO模型。

import ecdsa
import hashlib
import json
from bitcoin_lib import generate_address

# --- 准备工作: 函数与场景设置 ---

def create_and_sign_transaction(transaction_draft, private_key):
    """
    接收一个构建好的交易草稿和私钥，完成签名。
    
    这个函数是“纯粹”的，它不进行任何逻辑判断（如计算找零）。
    它的唯一工作就是对传入的数据进行确定性哈希和签名。
    """
    # 1. 确定性序列化与哈希
    # 为了教学清晰，我们假设vin/vout的顺序是固定的。
    tx_payload = json.dumps(transaction_draft, sort_keys=True, separators=(',', ':'))
    tx_hash = hashlib.sha256(tx_payload.encode('utf-8')).digest()
    
    # 2. 签名
    signature = private_key.sign(tx_hash)
    
    # 3. 将签名和公钥填充回报文
    public_key_hex = private_key.get_verifying_key().to_string("compressed").hex()
    
    # 假设只有一个输入需要填充
    transaction_draft['vin'][0]['scriptSig'] = {
        'signature': signature.hex(),
        'publicKey': public_key_hex
    }
    
    # 计算交易ID (双SHA256哈希)
    txid = hashlib.sha256(hashlib.sha256(tx_payload.encode('utf-8')).digest()).hexdigest()

    return transaction_draft, txid

# 场景设置
curve = ecdsa.SECP256k1
alice_private_key = ecdsa.SigningKey.generate(curve=curve)
alice_public_key = alice_private_key.get_verifying_key()
alice_address = generate_address(alice_public_key.to_string("compressed"))

bob_private_key = ecdsa.SigningKey.generate(curve=curve)
bob_public_key = bob_private_key.get_verifying_key()
bob_address = generate_address(bob_public_key.to_string("compressed"))

# --- 1. 模拟全局UTXO池 ---
# 这是一个字典，模拟了全网所有未被花费的“钱”。
# 键(Key): 来源交易ID_输出索引 (txid_voutIndex)
# 值(Value): 输出的详细信息 (锁定地址, 金额)
utxo_pool = {}

# 假设Alice之前通过一笔交易收到了100聪，我们手动将其加入UTXO池。
initial_txid = '0'*63 + '1'
utxo_to_spend_key = f"{initial_txid}_0"
utxo_pool[utxo_to_spend_key] = {
    'address': alice_address,
    'amount': 100
}

print("--- 1. 交易前的UTXO池 ---")
print(json.dumps(utxo_pool, indent=2), "\n")


# --- 2. Alice手动构建一笔交易 ---
# 目标：Alice想用她那100聪的UTXO，给Bob转30聪。

# a) 构建输入 (vin)
#    明确指出要花费哪个UTXO。
vin = [
    {
        'txid': initial_txid,
        'vout': 0,
        'scriptSig': None # 签名部分暂时留空
    }
]

# b) 构建输出 (vout) - 核心教学点！
#    我们必须明确指定每一分钱(聪)的去向。
#    总输入是100聪，所以总输出也必须是100聪。
#    如果总输出小于总输入，差额将作为矿工费被矿工拿走。
vout = [
    {
        'to_address': bob_address,
        'amount': 30  # 给Bob的支付
    },
    {
        'to_address': alice_address, # 必须明确创建找零！
        'amount': 70
    }
]

# c) 组装交易草稿
transaction_draft = {
    'vin': vin,
    'vout': vout
}

print("--- 2. Alice构建的交易草稿 (尚未签名) ---")
print(json.dumps(transaction_draft, indent=2), "\n")


# --- 3. 对交易进行签名 ---
signed_transaction, new_txid = create_and_sign_transaction(transaction_draft, alice_private_key)

print("--- 3. 签名后的完整交易 ---")
print(f"新交易ID (TXID): {new_txid}")
print(json.dumps(signed_transaction, indent=2), "\n")


# --- 4. 模拟交易被确认，更新UTXO池 ---
# a) 销毁被花费的UTXO
del utxo_pool[utxo_to_spend_key]

# b) 创建新的UTXO
#    新UTXO的来源是刚刚这笔交易
#    第一个输出 (index 0) 是给Bob的
utxo_pool[f"{new_txid}_0"] = signed_transaction['vout'][0]
#    第二个输出 (index 1) 是找零给Alice的
utxo_pool[f"{new_txid}_1"] = signed_transaction['vout'][1]

print("--- 4. 交易后的UTXO池 ---")
print(json.dumps(utxo_pool, indent=2), "\n")

print("核心概念:")
print("-> UTXO被销毁: Alice那100聪的UTXO从池中消失了。")
print("-> 新UTXO被创造: 凭空出现了两个新的UTXO，一个属于Bob(30聪)，一个属于Alice(70聪)。")
print("-> 记账的本质: 比特币网络不关心Alice或Bob的“余额”，它只记录UTXO的诞生与毁灭。")