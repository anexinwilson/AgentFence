# install.ps1 - Automated 1-click installer for AgentFence on Windows
$ErrorActionPreference = "Stop"

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "  AgentFence Automated Installer (Windows)" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# 1. Target installation directory in user profile
$installDir = "$HOME\.agentfence\bin"
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# 2. Build or copy static binary
$binaryTarget = "$installDir\agentfence.exe"
$localBinary = ".\agentfence.exe"

if (Test-Path $localBinary) {
    Write-Host "`n[1/3] Copying prebuilt binary to $binaryTarget..." -ForegroundColor Cyan
    Copy-Item $localBinary $binaryTarget -Force
} else {
    Write-Host "`n[1/3] Building AgentFence binary with Go..." -ForegroundColor Cyan
    go build -o $binaryTarget ./cmd/agentfence
}

# 3. Add to User PATH if not already present
Write-Host "[2/3] Configuring User Environment PATH..." -ForegroundColor Cyan
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    $newUserPath = "$userPath;$installDir"
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    Write-Host "  -> Added $installDir to User PATH." -ForegroundColor Green
} else {
    Write-Host "  -> $installDir is already in User PATH." -ForegroundColor Gray
}

# Update current session PATH
if ($env:Path -notlike "*$installDir*") {
    $env:Path = "$env:Path;$installDir"
}

# 4. Verify installation
Write-Host "[3/3] Verifying agentfence installation..." -ForegroundColor Cyan
try {
    $versionOutput = & "$binaryTarget" version
    Write-Host "`n=============================================" -ForegroundColor Green
    Write-Host "  SUCCESS! $versionOutput is installed." -ForegroundColor Green
    Write-Host "  Binary: $binaryTarget" -ForegroundColor Green
    Write-Host "=============================================" -ForegroundColor Green
    Write-Host "`nYou can now run 'agentfence init' in any project to protect it!`n" -ForegroundColor White
} catch {
    Write-Host "Verification failed: $_" -ForegroundColor Red
}
