#!/bin/sh
set -eu

# --- Platform detection ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# --- Version pinning (VERSION env var or latest from GitHub) ---
if [ -z "${VERSION:-}" ]; then
    echo "Detecting latest Vortex version..."
    VERSION=$(curl -sL https://api.github.com/repos/vortex-lang/vortex/releases/latest \
        | grep '"tag_name"' | cut -d'"' -f4)
    if [ -z "$VERSION" ]; then
        echo "Error: could not detect latest version from GitHub."
        echo "Set VERSION=v0.1.0 and re-run."
        exit 1
    fi
fi

# --- Download ---
URL="https://github.com/vortex-lang/vortex/releases/download/${VERSION}/vortex-${VERSION}-${OS}-${ARCH}.zip"
TMP=$(mktemp -d)
echo "Downloading Vortex ${VERSION} (${OS}/${ARCH})..."

curl -fSL "$URL" -o "$TMP/vortex.zip"
unzip -q "$TMP/vortex.zip" -d "$TMP"

# --- Install ---
mkdir -p "$HOME/.vortex/bin"
cp "$TMP/vortex-${OS}-${ARCH}" "$HOME/.vortex/bin/vortex"
chmod +x "$HOME/.vortex/bin/vortex"

# --- PATH registration ---
RC=""
case "${SHELL:-}" in
    */zsh) RC="$HOME/.zshrc" ;;
    */bash) RC="$HOME/.bashrc" ;;
    *) RC="$HOME/.profile" ;;
esac

if ! grep -q '\.vortex/bin' "$RC" 2>/dev/null; then
    printf '\nexport PATH="$HOME/.vortex/bin:$PATH"\n' >> "$RC"
fi

# --- Cleanup ---
rm -rf "$TMP"

# --- Success ---
printf '\033[32m✓ Vortex %s installed to ~/.vortex/bin/vortex\033[0m\n' "$VERSION"
printf '  Run: source %s && vortex --version\n' "$RC"
