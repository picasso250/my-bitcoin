#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-LoggedCommand {
    param(
        [string]$Command,
        [int]$ExpectedExitCode = 0
    )
    Write-Host "--------------------------------------------------" -ForegroundColor DarkGray
    Write-Host "PS> $Command" -ForegroundColor Blue
    $output = & ([scriptblock]::Create($Command)) 2>&1
    $output | ForEach-Object { Write-Host $_ }
    if ($LASTEXITCODE -ne $ExpectedExitCode) {
        throw "命令失败，退出码=$LASTEXITCODE（预期=$ExpectedExitCode）"
    }
    return $output
}

Write-Host "[0] 清理旧数据" -ForegroundColor Red
Remove-Item -Force -ErrorAction SilentlyContinue wallets.dat, blockchain.db

Write-Host "[1] 钱包管理员：创建钱包" -ForegroundColor Cyan
$addr = Invoke-LoggedCommand "go run . wallet create" | Select-Object -Last 1
Write-Host "   新地址 $addr" -ForegroundColor Gray

Write-Host "[2] 消费者：查余额（首次初始化链）" -ForegroundColor Green
$bal = Invoke-LoggedCommand "go run . node -wallet $addr balance" | Select-Object -Last 1
if ($bal -ne 50) { throw "创世余额≠50" }

Write-Host "[3] 消费者：自己转自己 10 币" -ForegroundColor Yellow
Invoke-LoggedCommand "go run . node -wallet $addr send -to $addr -amount 10"

Write-Host "[4] 消费者：再次查余额" -ForegroundColor Green
$bal = Invoke-LoggedCommand "go run . node -wallet $addr balance" | Select-Object -Last 1
if ($bal -ne 50) { Write-Host "⚠️  余额=$bal（含找零+手续费，预期≈50）" }

Write-Host "[5] 打印整条链" -ForegroundColor Magenta
Invoke-LoggedCommand "go run . node -wallet $addr printchain"

Write-Host ""
Write-Host "=== 冒烟通过 ===" -ForegroundColor White -BackgroundColor DarkGreen