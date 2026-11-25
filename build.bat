@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>nul
if errorlevel 1 (
    echo Go was not found in PATH.
    pause
    exit /b 1
)

if not defined BUILD_VERSION (
    for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy.MM.dd-HHmm"') do set "BUILD_VERSION=%%i"
)
if not defined BUILD_TIME (
    for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-ddTHH:mm:sszzz"') do set "BUILD_TIME=%%i"
)

echo.
echo Building OpenMD.exe...
echo   Version:    %BUILD_VERSION%
echo   Build time: %BUILD_TIME%
go build -trimpath -ldflags "-s -w -H=windowsgui -X openmd/internal/buildinfo.Version=%BUILD_VERSION% -X openmd/internal/buildinfo.BuildTime=%BUILD_TIME%" -o OpenMD.exe .
if errorlevel 1 (
    echo.
    echo Build failed.
    pause
    exit /b 1
)

echo.
echo Build complete: %CD%\OpenMD.exe

set "TARGET=%~1"
if not defined TARGET (
    set /p "TARGET=Open a file or folder after build? (path, Enter to skip): "
)
if defined TARGET (
    echo Opening: %TARGET%
    start "" "%CD%\OpenMD.exe" "%TARGET%"
) else (
    echo Tip: drag a Markdown file or folder onto OpenMD.exe to open it.
)
pause
