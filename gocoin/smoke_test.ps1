#!/usr/bin/env pwsh
# GoCoin 冒烟脚本 —— 先清场、再提纯净地址
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "[0] 强制清理旧数据..." -ForegroundColor Red
Remove-Item -Force -ErrorAction SilentlyContinue wallets.dat, blockchain.db

Write-Host "[1] 编译..." -ForegroundColor Cyan
go build
if ($LASTEXITCODE -ne 0) { throw "编译失败" }

Write-Host "[2] 创建钱包..." -ForegroundColor Cyan
$raw = (& .\gocoin.exe createwallet) -join "`n"                     # 捕获所有输出
if ($raw -match 'Your address:\s*([a-fA-F0-9]+)') {                 # 正则抠地址
    $addr = $Matches[1]
} else { throw "无法提取钱包地址，原始输出：$raw" }
Write-Host "   得到地址 $addr"

Write-Host "[3] 查初始余额..." -ForegroundColor Green
& .\gocoin.exe getbalance -address $addr

Write-Host "[4] 自己转自己 10 币..." -ForegroundColor Yellow
& .\gocoin.exe send -from $addr -to $addr -amount 10
if ($LASTEXITCODE -ne 0) { throw "交易失败" }

Write-Host "[5] 再次查余额..." -ForegroundColor Green
& .\gocoin.exe getbalance -address $addr

Write-Host "[6] 打印链..." -ForegroundColor Magenta
& .\gocoin.exe printchain

Write-Host ""
Write-Host "=== 冒烟结束 ===" -ForegroundColor White -BackgroundColor DarkGreen
Write-Host "如果看到 2 个区块且没有 panic，就算通过！"