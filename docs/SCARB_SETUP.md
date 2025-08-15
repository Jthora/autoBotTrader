## Scarb & Cairo Toolchain Setup (macOS)

This project’s Cairo contracts compile with Scarb (Cairo 2 toolchain). Follow the steps below to install Scarb and supporting Starknet tools on macOS (zsh).

### 1. Prerequisites

Ensure you have:

- Homebrew (recommended) — https://brew.sh
- Rust toolchain (only if building from source) — `curl https://sh.rustup.rs -sSf | sh`

### 2. Install via Homebrew (Recommended)

```bash
brew update
brew tap software-mansion/scarb
brew install scarb
```

If you previously installed, upgrade with:

```bash
brew upgrade scarb
```

### 3. Alternative: Direct Installer Script

```bash
curl -L https://github.com/software-mansion/scarb/releases/latest/download/scarb-x86_64-apple-darwin.tar.gz -o scarb.tar.gz
tar -xzf scarb.tar.gz
sudo mv scarb /usr/local/bin/
rm scarb.tar.gz
```

(Adjust architecture if on Apple Silicon: replace `x86_64` with `aarch64`).

### 4. Verify

```bash
scarb --version
```

Expected: prints version (e.g. `scarb 2.x.x`).

### 5. Create & Build Test Project (Optional Sanity Check)

```bash
mkdir -p /tmp/scarb_check && cd /tmp/scarb_check
scarb new demo
cd demo
scarb build
```

Should finish with `Finished release target(s)`.

### 6. Install Starkli (for deployment later)

```bash
cargo install starkli --locked
```

Add cargo bin directory to PATH (if not already):

```bash
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### 7. Project Build

From repository root:

```bash
cd contracts
scarb build
```

Artifacts land under `target/dev/`.

### 8. Common Issues

| Symptom                                  | Fix                                                                                                              |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `command not found: scarb` after install | Ensure Homebrew prefix in PATH: `echo $PATH`; add `eval "$(/opt/homebrew/bin/brew shellenv)"` for Apple Silicon. |
| Mismatch Cairo version                   | Run `brew upgrade scarb` or remove old binary in `/usr/local/bin` before reinstall.                              |
| Permission denied moving binary          | Use `sudo mv` or choose a directory in PATH you own (e.g. `$HOME/.local/bin`).                                   |

### 9. Optional: Scripted Install

Run provided helper script:

```bash
./scripts/install_scarb.sh
```

### 10. Next Step

After installation, re-run:

```bash
cd contracts
scarb build
```

Then we will add Cairo tests and move `contract-skeleton` toward COMPLETE.
