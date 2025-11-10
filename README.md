
# Python 比特币从零实现：教学版
<!-- Simplified Chinese -->

这是一个通过纯 Python 代码、分步从零开始构建一个简化版比特币的项目。旨在帮助开发者通过亲手实践，深入理解比特币背后的核心密码学、数据结构和共识机制。

本项目专为希望通过代码理解比特币原理的开发者设计，而非生产环境的钱包或节点。

---

### ✨ 项目特色

*   **循序渐进**: 从 `step1` 到 `step10`，每个脚本都聚焦于一个独立的核心概念。
*   **代码即文档**: 高质量、教学式的注释解释了每一行代码背后的“为什么”。
*   **核心逻辑**: 聚焦于密码学、交易结构 (UTXO)、区块构建和工作量证明 (PoW)。
*   **简化非核心**: 暂时忽略了复杂的P2P网络层和Base58编码，让学习曲线更平滑。

---

### 🚀 环境准备

本项目需要 **Python 3.7+**。

1.  **克隆或下载项目**
    ```bash
    git clone <your-repo-url>
    cd <repo-folder>
    ```

2.  **安装依赖**
    项目依赖 `ecdsa` 库来处理椭圆曲线数字签名。我们推荐使用虚拟环境。

    ```bash
    # (可选) 创建并激活虚拟环境
    python3 -m venv venv
    source venv/bin/activate  # macOS/Linux
    # venv\Scripts\activate   # Windows

    # 安装所有依赖
    pip install -r requirements.txt
    ```

---

### 📚 学习路径

我们强烈建议您按照数字顺序依次执行和阅读每个脚本，因为后一步骤的知识建立在前一步骤之上。

*   **`step1_generate_keys.py`**: 生成比特币的私钥和公钥。
*   **`step2_sign_and_verify.py`**: 理解数字签名的核心过程。
*   **`step3_public_key_to_address.py`**: 从公钥派生出地址。
*   **`step4_build_simple_transaction.py`**: 构建一个结构化的交易。
*   **`step5_create_a_block.py`**: 将已签名的交易打包进一个区块。
*   **`step6_mine_a_block.py`**: 引入工作量证明（挖矿）和区块奖励。
*   **`step7_single_utxo_transaction.py`**: 介绍核心概念UTXO（单输入，多输出）。
*   **`step8_multi_input_utxo_transaction.py`**: 实现更真实的凑钱支付（多输入）。
*   **`step9_coinbase_in_utxo_block.py`**: 完整演示Coinbase交易和矿工费。
*   **`step10_simulated_network_consensus.py`**: 模拟多个节点如何通过广播达成共识。

#### 如何运行

打开您的终端，并按顺序执行：
```bash
python step1_generate_keys.py
python step2_sign_and_verify.py
# ...以此类推
```

---

### 🤝 贡献

欢迎提出任何改进建议或修复问题！请通过 Pull Request 或 Issue 进行。
