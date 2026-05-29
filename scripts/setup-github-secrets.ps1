# GitHub Secrets Setup Helper Script
# This script helps you copy keys and values for GitHub Secrets

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  GitHub Secrets Setup Helper" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$sshKeyPath = "$env:USERPROFILE\.ssh\gbs-deploy"
$sshPubKeyPath = "$env:USERPROFILE\.ssh\gbs-deploy.pub"

# Check if keys exist
if (-not (Test-Path $sshKeyPath)) {
    Write-Host "❌ SSH keys not found!" -ForegroundColor Red
    Write-Host "Please run: ssh-keygen -t ed25519 -C 'github-actions-gbs-deploy' -f $sshKeyPath" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ SSH keys found!`n" -ForegroundColor Green

# Function to copy to clipboard and display
function Copy-AndDisplay {
    param (
        [string]$Title,
        [string]$Content,
        [string]$SecretName
    )
    
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "📋 $Title" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "`nSecret Name: " -NoNewline
    Write-Host "$SecretName" -ForegroundColor Green
    Write-Host "`nValue (copied to clipboard):" -ForegroundColor Gray
    Write-Host $Content -ForegroundColor DarkGray
    Write-Host "`n✅ Copied to clipboard! Paste in GitHub." -ForegroundColor Green
    Write-Host "`nPress any key to continue..." -ForegroundColor Yellow
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    Write-Host "`n"
}

# Menu
while ($true) {
    Clear-Host
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "  GitHub Secrets Setup Menu" -ForegroundColor Cyan
    Write-Host "========================================`n" -ForegroundColor Cyan
    
    Write-Host "1. Copy STAGING_SSH_KEY" -ForegroundColor White
    Write-Host "2. Copy PRODUCTION_SSH_KEY" -ForegroundColor White
    Write-Host "3. View Public Key (for servers)" -ForegroundColor White
    Write-Host "4. Generate Secure Password" -ForegroundColor White
    Write-Host "5. Generate JWT Secret" -ForegroundColor White
    Write-Host "6. Open GitHub Secrets Page" -ForegroundColor White
    Write-Host "7. Show All Secrets Template" -ForegroundColor White
    Write-Host "0. Exit" -ForegroundColor Red
    
    Write-Host "`nSelect option: " -NoNewline -ForegroundColor Yellow
    $choice = Read-Host
    
    switch ($choice) {
        "1" {
            $privateKey = Get-Content $sshKeyPath -Raw
            $privateKey | Set-Clipboard
            Copy-AndDisplay "Staging SSH Private Key" $privateKey "STAGING_SSH_KEY"
        }
        "2" {
            $privateKey = Get-Content $sshKeyPath -Raw
            $privateKey | Set-Clipboard
            Copy-AndDisplay "Production SSH Private Key" $privateKey "PRODUCTION_SSH_KEY"
        }
        "3" {
            $publicKey = Get-Content $sshPubKeyPath -Raw
            $publicKey | Set-Clipboard
            Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "📋 SSH Public Key (for servers)" -ForegroundColor Yellow
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "`n$publicKey" -ForegroundColor Green
            Write-Host "`n✅ Copied to clipboard!" -ForegroundColor Green
            Write-Host "`nAdd this to ~/.ssh/authorized_keys on your servers" -ForegroundColor Yellow
            Write-Host "`nPress any key to continue..." -ForegroundColor Yellow
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        "4" {
            Write-Host "`nGenerating secure password..." -ForegroundColor Yellow
            $bytes = New-Object byte[] 32
            [Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
            $password = [Convert]::ToBase64String($bytes)
            $password | Set-Clipboard
            Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "🔐 Secure Password Generated" -ForegroundColor Yellow
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "`n$password" -ForegroundColor Green
            Write-Host "`n✅ Copied to clipboard!" -ForegroundColor Green
            Write-Host "`nUse this for POSTGRES_PASSWORD in .env.production" -ForegroundColor Yellow
            Write-Host "`nPress any key to continue..." -ForegroundColor Yellow
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        "5" {
            Write-Host "`nGenerating JWT secret..." -ForegroundColor Yellow
            $bytes = New-Object byte[] 32
            [Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
            $jwtSecret = [BitConverter]::ToString($bytes).Replace("-", "").ToLower()
            $jwtSecret | Set-Clipboard
            Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "🔑 JWT Secret Generated" -ForegroundColor Yellow
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "`n$jwtSecret" -ForegroundColor Green
            Write-Host "`n✅ Copied to clipboard!" -ForegroundColor Green
            Write-Host "`nUse this for JWT_SECRET in .env.production" -ForegroundColor Yellow
            Write-Host "`nPress any key to continue..." -ForegroundColor Yellow
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        "6" {
            Write-Host "`nOpening GitHub Secrets page..." -ForegroundColor Yellow
            Start-Process "https://github.com/fariziadam11/gbs-pos-cms-api/settings/secrets/actions"
            Write-Host "✅ Browser opened!" -ForegroundColor Green
            Start-Sleep -Seconds 2
        }
        "7" {
            Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "📋 All GitHub Secrets Template" -ForegroundColor Yellow
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "`nAdd these secrets in GitHub:" -ForegroundColor White
            Write-Host "`n1. STAGING_SSH_KEY" -ForegroundColor Green
            Write-Host "   Value: [Use option 1 to copy]" -ForegroundColor Gray
            Write-Host "`n2. STAGING_HOST" -ForegroundColor Green
            Write-Host "   Value: [Your staging server IP or hostname]" -ForegroundColor Gray
            Write-Host "   Example: staging.gbs-pos.com or 192.168.1.100" -ForegroundColor DarkGray
            Write-Host "`n3. STAGING_USER" -ForegroundColor Green
            Write-Host "   Value: [SSH username]" -ForegroundColor Gray
            Write-Host "   Example: ubuntu or deploy" -ForegroundColor DarkGray
            Write-Host "`n4. PRODUCTION_SSH_KEY" -ForegroundColor Green
            Write-Host "   Value: [Use option 2 to copy]" -ForegroundColor Gray
            Write-Host "`n5. PRODUCTION_HOST" -ForegroundColor Green
            Write-Host "   Value: [Your production server IP or hostname]" -ForegroundColor Gray
            Write-Host "   Example: api.gbs-pos.com or 192.168.1.200" -ForegroundColor DarkGray
            Write-Host "`n6. PRODUCTION_USER" -ForegroundColor Green
            Write-Host "   Value: [SSH username]" -ForegroundColor Gray
            Write-Host "   Example: ubuntu or deploy" -ForegroundColor DarkGray
            Write-Host "`n7. SLACK_WEBHOOK (Optional)" -ForegroundColor Green
            Write-Host "   Value: [Slack webhook URL]" -ForegroundColor Gray
            Write-Host "   Example: https://hooks.slack.com/services/YOUR/WEBHOOK/URL" -ForegroundColor DarkGray
            Write-Host "`nPress any key to continue..." -ForegroundColor Yellow
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        "0" {
            Write-Host "`n✅ Done! Good luck with your deployment! 🚀`n" -ForegroundColor Green
            exit 0
        }
        default {
            Write-Host "`n❌ Invalid option. Please try again." -ForegroundColor Red
            Start-Sleep -Seconds 2
        }
    }
}
