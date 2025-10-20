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

Write-Host "[1] 创建钱包" -ForegroundColor Cyan
$addr = Invoke-LoggedCommand "go run . wallet create" | Select-Object -Last 1
Write-Host "   新地址 $addr" -ForegroundColor Gray

Write-Host "[2] 启动本地矿工（10 秒后自动停止）" -ForegroundColor Green
$job = Start-Job -ScriptBlock {
    Set-Location $using:PWD
    & go run . miner local --coinbase $using:addr
}
Start-Sleep -Seconds 10
if ($job.State -eq "Running") { Stop-Job $job }
Receive-Job $job
Remove-Job $job

Write-Host ""
Write-Host "=== 冒烟通过 ===" -ForegroundColor White -BackgroundColor DarkGreen