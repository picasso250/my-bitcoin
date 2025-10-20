#!/usr/bin/env pwsh
# GoCoin 冒烟脚本 —— 纯 go run 版，不留编译残渣
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "[0] 强制清理旧数据..." -ForegroundColor Red
Remove-Item -Force -ErrorAction SilentlyContinue wallets.dat, blockchain.db

Write-Host "[1] 创建钱包..." -ForegroundColor Cyan
go run . createwallet
if ($LASTEXITCODE -ne 0) { throw "创建钱包失败" }

Write-Host "[1.1] 获取钱包地址..." -ForegroundColor Cyan
$addr = go run . getaddress
if (-not $addr) { throw "无法获取钱包地址" }
Write-Host "   得到地址 $addr"

Write-Host "[2] 查初始余额..." -ForegroundColor Green
go run . getbalance -address $addr

Write-Host "[3] 自己转自己 10 币..." -ForegroundColor Yellow
go run . send -from $addr -to $addr -amount 10
if ($LASTEXITCODE -ne 0) { throw "交易失败" }

Write-Host "[4] 再次查余额..." -ForegroundColor Green
go run . getbalance -address $addr

Write-Host "[5] 打印链..." -ForegroundColor Magenta
go run . printchain

Write-Host ""
Write-Host "=== 冒烟结束 ===" -ForegroundColor White -BackgroundColor DarkGreen
Write-Host "如果看到 2 个区块且没有 panic，就算通过！"