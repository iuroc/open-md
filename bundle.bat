@echo off
setlocal
cd /d "%~dp0"

where git >nul 2>nul
if errorlevel 1 (
    echo Git was not found in PATH.
    pause
    exit /b 1
)

git bundle create source.git.bundle --all
if errorlevel 1 (
    echo.
    echo Bundle creation failed.
    pause
    exit /b 1
)

echo.
echo Created: %CD%\source.git.bundle
pause
