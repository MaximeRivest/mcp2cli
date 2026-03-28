$ErrorActionPreference = "Stop"

$repo = "MaximeRivest/mcp2cli"
$asset = "mcp2cli-windows-amd64.exe"
$url = "https://github.com/$repo/releases/latest/download/$asset"
$installDir = "$env:USERPROFILE\bin"
$binary = "$installDir\mcp2cli.exe"

# ── download ─────────────────────────────────────────────────────────────────

Write-Host "Downloading mcp2cli..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Invoke-WebRequest -Uri $url -OutFile $binary

# ── add to PATH ──────────────────────────────────────────────────────────────

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:PATH += ";$installDir"
    Write-Host "Added $installDir to PATH"
}

# ── verify ───────────────────────────────────────────────────────────────────

$version = & $binary version 2>$null | Select-Object -First 1
Write-Host ""
Write-Host "✓ mcp2cli $version"
Write-Host ""
Write-Host "Get started:"
Write-Host "  mcp2cli add myserver 'npx -y @modelcontextprotocol/server-filesystem C:\tmp'"
Write-Host "  mcp2cli myserver tools"
Write-Host ""
Write-Host "Open a new terminal for PATH changes to take effect."
