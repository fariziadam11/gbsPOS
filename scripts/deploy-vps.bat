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
set VPS_HOST=159.89.198.208
set VPS_USER=bedul
set VPS_PATH=/home/bedul/gbs-pos-cms-api
set VPS_ENV_FILE=/home/bedul/gbs-pos-cms-api/.env

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
set "COMPOSE_SERVICES=pos-api cms-api"
set "COMPOSE_FILES=-f compose.yml"
set "BTN_SNAP_PRIVATE_KEY=gbs-pos-api\.secrets\btn-snap-private.pem"
if /i "%SERVICE%"=="pos" set "COMPOSE_SERVICES=pos-api"
if /i "%SERVICE%"=="cms" set "COMPOSE_SERVICES=cms-api"

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
if errorlevel 1 (
    echo %RED%[ERROR] Push POS failed!%NC%
    pause
    exit /b 1
)
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
if errorlevel 1 (
    echo %RED%[ERROR] Push CMS failed!%NC%
    pause
    exit /b 1
)
echo %GREEN%[SUCCESS] Pushed: %REGISTRY%/gbs-cms-api:latest%NC%
echo.

REM =============================================
REM 2. Deploy ke VPS
REM =============================================
:deploy_vps
echo %CYAN%[STEP] Deploying to VPS...%NC%

REM Copy the BTN key out-of-band from the image and mount it read-only.
if /i not "%SERVICE%"=="cms" (
    if exist "%BTN_SNAP_PRIVATE_KEY%" (
        echo %CYAN%[STEP] Installing BTN SNAP secret on VPS...%NC%
        ssh %VPS_USER%@%VPS_HOST% "mkdir -p %VPS_PATH%/gbs-pos-api/.secret && chmod 700 %VPS_PATH%/gbs-pos-api/.secret"
        if errorlevel 1 (
            echo %RED%[ERROR] Could not prepare BTN secret directory on VPS!%NC%
            pause
            exit /b 1
        )
        scp docker-compose.btn-snap.yml %VPS_USER%@%VPS_HOST%:%VPS_PATH%/compose.btn-snap.yml
        if errorlevel 1 (
            echo %RED%[ERROR] Could not upload BTN Compose overlay!%NC%
            pause
            exit /b 1
        )
        scp "%BTN_SNAP_PRIVATE_KEY%" %VPS_USER%@%VPS_HOST%:%VPS_PATH%/gbs-pos-api/.secret/btn-snap-private.pem
        if errorlevel 1 (
            echo %RED%[ERROR] Could not upload BTN private key!%NC%
            pause
            exit /b 1
        )
        ssh %VPS_USER%@%VPS_HOST% "chmod 600 %VPS_PATH%/gbs-pos-api/.secret/btn-snap-private.pem"
        if errorlevel 1 (
            echo %RED%[ERROR] Could not restrict BTN private key permissions on VPS!%NC%
            pause
            exit /b 1
        )
        set "COMPOSE_FILES=-f compose.yml -f compose.btn-snap.yml"
    )
)

ssh %VPS_USER%@%VPS_HOST% "cd %VPS_PATH% && docker compose --env-file %VPS_ENV_FILE% %COMPOSE_FILES% pull %COMPOSE_SERVICES% && docker compose --env-file %VPS_ENV_FILE% %COMPOSE_FILES% up -d --no-deps %COMPOSE_SERVICES% && docker compose --env-file %VPS_ENV_FILE% %COMPOSE_FILES% ps"

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
set "HEALTH_FAILED=0"

REM Check POS API
curl -fsS --connect-timeout 10 https://pos-api.armmada.id/health >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR] gbs-pos-api health check FAILED%NC%
    set "HEALTH_FAILED=1"
) else (
    echo %GREEN%[OK] gbs-pos-api is healthy%NC%
)

REM Check CMS API
curl -fsS --connect-timeout 10 https://cms-api.armmada.id/health >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR] gbs-cms-api health check FAILED%NC%
    set "HEALTH_FAILED=1"
) else (
    echo %GREEN%[OK] gbs-cms-api is healthy%NC%
)

if "%HEALTH_FAILED%"=="1" (
    echo %RED%[ERROR] Deployment finished but at least one API is unhealthy.%NC%
    pause
    exit /b 1
)

echo.
echo %GREEN%===========================================%NC%
echo %GREEN%  Deploy completed!%NC%
echo %GREEN%===========================================%NC%

pause
endlocal
