# Python从零实现比特币：教学版

一个使用纯Python代码从零开始构建简化版比特币的逐步项目。旨在帮助开发者通过动手实践深入理解比特币的核心密码学、数据结构和共识机制。

本项目专为希望通过代码理解比特币原理的开发者设计，而非生产就绪的钱包或节点。

---

### ✨ 项目特色

*   **渐进式学习**：从`step1`到`step10`，每个脚本专注于单一核心概念
*   **代码即文档**：高质量的教育性注释解释每行代码背后的"为什么"
*   **核心逻辑**：专注于密码学、交易结构(UTXO)、区块构建和工作量证明(PoW)
*   **简化非核心**：暂时忽略复杂的P2P网络和Base58编码，让学习曲线更平滑

---

### 🚀 环境设置

本项目需要**Python 3.9+**。

1.  **克隆或下载项目**

        git clone https://github.com/picasso250/my-bitcoin.git
        cd my-bitcoin

2.  **安装依赖**
    项目依赖`ecdsa`库进行椭圆曲线数字签名。建议使用虚拟环境。

        # (可选)创建并激活虚拟环境
        python3 -m venv venv
        source venv/bin/activate  # macOS/Linux
        # venv\Scripts\activate   # Windows

        # 安装所有依赖
        pip install -r requirements.txt

---

### 📚 学习路径

我们强烈建议按数字顺序执行和阅读每个脚本，因为后续步骤的知识建立在前面步骤的基础上。

**💡 学习技巧：与AI结对编程**

在学习过程中，我们特别鼓励您与**Gemini**等AI助手频繁互动。当遇到不理解的代码、不清楚的概念，或想要更深入地探索密码学细节时，不要犹豫。**直接将代码片段和您的问题分享给AI是巩固知识和激发灵感的绝佳方式。**养成这个习惯将大大加速您的学习进程。

*   **`step1_generate_keys.py`**：生成比特币私钥和公钥
*   **`step2_sign_and_verify.py`**：理解数字签名的核心过程
*   **`step3_public_key_to_address.py`**：从公钥推导地址
*   **`step4_build_simple_transaction.py`**：构建结构化交易
*   **`step5_create_a_block.py`**：将已签名交易打包成区块
*   **`step6_mine_a_block.py`**：介绍工作量证明(挖矿)和区块奖励
*   **`step7_single_utxo_transaction.py`**：介绍UTXO核心概念(单输入，多输出)
*   **`step8_multi_input_utxo_transaction.py`**：实现现实支付场景(多输入)
*   **`step9_coinbase_in_utxo_block.py`**：完整演示Coinbase交易和矿工费
*   **`step10_simulated_network_consensus.py`**：模拟多节点如何通过广播达成共识

#### 运行方式

打开终端，按顺序执行：

    python step1_generate_keys.py
    python step2_sign_and_verify.py
    # ...以此类推

---

### 🤝 贡献

欢迎任何改进建议或错误修复！请通过Pull Requests或Issues贡献。