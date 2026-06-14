[CmdletBinding()] param()
$ErrorActionPreference = "Stop"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# --- Architecture detection ---
$arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
if ($arch -ne 'amd64' -and $arch -ne 'arm64') {
    Write-Error "Unsupported architecture: $arch (only amd64 and arm64 are supported)"
    exit 1
}
$assetName = "vortex-main-windows-$arch.zip"

# --- Ensure Python is available for compiler.py ---
$python = $null
if (Get-Command "python3" -ErrorAction SilentlyContinue) {
    $python = "python3"
} elseif (Get-Command "python" -ErrorAction SilentlyContinue) {
    $python = "python"
} else {
    Write-Host "Python not found. Attempting to install..."
    try {
        winget install -e --id Python.Python.3.12 --accept-package-agreements --accept-source-agreements
        $python = "python"
        $env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")
    } catch {
        Write-Host "winget failed. Downloading Python installer from python.org..."
        $tmp = Join-Path $env:TEMP "vortex-install"
        New-Item -ItemType Directory -Force "$tmp" | Out-Null
        $pyInstaller = Join-Path $tmp "python-installer.exe"
        try {
            Invoke-WebRequest -Uri "https://www.python.org/ftp/python/3.12.3/python-3.12.3-amd64.exe" -OutFile $pyInstaller
            Write-Host "Running Python installer silently..."
            Start-Process -Wait -FilePath $pyInstaller -ArgumentList "/quiet InstallAllUsers=1 PrependPath=1"
            $python = "python"
            $env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")
        } catch {
            Write-Error "Failed to install Python automatically. Install it from https://python.org and re-run."
            exit 1
        }
    }
}

# --- Always fetch release metadata ---
Write-Host "Fetching nightly Vortex build..."
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/tanmaya2713/VORTEX-LANG/releases/tags/nightly"

# --- Version: env override or tag ---
$version = [Environment]::GetEnvironmentVariable("VERSION")
if (-not $version) {
    $version = $release.tag_name
    if (-not $version) {
        Write-Error "Could not detect version from the nightly release."
        exit 1
    }
}

# --- Download ---
$asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1
if (-not $asset) {
    Write-Error "Could not find $assetName in the nightly release."
    exit 1
}
$url = $asset.browser_download_url
$tmp = Join-Path $env:TEMP "vortex-install"
New-Item -ItemType Directory -Force "$tmp" | Out-Null
$zip = Join-Path $tmp "vortex.zip"

Write-Host "Downloading Vortex ${version} (windows/${arch})..."
Invoke-WebRequest -Uri $url -OutFile "$zip"
try {
    Expand-Archive -Path "$zip" -DestinationPath "$tmp" -Force
} catch {
    Write-Host "Expand-Archive failed, falling back to System.IO.Compression.ZipFile..."
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::ExtractToDirectory($zip, $tmp)
}

# --- Install ---
$binDir = Join-Path (Join-Path $HOME ".vortex") "bin"
New-Item -ItemType Directory -Force "$binDir" | Out-Null
Copy-Item "$(Join-Path $tmp "vortex-windows-${arch}.exe")" "$(Join-Path $binDir "vortex.exe")" -Force

# --- Install compiler.py companion ---
$compilerPy = Join-Path $tmp "compiler.py"
if (Test-Path $compilerPy) {
    Copy-Item $compilerPy "$(Join-Path $binDir "compiler.py")" -Force
    $wrapper = "@echo off`r`n`"$python`" `"%~dp0compiler.py`" %*"
    Set-Content -Path "$(Join-Path $binDir "vx.bat")" -Value $wrapper
}

# --- PATH registration ---
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binDir", "User")
}

# --- Cleanup ---
Remove-Item -Recurse -Force "$tmp"

# --- Success ---
Write-Host "✓ Vortex ${version} installed to ${binDir}" -ForegroundColor Green
Write-Host "  Binaries: vortex.exe (native), vx.bat (interpreter)"
Write-Host "  Restart your terminal and run: vortex --version"
