<#
.SYNOPSIS
    One-liner installer for ztictl on Windows.

.DESCRIPTION
    Downloads and installs ztictl to %USERPROFILE%\Tools and adds to PATH.

.EXAMPLE
    irm https://github.com/zsoftly/ztiaws/releases/latest/download/install.ps1 | iex
#>

# PSScriptAnalyzer: Write-Host is intentional for colored installer output
[Diagnostics.CodeAnalysis.SuppressMessageAttribute('PSAvoidUsingWriteHost', '')]
param()

$ErrorActionPreference = "Stop"

$repo = "zsoftly/ztiaws"
$installDir = "$env:USERPROFILE\Tools"
$binaryName = "ztictl.exe"

# Detect architecture (64-bit only)
if (-not [Environment]::Is64BitOperatingSystem) {
    Write-Host "  [ERROR] 32-bit Windows is not supported" -ForegroundColor Red
    exit 1
}

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

$assetName = "ztictl-windows-$arch.exe"
$downloadUrl = "https://github.com/$repo/releases/latest/download/$assetName"

Write-Host ""
Write-Host "  ztictl installer" -ForegroundColor Cyan
Write-Host "  ----------------" -ForegroundColor Cyan
Write-Host ""

# Create install directory
if (-not (Test-Path $installDir)) {
    Write-Host "  Creating $installDir..." -ForegroundColor Gray
    New-Item -ItemType Directory -Force $installDir | Out-Null
}

# Download binary
$destPath = Join-Path $installDir $binaryName
Write-Host "  Downloading $assetName..." -ForegroundColor Gray

try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $downloadUrl -OutFile $destPath -UseBasicParsing
} catch {
    Write-Host "  [ERROR] Download failed: $_" -ForegroundColor Red
    exit 1
}

# Add to PATH if not already there (persists across sessions)
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "  Adding to PATH..." -ForegroundColor Gray
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$installDir", "User")
}

# Verify
try {
    $version = & $destPath --version 2>&1
    Write-Host ""
    Write-Host "  [OK] Installed successfully!" -ForegroundColor Green
    Write-Host "  $version" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Usage:" -ForegroundColor Yellow
    Write-Host "    ztictl auth login"
    Write-Host "    ztictl ssm list --region cac1"
    Write-Host "    ztictl --help"
    Write-Host ""
    Write-Host "  [!] Restart your terminal for PATH changes" -ForegroundColor Yellow
    Write-Host ""
} catch {
    Write-Host "  [OK] Installed to $destPath" -ForegroundColor Green
    Write-Host "  [!] Restart terminal, then: ztictl --version" -ForegroundColor Yellow
}
