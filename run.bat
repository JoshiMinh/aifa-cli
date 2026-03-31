@echo off
setlocal
set BINARY=aifiler.exe

if not "%~1"=="" goto :run_with_args

:menu
cls
echo ==========================================
echo   Aifiler CLI - Quick Actions
echo ==========================================
echo  1. Run aifiler (Standard)
echo  2. Build / Rebuild
echo  3. Exit
echo ==========================================
set /p choice="Choose an option (1-3): "

if "%choice%"=="1" goto :run_default
if "%choice%"=="2" goto :build
if "%choice%"=="3" goto :exit
goto :menu

:run_default
if not exist "%BINARY%" call :build
"%BINARY%"
pause
goto :menu

:run_with_args
if not exist "%BINARY%" call :build
"%BINARY%" %*
exit /b %ERRORLEVEL%

:build
powershell -Command "Write-Host 'Tidying Go dependencies...' -ForegroundColor Cyan"
go mod tidy
if %ERRORLEVEL% neq 0 (
    powershell -Command "Write-Host 'X go mod tidy failed' -ForegroundColor Red"
    pause
    goto :menu
)
powershell -Command "Write-Host 'Building aifiler...' -ForegroundColor Cyan"
go build -o %BINARY% ./cmd/aifiler
if %ERRORLEVEL% neq 0 (
    powershell -Command "Write-Host 'Build failed' -ForegroundColor Red"
    pause
    goto :menu
)
powershell -Command "Write-Host 'Successfully built %BINARY%' -ForegroundColor Green"
if "%~1"=="" (
    pause
    goto :menu
)
goto :eof

:exit
exit /b 0
