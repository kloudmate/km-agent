@echo off
setlocal enabledelayedexpansion

rem Set the config directory path using %USERPROFILE% (Windows equivalent of $HOME)
set CONFIG_DIR=%USERPROFILE%\.kloudmate

:create-config-dir
echo.
echo ========================================
echo Starting Task: Creating Config Directory
echo ========================================
if not exist "%CONFIG_DIR%" (
    mkdir "%CONFIG_DIR%"
    echo Created directory: %CONFIG_DIR%
) else (
    echo Directory already exists: %CONFIG_DIR%
)
echo Task Completed: Config Directory Setup
echo.
goto :eof

:setup-config
call :create-config-dir
echo.
echo ==================================================
echo Starting Task: Setting Up Initial Configuration
echo ==================================================
echo Copying default configuration file...
xcopy /y /q "configs\default.yaml" "%CONFIG_DIR%\agent-config.yaml"
echo Task Completed: Initial Configuration Setup
echo.
goto :eof

:build
call :setup-config
echo.
echo ==================================================
echo Starting Task: Building Application
echo ==================================================
echo Running go build command...
go build cmd\kmagent\main.go
if %ERRORLEVEL% EQU 0 (
    echo Build completed successfully!
) else (
    echo Build failed with error code: %ERRORLEVEL%
)
echo Task Completed: Build Process
echo.
goto :eof

:run
call :build
echo.
echo ==================================================
echo Starting Task: Running Application
echo ==================================================
echo Launching the application...
main.exe
echo Task Completed: Application Execution
echo.
goto :eof

rem If no arguments provided, run the default target (run)
if "%~1"=="" (
    call :run
) else (
    call :%~1
)

echo.
echo ==================================================
echo All tasks completed!
echo ==================================================
echo.
pause