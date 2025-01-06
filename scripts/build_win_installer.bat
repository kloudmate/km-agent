@echo off

cd ..

:: Build the Go binary
echo Building Go binary...
go build -o builds/bin/kmagent.exe cmd/kmagent/kmagent.go

:: Check if build was successful
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

echo Build successful!

makensis ./scripts/win_installation_helper.nsi

:: Check if installer build script execution was successful
if %ERRORLEVEL% NEQ 0 (
    echo NSIS Installer Build failed!
    exit /b %ERRORLEVEL%
)