# Copilot Instructions for gpsp-bot

## Repository Overview

**gpsp-bot** is a General Purpose S-Posting Bot that provides useful and entertaining commands for group chats on Telegram and Discord platforms. The bot uses a chain-of-responsibility pattern to process messages through various handlers.

**Repository Stats:**
- Language: Go 1.23.0
- Size: ~40 Go source files
- Main entry point: `gpsp-bot.go`
- External dependencies: yt-dlp, ffmpeg, chromium (via playwright)

## Project Architecture

### Directory Structure
```
.
├── gpsp-bot.go              # Main entry point
├── go.mod, go.sum           # Go module files
├── Containerfile            # Container build definition
├── README.md                # User documentation
├── CHAIN_ARCHITECTURE.md    # Handler chain diagram
├── internal/
│   ├── chain/              # Chain-of-responsibility implementation
│   │   └── chain.go        # Handler chain setup
│   ├── config/             # Configuration management
│   │   └── env.go          # Environment variable parsing
│   ├── handlers/           # Message processing handlers (22 files)
│   │   ├── context.go      # Shared context for handlers
│   │   ├── *_handler.go    # Individual handler implementations
│   │   └── handlers_test.go # Handler tests
│   └── platforms/          # Platform-specific integrations
│       ├── common.go       # Shared platform code
│       ├── telegram_api.go # Telegram bot integration
│       └── discord_api.go  # Discord bot integration
└── pkg/
    └── utils/              # Utility functions
        ├── video_downloader.go  # yt-dlp wrapper
        ├── euribor.go           # Euribor data fetching (uses playwright)
        ├── llm_util.go          # Mistral API integration
        └── misc.go              # Miscellaneous utilities
```

### Key Architectural Elements

1. **Handler Chain Pattern**: Messages flow through a chain of handlers defined in `internal/chain/chain.go`. Each handler processes the message and passes it to the next handler. See `CHAIN_ARCHITECTURE.md` for visual representation.

2. **Platform Abstraction**: `internal/platforms/` contains platform-specific code for Telegram and Discord. Business logic in handlers is reused across platforms.

3. **Feature Flags**: Commands are enabled via the `ENABLED_FEATURES` environment variable (semicolon-separated list like `ping;dl;euribor`). Available features defined in `internal/handlers/context.go`:
   - `ping` - Simple ping/pong command
   - `dl` - Download videos using yt-dlp
   - `euribor` - Fetch Euribor interest rates
   - `tuplilla` - Dice roll with LLM-generated responses

4. **External Dependencies**:
   - **yt-dlp**: Required for video downloading (dl command)
   - **ffmpeg**: Required for video processing/re-encoding
   - **chromium/playwright**: Required for Euribor data scraping
   - **DuckDB**: Uses CGO, requires C compiler

## Building and Testing

### Prerequisites
- Go 1.23.0 or newer (go.mod specifies 1.23.0)
- CGO-enabled compiler (required for DuckDB dependency)
- For cross-compilation to arm64: `gcc-aarch64-linux-gnu` and `g++-aarch64-linux-gnu`

### Build Commands

**Standard build (amd64):**
```bash
go build -o gpsp-bot gpsp-bot.go
```
Build time: ~5-10 seconds. Output binary is ~68MB.

**Clean build:**
```bash
go clean
go build -o gpsp-bot gpsp-bot.go
```

**Cross-compilation (arm64):**
Cross-compilation for arm64 requires cross-compiler toolchain:
```bash
# Install cross-compiler first
sudo apt-get update
sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu

# Build
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
  go build -o gpsp-bot-linux-arm64 gpsp-bot.go
```

**Note**: Cross-compilation will fail without the proper cross-compiler installed due to CGO dependency on DuckDB.

### Dependency Management

**Download dependencies:**
```bash
go mod download
```

**Tidy dependencies (add missing and remove unused):**
```bash
go mod tidy
```
This may download additional transitive dependencies (10-20 packages).

### Testing

**Run all tests:**
```bash
go test -v ./...
```
Expected output: 1 test in `internal/handlers/handlers_test.go` (TestPingCommand). Test runtime: <1 second.

**Run tests with coverage:**
```bash
go test -v -coverprofile=coverage.out ./...
```
Coverage report saved to `coverage.out` (excluded from git via .gitignore).

**Current test coverage**: ~7.9% (only handlers have tests).

### Code Quality

**Format code:**
```bash
go fmt ./...
```
Should return no output if all code is properly formatted.

**Vet code (static analysis):**
```bash
go vet ./...
```
Should complete with no errors.

**Note**: No linter configuration files (`.golangci.yml`, etc.) exist in the repository. Standard Go tools only.

### Running the Bot

**Basic usage:**
```bash
# Requires platform argument
./gpsp-bot <telegram|discord>
```

**With features enabled (will fail without valid tokens):**
```bash
ENABLED_FEATURES=ping TELEGRAM_TOKEN=<token> ./gpsp-bot telegram
ENABLED_FEATURES=ping DISCORD_TOKEN=<token> ./gpsp-bot discord
```

**Using go run:**
```bash
ENABLED_FEATURES=ping go run gpsp-bot.go telegram
```

**Important**: The bot will panic at startup if:
1. `ENABLED_FEATURES` is empty or contains invalid action names
2. Platform argument is missing
3. Required token for platform is missing (TELEGRAM_TOKEN or DISCORD_TOKEN)

### Container Build

```bash
# Build container
podman build -t gpsp-bot .
# or: docker build -t gpsp-bot .

# Run with mounted yt-dlp (allows updates without rebuilding)
podman run -v /usr/bin/yt-dlp:/usr/bin/yt-dlp:z \
  -e ENABLED_FEATURES="ping;dl" \
  -e TELEGRAM_TOKEN=<token> gpsp-bot telegram
```

The Containerfile installs Go 1.23.6, ffmpeg, yt-dlp, and playwright with chromium.

## CI/CD Pipeline

### GitHub Actions Workflow (`.github/workflows/build.yml`)

**Triggered on:**
- Push to main branch
- Pull requests to main branch  
- Manual workflow dispatch

**Jobs:**

1. **test**: Runs on ubuntu-latest
   - Sets up Go 1.23.0
   - Runs `go test -v ./...`
   - Must pass before build job runs

2. **build**: Runs on ubuntu-latest (needs test job)
   - Matrix build for linux-amd64 and linux-arm64
   - For arm64: Installs `gcc-aarch64-linux-gnu` and `g++-aarch64-linux-gnu`
   - Sets CGO_ENABLED=1 and appropriate CC/CXX compilers
   - Uploads artifacts with 30-day retention

**Critical for PRs**: Tests must pass. Build failures on arm64 indicate missing cross-compiler setup.

## Common Pitfalls and Workarounds

### Build Issues

1. **CGO errors**: DuckDB requires CGO. Always ensure `CGO_ENABLED=1` for builds.

2. **Cross-compilation failures**: arm64 builds fail without cross-compiler. See "Build Commands" section for proper setup.

3. **Missing external tools**: Runtime commands (yt-dlp, ffmpeg) are not checked at build time. The bot will fail at runtime if these are missing and corresponding features are enabled.

### Runtime Issues

1. **Empty ENABLED_FEATURES**: Bot panics with "Action '' does not exist". Always set this variable, even if just to "ping".

2. **Temporary directories**: Bot creates temp dirs at startup:
   - `YTDLP_TMP_DIR` (default: `/tmp/ytdlp`)
   - `EURIBOR_GRAPH_DIR` (default: `/tmp/euribor-graphs`)
   - `EURIBOR_CSV_DIR` (default: `/tmp/euribor-exports`)
   These must be writable. Bot will panic if creation fails.

3. **yt-dlp updates**: Video downloads may fail with outdated yt-dlp. Run `yt-dlp -U` regularly to update.

### Testing Issues

1. **Limited test coverage**: Only one test exists (`TestPingCommand`). When adding features, add corresponding tests following the existing pattern.

2. **Platform integration tests**: No integration tests for Telegram/Discord exist. Changes to platform code require manual verification.

## Making Changes

### Adding a New Handler

1. Create handler file in `internal/handlers/` following naming pattern `*_handler.go`
2. Implement `ContextHandler` interface (Execute and SetNext methods)
3. Add handler to chain in `internal/chain/chain.go` at appropriate position
4. If adding new command, add to `ActionMap` in `internal/handlers/context.go`
5. Add test in `internal/handlers/handlers_test.go`

### Adding a New Feature/Command

1. Define action constant in `internal/handlers/context.go` (e.g., `NewFeature Action = "newfeature"`)
2. Add to `ActionMap` with description
3. Create handler(s) to process the feature
4. Wire into chain in `internal/chain/chain.go`
5. Update README.md with feature documentation
6. Test with `ENABLED_FEATURES=newfeature`

### Modifying Platform Integration

Platform-specific code in `internal/platforms/`:
- `telegram_api.go` - Telegram bot setup and message conversion
- `discord_api.go` - Discord bot setup and message conversion
- `common.go` - Shared validation and initialization

Changes here affect how messages are received and responses are sent. Test with both platforms if possible.

## Environment Variables

All environment variables parsed in `internal/config/env.go`:

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| TELEGRAM_TOKEN | For Telegram | none | Telegram bot token |
| DISCORD_TOKEN | For Discord | none | Discord bot token |
| MISTRAL_TOKEN | For tuplilla | none | Mistral AI API token |
| ENABLED_FEATURES | Yes | none | Semicolon-separated feature list |
| YTDLP_TMP_DIR | No | `/tmp/ytdlp` | Video download directory |
| EURIBOR_GRAPH_DIR | No | `/tmp/euribor-graphs` | Euribor graph output |
| EURIBOR_CSV_DIR | No | `/tmp/euribor-exports` | Euribor CSV exports |
| PROXY_URLS | No | none | Semicolon-separated SOCKS5 proxies for yt-dlp |
| ALWAYS_RE_ENCODE | No | false | Force H.264 re-encoding (true/yes/1) |

## Files in Repository Root

- `gpsp-bot.go` - Main entry point (33 lines)
- `go.mod`, `go.sum` - Go module definition and checksums
- `README.md` - User-facing documentation
- `CHAIN_ARCHITECTURE.md` - Mermaid diagram of handler chain
- `Containerfile` - Container image definition
- `.gitignore` - Excludes target/, gpsp-bot binaries, coverage.out
- `.containerignore` - Whitelist approach for container context
- `.github/workflows/build.yml` - CI/CD pipeline

## Quick Reference

**Trust these instructions.** Only search for additional information if these instructions are incomplete or incorrect.

When making changes:
1. Run `go build` early to catch compilation errors
2. Run `go test -v ./...` to verify existing tests pass
3. Run `go fmt ./...` before committing
4. For cross-compilation changes, verify both amd64 and arm64 builds
5. Check that CI workflow still passes
6. Update README.md if adding user-facing features
