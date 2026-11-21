# Instructions for AI Agents for gpsp-bot

## Repository Overview
**gpsp-bot** is a Telegram/Discord bot using Go 1.27.0 with a chain-of-responsibility pattern. Main entry: `gpsp-bot.go`. External dependencies: yt-dlp, ffmpeg, chromium/playwright, DuckDB (requires CGO).

**Architecture**: `internal/chain/chain.go` defines message handler chain. `internal/handlers/` process messages (e.g., `stats_handler.go`, `repost_detection_handler.go`, `euribor_handler.go`, `tuplilla_response_handler.go`, `video_download_handler.go`, `video_stats_handler.go`). `internal/platforms/` handles Telegram/Discord. `pkg/utils/` has video/euribor/LLM utilities. Features enabled via `ENABLED_FEATURES` env var (semicolon-separated): `ping`, `dl` (video download), `euribor` (interest rates), `tuplilla` (dice+LLM), `stats`, `version`, `happener` (YLE teksti-TV page as image).

**Key Files**:
- `gpsp-bot.go` - Main entry point
- `internal/chain/chain.go` - Handler chain setup
- `internal/config/env.go` - Environment variable parsing  
- `internal/doctor/doctor.go` - Doctor command for diagnostics
- `internal/handlers/context.go` - Action definitions and context
- `internal/handlers/stats_handler.go` - Stats feature handler
- `internal/handlers/repost_detection_handler.go` - Repost detection handler
- `internal/handlers/euribor_handler.go` - Euribor rates handler
- `internal/handlers/happener_handler.go` - Teksti-TV page handler (`/häppener`, `/happener`); page fetching lives in `pkg/utils/teletext.go`
- `internal/handlers/tuplilla_response_handler.go` - Tuplilla (dice+LLM) handler
- `internal/handlers/video_download_handler.go` - Video download handler
- `internal/handlers/video_stats_handler.go` - Records bot video posts
- `internal/platforms/common.go` - Platform validation
- `.github/workflows/build.yml` - CI/CD pipeline

## Build & Test Commands

**Build**: `go build -o gpsp-bot gpsp-bot.go`  
**Clean build**: `go clean && go build -o gpsp-bot gpsp-bot.go`  
**Cross-compile static binaries** (output runs on any Linux regardless of the host glibc — this is what CI releases use; a dynamic binary built on ubuntu-latest requires that runner's glibc and broke deployments on older Ubuntu, the GLIBC_2.43 incident):
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc CXX=g++ \
  go build -ldflags "-linkmode external -extldflags=-static" -o gpsp-bot-linux-amd64 gpsp-bot.go
# arm64 needs gcc-aarch64-linux-gnu + g++-aarch64-linux-gnu:
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
  go build -ldflags "-linkmode external -extldflags=-static" -o gpsp-bot-linux-arm64 gpsp-bot.go
```
Static glibc links fine with DuckDB's prebuilt archives (they need `-lstdc++`, which the runner's g++ provides). Expect harmless link warnings about dlopen/getaddrinfo in static binaries; DNS via /etc/resolv.conf still works. Verify with `file` (must say "statically linked") and `readelf -d` (must have no NEEDED entries).
The `muslcompat` tag compiles `internal/muslcompat`, which shims the few glibc-only symbols (`__res_init`, `backtrace`, `backtrace_symbols`, `malloc_trim`) that the prebuilt DuckDB archives reference. Never combine it with a glibc toolchain — the shims would interpose libc symbols.
**Test**: `go test -v ./...`  
**Coverage**: `go test -v -coverprofile=coverage.out ./...`  
**Format**: `go fmt ./...`  
**Vet**: `go vet ./...`  
**Dead code**: `go run golang.org/x/tools/cmd/deadcode@v0.49.0 -test ./...` (should print nothing; any output is an unreachable function)  
**Dependencies**: `go mod download` or `go mod tidy`

**Run**: `ENABLED_FEATURES=ping ./gpsp-bot telegram` or `./gpsp-bot discord` (needs TELEGRAM_TOKEN or DISCORD_TOKEN). Use `./gpsp-bot doctor` to run diagnostics and validate configuration.  
**Container**: `podman build -t gpsp-bot . && podman run -e ENABLED_FEATURES=ping -e TELEGRAM_TOKEN=<token> gpsp-bot telegram`

**Critical**: CGO required (DuckDB). Bot panics if ENABLED_FEATURES empty/invalid or missing platform token. No linter configs exist—use standard Go tools only.

## CI/CD Pipeline

`.github/workflows/build.yml` runs on push/PR to main, `workflow_dispatch`, or `workflow_call`:
1. **test** job: Go 1.27.1, runs `go test -v ./...` then `go run golang.org/x/tools/cmd/deadcode@v0.49.0 -test ./...` (fails the build on any unreachable function)
2. **build** job (after test): Matrix for linux-amd64 and linux-arm64, installs the arm64 cross-compiler via apt, and builds **statically linked** binaries (`-linkmode external -extldflags=-static`) so artifacts run on any Linux regardless of the host glibc. Verifies each binary is statically linked with no dynamic NEEDED entries and smoke-tests the amd64 one in an `alpine:3.22` container. Embeds version via `-ldflags`, uploads artifacts (configurable retention days, default 30)

**Inputs**:
- `version`: Version to embed in the binary
- `retention_days`: Number of days to retain artifacts (default: 30)

**PR requirement**: Tests must pass.

## Common Pitfalls

**Build**: DuckDB needs `CGO_ENABLED=1`. Release binaries must be **statically linked** (see above) — a dynamic binary built on ubuntu-latest requires that runner's glibc and will fail on older Ubuntu hosts (this caused a GLIBC_2.43 incident when ubuntu-latest moved to 26.04). yt-dlp/ffmpeg checked at runtime only.

**Runtime**: Empty/invalid ENABLED_FEATURES causes panic. Bot creates writable temp dirs: YTDLP_TMP_DIR (`/tmp/ytdlp`), EURIBOR_GRAPH_DIR (`/tmp/euribor-graphs`), EURIBOR_CSV_DIR (`/tmp/euribor-exports`), TELETEXT_IMAGE_DIR (`/tmp/teletext`). Update yt-dlp regularly (`yt-dlp -U`).

**Testing**: Add tests for new features. No platform integration tests—manual verification required.

## Making Changes

**Commit messages**: Prefer [Scoped Commits](https://scopedcommits.com/) (`<scope>: <description>`). This is guidance only—there is no CI/hook enforcement.

**New Handler**: Create `internal/handlers/*_handler.go` implementing `ContextHandler` (Execute, SetNext). Add to chain in `internal/chain/chain.go`. Add test. 

**New Feature**: Define Action constant in `internal/handlers/context.go`, add to ActionMap, create handler(s), wire to chain, update README, test with `ENABLED_FEATURES=newfeature`. 

**Platform Changes**: Modify `internal/platforms/{telegram,discord}_api.go` or `common.go`.

## Environment Variables (internal/config/env.go)

**Required**: ENABLED_FEATURES (semicolon-separated), TELEGRAM_TOKEN or DISCORD_TOKEN, MISTRAL_TOKEN (for tuplilla)  
**Optional**: YTDLP_TMP_DIR (`/tmp/ytdlp`), EURIBOR_GRAPH_DIR (`/tmp/euribor-graphs`), EURIBOR_CSV_DIR (`/tmp/euribor-exports`), REPOST_DB_DIR (`/tmp/repost-db`), TELETEXT_IMAGE_DIR (`/tmp/teletext`), DATABASE_FILE, PROXY_URLS (SOCKS5 proxies), ALWAYS_RE_ENCODE (false)

## Pre-commit Checklist

1. `go build` - catch compilation errors
2. `go test -v ./...` - verify tests pass
3. `go fmt ./...` - format code
4. For cross-compile changes: verify amd64 and arm64 builds
5. Update README.md if user-facing changes

**These instructions are verified and up-to-date.** Refer to the source code for ground truth when necessary.
