$ErrorActionPreference = "Stop"

$repo = "MaximeRivest/mcptocli"
$asset = "mcptocli-windows-amd64.exe"
$url = "https://github.com/$repo/releases/latest/download/$asset"
$installDir = "$env:USERPROFILE\bin"
$binary = "$installDir\mcptocli.exe"

# ── download ─────────────────────────────────────────────────────────────────

Write-Host "Downloading mcptocli..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Invoke-WebRequest -Uri $url -OutFile $binary

# ── add to PATH ──────────────────────────────────────────────────────────────

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:PATH += ";$installDir"
    Write-Host "Added $installDir to PATH"
}

# ── expose bin dir on PATH ───────────────────────────────────────────────────

$exposeBinDir = "$env:LOCALAPPDATA\mcptocli\bin"
New-Item -ItemType Directory -Force -Path $exposeBinDir | Out-Null

$currentPath2 = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath2 -notlike "*$exposeBinDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath2;$exposeBinDir", "User")
    $env:PATH += ";$exposeBinDir"
    Write-Host "Added $exposeBinDir to PATH (exposed commands like mcp-time)"
}

# ── verify ───────────────────────────────────────────────────────────────────

$version = & $binary version 2>$null | Select-Object -First 1
Write-Host ""
Write-Host "✓ mcptocli $version"
Write-Host ""
Write-Host "Get started:"
Write-Host "  mcptocli add time 'uvx mcp-server-time'"
Write-Host "  mcptocli time tools"
Write-Host ""
Write-Host "Open a new terminal for PATH changes to take effect."
