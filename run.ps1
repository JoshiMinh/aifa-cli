<#
.SYNOPSIS
Runs the aifiler CLI tool.

.DESCRIPTION
This script runs the compiled aifiler CLI tool. If the binary does not exist, it builds it first.
All arguments passed to this script will be forwarded directly to the executable.

.PARAMETER args
Arguments passed directly to the aifiler executable.

.EXAMPLE
.\run.ps1 doctor
.\run.ps1 rename "rename all files"
#>
[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$CommandArgs
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Test-Path .\aifiler.exe)) {
    Write-Host "Binary not found. Building aifiler..." -ForegroundColor Cyan
    go build -o aifiler.exe ./cmd/aifiler
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
        exit $LASTEXITCODE
    }
    Write-Host "Successfully built aifiler.exe" -ForegroundColor Green
}

& .\aifiler.exe $CommandArgs
exit $LASTEXITCODE
