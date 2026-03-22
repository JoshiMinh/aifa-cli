<#
.SYNOPSIS
Builds the aifiler Go binary.

.DESCRIPTION
This script compiles the aifiler CLI tool for Windows using 'go build'.
It ensures the execution location is the root of the repository before compiling.

.EXAMPLE
.\build.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

Write-Host "Building aifiler..." -ForegroundColor Cyan
go build -o aifiler.exe ./cmd/aifiler
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "Successfully built aifiler.exe" -ForegroundColor Green
