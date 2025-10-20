@echo off
rem === GoCoin 扁平化重命名脚本 ===
rem 用法：将此.bat放入原gocoin根目录，双击运行即可

rem 1. blockchain 模块
move blockchain\block.go          bc_block.go
move blockchain\blockchain.go     bc_blockchain.go
move blockchain\iterator.go       bc_iterator.go
move blockchain\json.go           bc_json.go
move blockchain\proofofwork.go    bc_pow.go
move blockchain\transaction.go    bc_tx.go
move blockchain\utxo_set.go       bc_utxo.go

rem 2. p2p 模块
move p2p\codec.go      p2p_codec.go
move p2p\messages.go   p2p_messages.go
move p2p\node.go       p2p_node.go
move p2p\peer.go       p2p_peer.go
move p2p\tx_pool.go    p2p_txpool.go
move p2p\types.go      p2p_types.go

rem 3. wallet 模块
move wallet\wallet.go  wallet.go
move wallet\wallets.go wallets.go

rem 4. 清理空目录
rd blockchain
rd p2p
rd wallet

echo 扁平化完成！
pause