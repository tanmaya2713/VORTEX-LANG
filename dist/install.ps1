[CmdletBinding()] param()
$ErrorActionPreference = "Stop"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# --- Version pinning ---
$version = [Environment]::GetEnvironmentVariable("VERSION")
if (-not $version) {
    Write-Host "Detecting latest Vortex version..."
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/vortex-lang/vortex/releases/latest"
    $version = $release.tag_name
    if (-not $version) {
        Write-Error "Could not detect latest version from GitHub."
        exit 1
    }
}

# --- Download ---
$url = "https://github.com/vortex-lang/vortex/releases/download/${version}/vortex-${version}-windows-amd64.zip"
$tmp = Join-Path $env:TEMP "vortex-install"
New-Item -ItemType Directory -Force $tmp | Out-Null
$zip = Join-Path $tmp "vortex.zip"

Write-Host "Downloading Vortex ${version} (windows/amd64)..."
Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $tmp -Force

# --- Install ---
$binDir = Join-Path $HOME ".vortex" "bin"
New-Item -ItemType Directory -Force $binDir | Out-Null
Copy-Item (Join-Path $tmp "vortex-windows-amd64.exe") (Join-Path $binDir "vortex.exe") -Force

# --- PATH registration ---
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binDir", "User")
}

# --- Cleanup ---
Remove-Item -Recurse -Force $tmp

# --- Success ---
Write-Host "✓ Vortex ${version} installed to ${binDir}" -ForegroundColor Green
Write-Host "  Restart your terminal and run: vortex --version"
