# Go Toolchain Setup (macOS)

This project requires Go >= 1.21. Below are recommended installation paths for macOS (Apple Silicon or Intel).

## 1. Homebrew Install (Recommended)
```bash
brew update
brew install go
```
Homebrew installs Go under:
```
/opt/homebrew/opt/go/libexec   # Apple Silicon
/usr/local/opt/go/libexec      # Intel
```

Ensure the Go bin directory is on PATH (add to ~/.zshrc):
```bash
echo 'export PATH="$(brew --prefix go)/libexec/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## 2. Verify Installation
```bash
which go
go version
go env GOROOT GOPATH
```
Expected: GOROOT points into the Homebrew path. GOPATH will default to `~/go` if unset.

## 3. (Optional) Set Explicit GOPATH & Update PATH
```bash
echo 'export GOPATH="$HOME/go"' >> ~/.zshrc
echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## 4. VS Code Configuration
If VS Code still cannot find Go:
1. Command Palette → "Developer: Reload Window".
2. Command Palette → "Go: Locate Configured Go Tools" (let it install tools).
3. Optionally set in `.vscode/settings.json`:
```jsonc
"go.goroot": "/opt/homebrew/opt/go/libexec"
```

## 5. Minimal Sanity Test
```bash
cat > /tmp/hello.go <<'EOF'
package main
import "fmt"
func main(){ fmt.Println("hello go") }
EOF
go run /tmp/hello.go
```
Should print: `hello go`.

## 6. Build Project Service
```bash
cd api
go build ./...
./cmd/server &
curl -s localhost:8080/health | jq
```

If `jq` missing, just `curl -s localhost:8080/health`.

## 7. Upgrading Go Later
```bash
brew upgrade go
go version
```

## 8. Alternative: asdf Version Manager
```bash
brew install asdf
asdf plugin add golang
asdf install golang 1.22.5
asdf global golang 1.22.5
echo '. $(brew --prefix asdf)/libexec/asdf.sh' >> ~/.zshrc
source ~/.zshrc
go version
```

## 9. Troubleshooting
| Symptom | Cause | Fix |
|---------|-------|-----|
| `command not found: go` | PATH not updated | Add brew go bin path to PATH |
| VS Code still complains | Window not reloaded | Reload window / set go.goroot |
| Wrong Go version | Old install in /usr/local/go | `sudo rm -rf /usr/local/go` if using brew |
| Permission denied | Shell config not sourced | `source ~/.zshrc` |

## 10. Clean Uninstall (Homebrew)
```bash
brew uninstall go
rm -rf "$GOPATH"/pkg/mod/cache
```

---
After this, run through BUILDING.md steps to continue.
