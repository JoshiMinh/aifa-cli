$ErrorActionPreference = 'Stop'

Set-Location $PSScriptRoot

if (-not (Test-Path .\aifiler.exe)) {
    go build -o aifiler.exe ./cmd/aifiler
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

& .\aifiler.exe @args
exit $LASTEXITCODE
