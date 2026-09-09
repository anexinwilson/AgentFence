#!/usr/bin/env bash
# install.sh - Automated 1-click installer for AgentFence on Linux/macOS
set -e

echo "============================================="
echo "  AgentFence Automated Installer (POSIX)"
echo "============================================="

INSTALL_DIR="$HOME/.agentfence/bin"
mkdir -p "$INSTALL_DIR"

BINARY_TARGET="$INSTALL_DIR/agentfence"

if [ -f "./agentfence" ]; then
    echo "[1/3] Copying prebuilt binary to $BINARY_TARGET..."
    cp "./agentfence" "$BINARY_TARGET"
    chmod +x "$BINARY_TARGET"
else
    echo "[1/3] Building AgentFence binary with Go..."
    go build -o "$BINARY_TARGET" ./cmd/agentfence
    chmod +x "$BINARY_TARGET"
fi

echo "[2/3] Checking PATH configuration..."
SHELL_PROFILE=""
if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "*/zsh" ]; then
    SHELL_PROFILE="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ] || [ "$SHELL" = "*/bash" ]; then
    SHELL_PROFILE="$HOME/.bashrc"
fi

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    if [ -n "$SHELL_PROFILE" ] && [ -f "$SHELL_PROFILE" ]; then
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_PROFILE"
        echo "  -> Added $INSTALL_DIR to $SHELL_PROFILE"
    fi
    export PATH="$INSTALL_DIR:$PATH"
fi

echo "[3/3] Verifying installation..."
"$BINARY_TARGET" version

echo "============================================="
echo "  SUCCESS! AgentFence is ready."
echo "============================================="
echo "You can now run 'agentfence init' in any project!"
