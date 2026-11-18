@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>nul
if errorlevel 1 (
    echo Go was not found in PATH.
    pause
    exit /b 1
)

where git >nul 2>nul
if errorlevel 1 (
    echo Git was not found in PATH.
    pause
    exit /b 1
)

if not exist ".git" (
    echo This script must be run inside the repository root.
    pause
    exit /b 1
)

echo Running tests...
go test ./... -count=1
if errorlevel 1 (
    echo.
    echo Tests failed. Build aborted.
    pause
    exit /b 1
)

echo.
echo Building OpenMD.exe...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
if errorlevel 1 (
    echo.
    echo Build failed.
    pause
    exit /b 1
)

echo.
echo Creating source.git.bundle...
git bundle create source.git.bundle --all
if errorlevel 1 (
    echo.
    echo Source bundle creation failed.
    pause
    exit /b 1
)

echo.
echo Build complete:
echo   %CD%\OpenMD.exe
echo   %CD%\source.git.bundle
echo.
echo Note: source.git.bundle contains committed refs only.
pause
