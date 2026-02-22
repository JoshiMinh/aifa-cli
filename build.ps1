$ErrorActionPreference = 'Stop'

Set-Location $PSScriptRoot

go build -o aifiler.exe ./cmd/aifiler
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "Built $PSScriptRoot\aifiler.exe"
