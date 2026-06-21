#!/bin/sh
set -eu

# --- Platform detection ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# --- macOS Dependency Check ---
if [ "$OS" = "darwin" ]; then
    MISSING_DEPS=""
    if ! command -v bash >/dev/null 2>&1; then MISSING_DEPS="$MISSING_DEPS bash"; fi
    if ! command -v curl >/dev/null 2>&1; then MISSING_DEPS="$MISSING_DEPS curl"; fi
    
    if [ -n "$MISSING_DEPS" ]; then
        echo "Missing macOS dependencies:$MISSING_DEPS. Attempting to install via Homebrew..."
        if command -v brew >/dev/null 2>&1; then
            brew install $MISSING_DEPS
        else
            echo "Error: Homebrew is required to install missing dependencies. Please install Homebrew or install$MISSING_DEPS manually."
            exit 1
        fi
    fi
fi

# --- Architecture Routing & Target Binaries ---
case "$OS" in
    linux)
        case "$ARCH" in
            x86_64|amd64) TARGET="linux-amd64" ;;
            aarch64|arm64) TARGET="linux-arm64" ;;
            *) echo "Unsupported architecture for Linux: $ARCH"; exit 1 ;;
        esac
        ;;
    darwin)
        case "$ARCH" in
            aarch64|arm64) TARGET="darwin-arm64" ;;
            *) echo "Unsupported architecture for macOS: $ARCH"; exit 1 ;;
        esac
        ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# --- Always fetch release metadata ---
echo "Fetching nightly Vortex build..."
RELEASE_JSON=$(curl -sL https://api.github.com/repos/tanmaya2713/VORTEX-LANG/releases/tags/nightly)

# --- Version: env override or tag ---
VERSION="${VERSION:-}"
if [ -z "$VERSION" ]; then
    if command -v jq >/dev/null 2>&1; then
        VERSION=$(echo "$RELEASE_JSON" | jq -r '.tag_name')
    else
        VERSION=$(echo "$RELEASE_JSON" | grep -m 1 '"tag_name":' | cut -d'"' -f4)
    fi
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo "Error: could not fetch nightly release from GitHub."
        exit 1
    fi
fi

# --- Download: extract browser_download_url for matching asset ---
ASSET="vortex-main-${TARGET}.zip"
if command -v jq >/dev/null 2>&1; then
    URL=$(echo "$RELEASE_JSON" | jq -r ".assets[] | select(.name == \"$ASSET\") | .browser_download_url")
else
    URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "$ASSET" | cut -d'"' -f4)
fi
if [ -z "$URL" ] || [ "$URL" = "null" ]; then
    echo "Error: could not find $ASSET in the nightly release."
    exit 1
fi

# --- Ensure unzip is available ---
if ! command -v unzip >/dev/null 2>&1; then
    echo "unzip not found. Attempting to install..."
    if   command -v apt-get >/dev/null 2>&1; then sudo apt-get update -qq && sudo apt-get install -y -qq unzip
    elif command -v pkg     >/dev/null 2>&1; then pkg install -y unzip
    elif command -v brew    >/dev/null 2>&1; then brew install unzip
    else
        echo "Error: unzip is required. Install it with your package manager, then re-run."
        exit 1
    fi
fi

# --- Ensure Python is available for compiler.py ---
PYTHON=""
if command -v python3 >/dev/null 2>&1; then
    PYTHON="python3"
elif command -v python >/dev/null 2>&1; then
    PYTHON="python"
else
    echo "Python not found. Attempting to install..."
    if   command -v apt-get >/dev/null 2>&1; then sudo apt-get update -qq && sudo apt-get install -y -qq python3
    elif command -v pkg     >/dev/null 2>&1; then pkg install -y python
    elif command -v brew    >/dev/null 2>&1; then brew install python
    else
        echo "Error: Python is required. Install python3 manually, then re-run."
        exit 1
    fi
    PYTHON="python3"
fi

TMP=$(mktemp -d)
echo "Downloading Vortex ${VERSION} (${TARGET})..."

curl -fSL "$URL" -o "$TMP/vortex.zip"
unzip -q "$TMP/vortex.zip" -d "$TMP"

# --- Install ---
mkdir -p "$HOME/.vortex/bin"
# Assuming the extracted binary matches the target name:
cp "$TMP/vortex-${TARGET}" "$HOME/.vortex/bin/vortex"
chmod +x "$HOME/.vortex/bin/vortex"

# --- Install compiler.py companion ---
if [ -f "$TMP/compiler.py" ]; then
    cp "$TMP/compiler.py" "$HOME/.vortex/bin/compiler.py"
    chmod +x "$HOME/.vortex/bin/compiler.py"
    printf '#!/bin/sh\nexec %s "$HOME/.vortex/bin/compiler.py" "$@"\n' "$PYTHON" \
      > "$HOME/.vortex/bin/vx"
    chmod +x "$HOME/.vortex/bin/vx"
fi

# --- PATH registration ---
RC=""
case "${SHELL:-}" in
    */zsh) RC="$HOME/.zshrc" ;;
    */bash)
        if [ "$OS" = "darwin" ]; then
            RC="$HOME/.bash_profile"
        else
            RC="$HOME/.bashrc"
        fi
        ;;
    *) RC="$HOME/.profile" ;;
esac

if ! grep -q '\.vortex/bin' "$RC" 2>/dev/null; then
    printf '\nexport PATH="$HOME/.vortex/bin:$PATH"\n' >> "$RC"
fi

# --- Cleanup ---
rm -rf "$TMP"

# --- Success ---
printf '\033[32m✓ Vortex %s installed to ~/.vortex/bin/\033[0m\n' "$VERSION"
printf '  Binaries: vortex (native), vx (interpreter)\n'

# --- Dynamic Post-Installation Message ---
if [ "$OS" = "darwin" ]; then
    printf '  Run: source ~/.bashrc && vortex --version\n'
else
    # Linux
    CURRENT_USER="${USER:-$(id -un)}"
    printf '  Run: source /home/%s/.bashrc && vortex --version\n' "$CURRENT_USER"
fi
