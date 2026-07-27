@echo off
REM ============================================
REM GBS POS-CMS Auto Deploy Script (Windows)
REM ============================================
REM Usage: double-click file ini atau jalankan
REM dari cmd/PowerShell
REM ============================================

setlocal EnableExtensions

REM === Configuration ===
set REGISTRY=ghcr.io/fariziadam11
set VPS_HOST=159.89.204.100
set VPS_USER=adam
set VPS_PATH=/opt/gbs/gbs-pos-cms-api
set VPS_ENV_FILE=/opt/gbs/.env

REM === Color codes (ANSI) ===
set BLUE=[94m
set GREEN=[92m
set YELLOW=[93m
set RED=[91m
set CYAN=[96m
set NC=[0m

echo.
echo %BLUE%===========================================%NC%
echo %BLUE%  GBS POS-CMS Deploy Script%NC%
echo %BLUE%===========================================%NC%
echo %BLUE%  Service : all%NC%
echo %BLUE%  Target  : %VPS_USER%@%VPS_HOST%%NC%
echo %BLUE%  Path    : %VPS_PATH%%NC%
echo %BLUE%===========================================%NC%
echo.

REM === Parse Arguments ===
set "SERVICE=%~1"
if "%SERVICE%"=="" set "SERVICE=all"

echo %CYAN%[STEP] Checking Docker...%NC%
docker --version >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR] Docker not found!%NC%
    echo Make sure Docker Desktop is installed and running.
    pause
    exit /b 1
)
echo %GREEN%[OK] Docker installed%NC%
echo.

REM =============================================
REM 1. Build & Push Images
REM =============================================
echo %CYAN%[STEP] Building Docker images...%NC%

if /i "%SERVICE%"=="pos" goto :build_pos
if /i "%SERVICE%"=="cms" goto :build_cms

REM Build ALL (default)
:build_pos
echo %BLUE%[INFO] Building gbs-pos-api...%NC%
docker build -t %REGISTRY%/gbs-pos-api:latest -f gbs-pos-api/Dockerfile .
if errorlevel 1 (
    echo %RED%[ERROR] Build POS failed!%NC%
    pause
    exit /b 1
)
echo %GREEN%[SUCCESS] Built: %REGISTRY%/gbs-pos-api:latest%NC%
docker push %REGISTRY%/gbs-pos-api:latest
echo %GREEN%[SUCCESS] Pushed: %REGISTRY%/gbs-pos-api:latest%NC%
echo.

:build_cms
if /i "%SERVICE%"=="pos" goto :deploy_vps
echo %BLUE%[INFO] Building gbs-cms-api...%NC%
docker build -t %REGISTRY%/gbs-cms-api:latest -f gbs-cms-api/Dockerfile .
if errorlevel 1 (
    echo %RED%[ERROR] Build CMS failed!%NC%
    pause
    exit /b 1
)
echo %GREEN%[SUCCESS] Built: %REGISTRY%/gbs-cms-api:latest%NC%
docker push %REGISTRY%/gbs-cms-api:latest
echo %GREEN%[SUCCESS] Pushed: %REGISTRY%/gbs-cms-api:latest%NC%
echo.

REM =============================================
REM 2. Deploy ke VPS
REM =============================================
:deploy_vps
echo %CYAN%[STEP] Deploying to VPS...%NC%

ssh %VPS_USER%@%VPS_HOST% "sudo docker compose -f %VPS_PATH%/docker-compose.prod.yml --env-file %VPS_ENV_FILE% pull && sudo docker compose -f %VPS_PATH%/docker-compose.prod.yml --env-file %VPS_ENV_FILE% up -d --force-recreate && sudo docker image prune -f && sudo docker compose -f %VPS_PATH%/docker-compose.prod.yml --env-file %VPS_ENV_FILE% ps"

if errorlevel 1 (
    echo.
    echo %RED%[ERROR] VPS deployment failed!%NC%
    echo Check your SSH connection and try again.
    pause
    exit /b 1
)

echo.
echo %CYAN%[STEP] Health check (waiting 15s for containers to start)...%NC%
timeout /t 15 /nobreak >nul

REM Check POS API
curl -s --connect-timeout 5 http://%VPS_HOST%:8080/health >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR] gbs-pos-api health check FAILED%NC%
) else (
    echo %GREEN%[OK] gbs-pos-api is healthy%NC%
)

REM Check CMS API
curl -s --connect-timeout 5 http://%VPS_HOST%:8081/health >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR] gbs-cms-api health check FAILED%NC%
) else (
    echo %GREEN%[OK] gbs-cms-api is healthy%NC%
)

echo.
echo %GREEN%===========================================%NC%
echo %GREEN%  Deploy completed!%NC%
echo %GREEN%===========================================%NC%

pause
endlocal
