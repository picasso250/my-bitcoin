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

# 怎么写呢？