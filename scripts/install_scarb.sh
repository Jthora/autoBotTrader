#!/usr/bin/env bash
set -euo pipefail

echo "[install_scarb] Detecting platform..."
ARCH=$(uname -m)
if [[ "$ARCH" == "arm64" ]]; then
  TARCH="aarch64"
else
  TARCH="x86_64"
fi

if command -v brew >/dev/null 2>&1; then
  echo "[install_scarb] Using Homebrew"
  brew tap software-mansion/scarb || true
  brew install scarb || brew upgrade scarb || true
else
  echo "[install_scarb] Homebrew not found, using direct tarball"
  URL="https://github.com/software-mansion/scarb/releases/latest/download/scarb-${TARCH}-apple-darwin.tar.gz"
  echo "[install_scarb] Downloading $URL"
  curl -L "$URL" -o scarb.tar.gz
  tar -xzf scarb.tar.gz
  sudo mv scarb /usr/local/bin/
  rm scarb.tar.gz
fi

echo "[install_scarb] Version: $(scarb --version)"
echo "[install_scarb] Done. Run 'cd contracts && scarb build' next."
