#!/usr/bin/env sh
set -e

REPO="MaximeRivest/mcptocli"
INSTALL_DIR="/usr/local/bin"
BINARY="mcptocli"

# ── detect platform ──────────────────────────────────────────────────────────

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  PLATFORM="linux" ;;
  darwin) PLATFORM="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

ASSET="${BINARY}-${PLATFORM}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

# ── download ─────────────────────────────────────────────────────────────────

echo "Downloading ${BINARY}..."
TMPFILE=$(mktemp)
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$TMPFILE" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMPFILE" "$URL"
else
  echo "Error: curl or wget is required"
  exit 1
fi

chmod +x "$TMPFILE"

# ── install ──────────────────────────────────────────────────────────────────

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
fi

# ── shell completions ────────────────────────────────────────────────────────

SHELL_NAME=$(basename "$SHELL" 2>/dev/null || echo "")
COMPLETION_LINE='source <(mcptocli completion bash)'

setup_completions() {
  case "$SHELL_NAME" in
    bash)
      RC="$HOME/.bashrc"
      LINE='source <(mcptocli completion bash)'
      ;;
    zsh)
      RC="$HOME/.zshrc"
      LINE='source <(mcptocli completion zsh)'
      ;;
    fish)
      RC="$HOME/.config/fish/config.fish"
      LINE='mcptocli completion fish | source'
      ;;
    *)
      return
      ;;
  esac

  if [ -f "$RC" ] && grep -qF "$LINE" "$RC" 2>/dev/null; then
    return
  fi

  echo "" >> "$RC"
  echo "# mcptocli shell completions" >> "$RC"
  echo "$LINE" >> "$RC"
  echo "Tab completions added to ${RC}"
}

setup_completions

# ── expose bin dir on PATH ───────────────────────────────────────────────────

EXPOSE_BIN_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/mcptocli/bin"
mkdir -p "$EXPOSE_BIN_DIR"

setup_expose_path() {
  case "$SHELL_NAME" in
    bash) RC="$HOME/.bashrc" ;;
    zsh)  RC="$HOME/.zshrc" ;;
    fish) RC="$HOME/.config/fish/config.fish" ;;
    *)    return ;;
  esac

  if [ "$SHELL_NAME" = "fish" ]; then
    LINE="fish_add_path $EXPOSE_BIN_DIR"
  else
    LINE="export PATH=\"$EXPOSE_BIN_DIR:\$PATH\""
  fi

  if [ -f "$RC" ] && grep -qF "$EXPOSE_BIN_DIR" "$RC" 2>/dev/null; then
    return
  fi

  echo "" >> "$RC"
  echo "# mcptocli exposed commands (mcp-time, mcp-notion, etc.)" >> "$RC"
  echo "$LINE" >> "$RC"
  echo "Exposed commands directory added to PATH in ${RC}"
}

setup_expose_path

# ── short alias ──────────────────────────────────────────────────────────────

setup_alias() {
  TARGET="${INSTALL_DIR}/mcp"

  if command -v mcp >/dev/null 2>&1; then
    EXISTING=$(command -v mcp)
    # If the existing 'mcp' is our own binary or shim, replace it
    case "$EXISTING" in
      "${INSTALL_DIR}"/mcp|"${INSTALL_DIR}"/mcp-*) ;;
      *)
        # Some other program owns 'mcp' — don't touch it
        return ;;
    esac
  fi

  if [ -w "$INSTALL_DIR" ]; then
    ln -sf "${INSTALL_DIR}/${BINARY}" "$TARGET"
  else
    sudo ln -sf "${INSTALL_DIR}/${BINARY}" "$TARGET"
  fi
  echo "Also available as: mcp"
}

setup_alias

# ── verify ───────────────────────────────────────────────────────────────────

VERSION=$(mcptocli version 2>/dev/null | head -n1 || echo "installed")
echo ""
echo "✓ mcptocli ${VERSION}"
echo ""
echo "Get started:"
echo "  mcptocli add time 'uvx mcp-server-time'"
echo "  mcptocli time tools"
echo ""
echo "Open a new terminal for tab completions to take effect."
