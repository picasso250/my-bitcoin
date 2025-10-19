# step7_single_utxo_transaction.py
#
# 教学目标：引入比特币最核心、也最独特的概念——UTXO (未花费的交易输出)。
#
# 核心洞见 (The "Aha!" Moment):
# 1. 比特币没有“账户余额”这个概念。你的“钱”不是一个数字，而是一堆分散的、
#    你尚未花费的“数字钞票”(UTXO)。
# 2. 交易的本质不是“从账户A划拨到账户B”，而是“消耗几张旧钞票，创造几张新钞票”。
# 3. UTXO是不可分割的。你不能花掉一张100元钞票的一部分。你必须把它全花掉，
#    然后接收找零——一张新的、属于你的“钞票”。
#
# 本脚本将模拟这个“单输入”过程：花费一张钞票，得到找零。

import ecdsa
import hashlib
import json
from bitcoin_lib import generate_address # 我们依然使用库里的地址生成函数

# --- 1. 在本文件中定义新的核心逻辑 ---
# 按照您的建议，我们将这个新概念的实现函数放在教学脚本内部，
# 以便学习者可以清晰地看到它的工作原理，而无需切换文件。

def create_single_input_utxo_transaction(sender_private_key, utxo_to_spend, recipient_address, amount_to_send_satoshi):
    """
    创建一个简化的、只包含单个输入的UTXO交易。

    Args:
        sender_private_key: 发送方的 ecdsa.SigningKey 对象。
        utxo_to_spend (dict): 代表要花费的那个UTXO，包含 'txid', 'vout_index', 'amount_satoshi'。
        recipient_address (str): 收款方的地址。
        amount_to_send_satoshi (int): 希望发送给收款方的金额。

    Returns:
        dict: 一个结构化的UTXO交易字典。
    """
    # 1. 从私钥派生出公钥和我们自己的地址（用于接收找零）
    sender_public_key = sender_private_key.get_verifying_key()
    sender_address = generate_address(sender_public_key.to_string("compressed"))
    
    # 2. 验证与计算
    input_amount = utxo_to_spend['amount_satoshi']
    if amount_to_send_satoshi > input_amount:
        raise ValueError("发送金额不能大于输入UTXO的金额!")
    
    change_amount = input_amount - amount_to_send_satoshi

    # 3. 构建交易结构
    tx = {
        'vin': [
            {
                'txid': utxo_to_spend['txid'],
                'vout': utxo_to_spend['vout_index'],
                'scriptSig': { # 签名脚本，暂时为空，等待签名
                    'pubKey': None,
                    'signature': None
                }
            }
        ],
        'vout': []
    }

    # 4. 创建输出
    # a. 给收款方的输出
    tx['vout'].append({
        'value': amount_to_send_satoshi,
        'scriptPubKey': recipient_address # 简化：锁定脚本直接就是地址
    })

    # b. 如果有找零，创建给自己的输出
    if change_amount > 0:
        tx['vout'].append({
            'value': change_amount,
            'scriptPubKey': sender_address
        })

    # 5. 签名
    # 为了签名，我们需要一个交易的“指纹”。我们通过确定性序列化交易（不含签名脚本的部分）来实现。
    tx_copy_for_signing = json.loads(json.dumps(tx)) # 创建一个深拷贝来进行修改
    
    # --- BUG FIX ---
    # 'vin' 是一个列表，所以我们必须通过索引 [0] 来访问第一个输入项。
    tx_copy_for_signing['vin'][0]['scriptSig'] = {} # 清空签名部分以生成待签名的哈希
    
    tx_payload = json.dumps(tx_copy_for_signing, sort_keys=True, separators=(',', ':'))
    tx_hash = hashlib.sha256(tx_payload.encode('utf-8')).digest()
    
    signature = sender_private_key.sign(tx_hash)

    # 6. 将签名和公钥填入交易
    # --- BUG FIX ---
    # 同样，这里也需要通过索引 [0] 来访问正确的输入项。
    tx['vin'][0]['scriptSig']['pubKey'] = sender_public_key.to_string('compressed').hex()
    tx['vin'][0]['scriptSig']['signature'] = signature.hex()
    
    return tx


# --- 2. 演示场景 ---
if __name__ == "__main__":
    
    # a. 身份设置
    curve = ecdsa.SECP256k1
    alice_private_key = ecdsa.SigningKey.generate(curve=curve)
    alice_address = generate_address(alice_private_key.get_verifying_key().to_string("compressed"))
    
    bob_private_key = ecdsa.SigningKey.generate(curve=curve)
    bob_address = generate_address(bob_private_key.get_verifying_key().to_string("compressed"))

    print("--- 身份 ---")
    print(f"Alice's Address: {alice_address}")
    print(f"Bob's Address:   {bob_address}\n")

    # b. 模拟全局 UTXO 池
    # 假设有一笔之前的交易(txid: genesis_tx_id)，其第0个输出给了Alice 100聪。
    # 我们用 f"{txid}_{vout_index}" 作为UTXO的唯一标识符。
    genesis_tx_id = 'a1b2c3d4' * 8 
    utxo_pool = {
        f"{genesis_tx_id}_0": {
            'owner_address': alice_address,
            'amount_satoshi': 100
        }
    }
    
    print("--- 交易前: UTXO 池状态 ---")
    print(json.dumps(utxo_pool, indent=2), "\n")

    # c. Alice 决定花费她的UTXO
    utxo_to_spend_key = f"{genesis_tx_id}_0"
    utxo_to_spend_details = {
        'txid': genesis_tx_id,
        'vout_index': 0,
        'amount_satoshi': utxo_pool[utxo_to_spend_key]['amount_satoshi']
    }
    
    # d. 调用本文件内的函数创建交易
    # Alice 花费她的100聪UTXO，给Bob转30聪，期望找零70聪。
    new_transaction = create_single_input_utxo_transaction(
        sender_private_key=alice_private_key,
        utxo_to_spend=utxo_to_spend_details,
        recipient_address=bob_address,
        amount_to_send_satoshi=30
    )

    print("--- 创建的UTXO交易 ---")
    print(json.dumps(new_transaction, indent=2), "\n")

    # e. 模拟节点验证和更新UTXO池
    # (在一个真实的节点中，验证会更复杂，这里我们假设签名有效)
    print("--- 交易后: 更新UTXO池 ---")
    
    # 1. 移除被花费的UTXO
    del utxo_pool[utxo_to_spend_key]
    print(f"  - [消耗] 移除了UTXO: {utxo_to_spend_key}")

    # 2. 添加新的UTXO
    # 计算新交易的ID (双SHA256哈希，真实比特币的做法)
    tx_payload_final = json.dumps(new_transaction, sort_keys=True, separators=(',', ':'))
    hash1 = hashlib.sha256(tx_payload_final.encode('utf-8')).digest()
    new_tx_id = hashlib.sha256(hash1).hexdigest()
    
    # 遍历新交易的输出，并将它们作为新的UTXO添加到池中
    for index, vout in enumerate(new_transaction['vout']):
        new_utxo_key = f"{new_tx_id}_{index}"
        utxo_pool[new_utxo_key] = {
            'owner_address': vout['scriptPubKey'],
            'amount_satoshi': vout['value']
        }
        owner_name = 'Bob' if vout['scriptPubKey'] == bob_address else 'Alice (Change)'
        print(f"  - [创造] 添加了新的UTXO: {new_utxo_key} (价值: {vout['value']}, 归属: {owner_name})")

    print("\n--- 最终UTXO池状态 ---")
    print(json.dumps(utxo_pool, indent=2))
    print("\n核心概念:")
    print("-> UTXO池中代表Alice的“钞票”消失了。")
    print("-> 同时，池中出现了一张属于Bob的新“钞票”（30聪）和一张属于Alice的新“找零钞票”（70聪）。")
    print("-> 这就是比特币记账的本质：销毁旧价值，创造新价值。")