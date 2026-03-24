@echo off
setlocal

if not exist "%~dp0aifiler.exe" (
    echo Binary not found. Triggering build...
    call "%~dp0build.bat"
    if %ERRORLEVEL% neq 0 (
        echo [E] Build failed, cannot run.
        exit /b %ERRORLEVEL%
    )
)

"%~dp0aifiler.exe" %*
exit /b %ERRORLEVEL%
