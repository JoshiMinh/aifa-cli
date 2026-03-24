@echo off
setlocal

powershell -Command "Write-Host '[-] Tiding Go dependencies...' -ForegroundColor Cyan"
go mod tidy
if %ERRORLEVEL% neq 0 (
    powershell -Command "Write-Host '[!] go mod tidy failed!' -ForegroundColor Red"
    exit /b %ERRORLEVEL%
)

powershell -Command "Write-Host '[-] Building aifiler...' -ForegroundColor Cyan"
go build -o aifiler.exe ./cmd/aifiler
if %ERRORLEVEL% neq 0 (
    powershell -Command "Write-Host '[!] Build failed!' -ForegroundColor Red"
    exit /b %ERRORLEVEL%
)

powershell -Command "Write-Host '[+] Successfully built aifiler.exe' -ForegroundColor Green"
endlocal
