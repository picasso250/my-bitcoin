# Bitcoin from Scratch in Python: Educational Implementation

A step-by-step project that builds a simplified version of Bitcoin from scratch using pure Python code. Designed to help developers deeply understand Bitcoin's core cryptography, data structures, and consensus mechanisms through hands-on practice.

This project is specifically designed for developers who want to understand Bitcoin principles through code, not as a production-ready wallet or node.

---

### ✨ Project Features

*   **Progressive Learning**: From `step1` to `step10`, each script focuses on a single core concept.
*   **Code as Documentation**: High-quality, educational comments explain the "why" behind every line of code.
*   **Core Logic**: Focus on cryptography, transaction structure (UTXO), block construction, and Proof of Work (PoW).
*   **Simplified Non-essentials**: Temporarily ignores complex P2P networking and Base58 encoding for a smoother learning curve.

---

### 🚀 Environment Setup

This project requires **Python 3.9+**.

1.  **Clone or Download the Project**

        git clone https://github.com/picasso250/my-bitcoin.git
        cd my-bitcoin

2.  **Install Dependencies**
    The project depends on the `ecdsa` library for elliptic curve digital signatures. We recommend using a virtual environment.

        # (Optional) Create and activate virtual environment
        python3 -m venv venv
        source venv/bin/activate  # macOS/Linux
        # venv\Scripts\activate   # Windows

        # Install all dependencies
        pip install -r requirements.txt

---

### 📚 Learning Path

We strongly recommend executing and reading each script in numerical order, as knowledge from later steps builds upon earlier ones.

**💡 Learning Tip: Pair Programming with AI**

During your learning journey, we especially encourage you to interact frequently with AI assistants like **Gemini**. When you encounter code you don't understand, unclear concepts, or want to explore cryptographic details more deeply, don't hesitate to ask. **Sharing code snippets and your questions directly with AI is an excellent way to consolidate knowledge and spark inspiration.** Making this a habit will greatly accelerate your learning process.

*   **`step1_generate_keys.py`**: Generate Bitcoin private and public keys.
*   **`step2_sign_and_verify.py`**: Understand the core process of digital signatures.
*   **`step3_public_key_to_address.py`**: Derive addresses from public keys.
*   **`step4_build_simple_transaction.py`**: Build a structured transaction.
*   **`step5_create_a_block.py`**: Pack signed transactions into a block.
*   **`step6_mine_a_block.py`**: Introduce Proof of Work (mining) and block rewards.
*   **`step7_single_utxo_transaction.py`**: Introduce the core concept of UTXO (single input, multiple outputs).
*   **`step8_multi_input_utxo_transaction.py`**: Implement realistic payment scenarios (multiple inputs).
*   **`step9_coinbase_in_utxo_block.py`**: Complete demonstration of Coinbase transactions and miner fees.
*   **`step10_simulated_network_consensus.py`**: Simulate how multiple nodes achieve consensus through broadcasting.

#### How to Run

Open your terminal and execute in order:

    python step1_generate_keys.py
    python step2_sign_and_verify.py
    # ...and so on

---

### 🤝 Contributing

Welcome any improvement suggestions or bug fixes! Please contribute through Pull Requests or Issues.