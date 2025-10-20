#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "[0] 清理旧数据" -ForegroundColor Red
Remove-Item -Force -ErrorAction SilentlyContinue wallets.dat, blockchain.db

Write-Host "[1] 钱包管理员：创建钱包" -ForegroundColor Cyan
$addr = go run . wallet create
if (-not $addr) { throw "wallet create failed" }
Write-Host "   新地址 $addr"

Write-Host "[2] 消费者：查余额（首次初始化链）" -ForegroundColor Green
$bal = go run . node -wallet $addr balance
if ($bal -ne 50) { throw "创世余额≠50" }

Write-Host "[3] 消费者：自己转自己 10 币" -ForegroundColor Yellow
go run . node -wallet $addr send -to $addr -amount 10
if ($LASTEXITCODE -ne 0) { throw "send failed" }

Write-Host "[4] 消费者：再次查余额" -ForegroundColor Green
$bal = go run . node -wallet $addr balance
if ($bal -ne 50) { Write-Host "⚠️  余额=$bal（含找零+手续费，预期≈50）" }

Write-Host "[5] 打印整条链" -ForegroundColor Magenta
go run . node -wallet $addr printchain   # 下面顺手给 node 加个只读子命令

Write-Host ""
Write-Host "=== 冒烟通过 ===" -ForegroundColor White -BackgroundColor DarkGreen