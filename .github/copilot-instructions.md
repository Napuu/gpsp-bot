# Copilot Instructions for gpsp-bot

## Repository Overview
**gpsp-bot** is a Telegram/Discord bot using Go 1.23.0 with a chain-of-responsibility pattern. Main entry: `gpsp-bot.go` (33 lines). External dependencies: yt-dlp, ffmpeg, chromium/playwright, DuckDB (requires CGO).

**Architecture**: `internal/chain/chain.go` defines message handler chain. `internal/handlers/` (22 files) process messages. `internal/platforms/` handles Telegram/Discord. `pkg/utils/` has video/euribor/LLM utilities. Features enabled via `ENABLED_FEATURES` env var (semicolon-separated): `ping`, `dl` (video download), `euribor` (interest rates), `tuplilla` (dice+LLM).

**Key Files**:
- `gpsp-bot.go` - Main entry point
- `internal/chain/chain.go` - Handler chain setup
- `internal/config/env.go` - Environment variable parsing  
- `internal/handlers/context.go` - Action definitions and context
- `internal/platforms/common.go` - Platform validation
- `.github/workflows/build.yml` - CI/CD pipeline

## Build & Test Commands

**Build**: `go build -o gpsp-bot gpsp-bot.go` (~5-10s, 68MB binary)  
**Clean build**: `go clean && go build -o gpsp-bot gpsp-bot.go`  
**Cross-compile arm64** (needs `gcc-aarch64-linux-gnu` + `g++-aarch64-linux-gnu`):  
```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
  go build -o gpsp-bot-linux-arm64 gpsp-bot.go
```
**Test**: `go test -v ./...` (1 test, <1s)  
**Coverage**: `go test -v -coverprofile=coverage.out ./...` (~7.9% coverage)  
**Format**: `go fmt ./...`  
**Vet**: `go vet ./...`  
**Dependencies**: `go mod download` or `go mod tidy`

**Run**: `ENABLED_FEATURES=ping ./gpsp-bot telegram` or `./gpsp-bot discord` (needs TELEGRAM_TOKEN or DISCORD_TOKEN)  
**Container**: `podman build -t gpsp-bot . && podman run -e ENABLED_FEATURES=ping -e TELEGRAM_TOKEN=<token> gpsp-bot telegram`

**Critical**: CGO required (DuckDB). Bot panics if ENABLED_FEATURES empty/invalid or missing platform token. No linter configs exist—use standard Go tools only.

## CI/CD Pipeline

`.github/workflows/build.yml` runs on push/PR to main:
1. **test** job: Go 1.23.0, runs `go test -v ./...`
2. **build** job (after test): Matrix for linux-amd64 and linux-arm64, sets CGO_ENABLED=1, installs cross-compiler for arm64, uploads artifacts (30d retention)

**PR requirement**: Tests must pass. arm64 build failures are typically caused by missing cross-compiler.

## Common Pitfalls

**Build**: DuckDB needs `CGO_ENABLED=1`. arm64 cross-compile needs cross-compiler toolchain. yt-dlp/ffmpeg checked at runtime only.

**Runtime**: Empty/invalid ENABLED_FEATURES causes panic. Bot creates writable temp dirs: YTDLP_TMP_DIR (`/tmp/ytdlp`), EURIBOR_GRAPH_DIR (`/tmp/euribor-graphs`), EURIBOR_CSV_DIR (`/tmp/euribor-exports`). Update yt-dlp regularly (`yt-dlp -U`).

**Testing**: Only 1 test exists (TestPingCommand). Add tests for new features. No platform integration tests—manual verification required.

## Making Changes

**New Handler**: Create `internal/handlers/*_handler.go` implementing `ContextHandler` (Execute, SetNext). Add to chain in `internal/chain/chain.go`. Add test.

**New Feature**: Define Action constant in `internal/handlers/context.go`, add to ActionMap, create handler(s), wire to chain, update README, test with `ENABLED_FEATURES=newfeature`.

**Platform Changes**: Modify `internal/platforms/{telegram,discord}_api.go` or `common.go`. Test both platforms if possible.

## Environment Variables (internal/config/env.go)

**Required**: ENABLED_FEATURES (semicolon-separated), TELEGRAM_TOKEN or DISCORD_TOKEN, MISTRAL_TOKEN (for tuplilla)  
**Optional**: YTDLP_TMP_DIR (`/tmp/ytdlp`), EURIBOR_GRAPH_DIR (`/tmp/euribor-graphs`), EURIBOR_CSV_DIR (`/tmp/euribor-exports`), PROXY_URLS (SOCKS5 proxies), ALWAYS_RE_ENCODE (false)

## Pre-commit Checklist

1. `go build` - catch compilation errors
2. `go test -v ./...` - verify tests pass
3. `go fmt ./...` - format code
4. For cross-compile changes: verify amd64 and arm64 builds
5. Update README.md if user-facing changes

**These instructions are verified and up-to-date.** Search for additional context only when needed.
