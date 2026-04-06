package handlers

// E2E tests for the full chain of responsibility.
//
// These tests exercise the complete message-processing pipeline (from
// GenericMessageHandler through EndOfChainHandler) without any external
// dependencies such as Telegram or Discord connections.  Platform-specific
// response delivery is skipped automatically when no Service is set (Service
// zero-value), so tests can inspect the prepared Context state to verify
// correct behaviour.
//
// Video downloads are replaced by an injected mock function so that yt-dlp,
// ffmpeg, and network access are never required.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/napuu/gpsp-bot/internal/version"
)

// buildFullChain wires every handler in the same order as
// internal/chain/chain.go.  Pass a non-nil downloader to override the
// default utils.DownloadVideo call inside VideoDownloadHandler.
func buildFullChain(downloader func(url string, targetSizeInMB uint64) string) ContextHandler {
	genericMessageHandler := &GenericMessageHandler{}
	urlParsingHandler := &URLParsingHandler{}
	typingHandler := &TypingHandler{}
	videoCutArgsHandler := &VideoCutArgsHandler{}
	videoDownloadHandler := &VideoDownloadHandler{Downloader: downloader}
	videoPostprocessingHandler := &VideoPostprocessingHandler{}
	repostDetectionHandler := &RepostDetectionHandler{}
	euriborHandler := &EuriborHandler{}
	tuplillaResponseHandler := &TuplillaResponseHandler{}
	hyvaSuomiResponseHandler := &HyvaSuomiResponseHandler{}
	videoResponseHandler := &VideoResponseHandler{}
	markForNaggingHandler := &MarkForNaggingHandler{}
	markForDeletionHandler := &MarkForDeletionHandler{}
	constructTextResponseHandler := &ConstructTextResponseHandler{}
	imageResponseHandler := &ImageResponseHandler{}
	deleteMessageHandler := &DeleteMessageHandler{}
	textResponseHandler := &TextResponseHandler{}
	endOfChainHandler := &EndOfChainHandler{}

	genericMessageHandler.SetNext(urlParsingHandler)
	urlParsingHandler.SetNext(typingHandler)
	typingHandler.SetNext(videoCutArgsHandler)
	videoCutArgsHandler.SetNext(videoDownloadHandler)
	videoDownloadHandler.SetNext(videoPostprocessingHandler)
	videoPostprocessingHandler.SetNext(repostDetectionHandler)
	repostDetectionHandler.SetNext(euriborHandler)
	euriborHandler.SetNext(tuplillaResponseHandler)
	tuplillaResponseHandler.SetNext(hyvaSuomiResponseHandler)
	hyvaSuomiResponseHandler.SetNext(videoResponseHandler)
	videoResponseHandler.SetNext(markForNaggingHandler)
	markForNaggingHandler.SetNext(markForDeletionHandler)
	markForDeletionHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(imageResponseHandler)
	imageResponseHandler.SetNext(deleteMessageHandler)
	deleteMessageHandler.SetNext(textResponseHandler)
	textResponseHandler.SetNext(endOfChainHandler)

	return genericMessageHandler
}

// TestE2EPingCommand runs the full chain for a /ping message and verifies that
// the correct action is set and the response text is "pong".
func TestE2EPingCommand(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "ping")

	ctx := &Context{rawText: "/ping"}
	buildFullChain(nil).Execute(ctx)

	if ctx.action != Ping {
		t.Errorf("expected action %q, got %q", Ping, ctx.action)
	}
	if ctx.textResponse != "pong" {
		t.Errorf("expected textResponse %q, got %q", "pong", ctx.textResponse)
	}
}

// TestE2EVersionCommand runs the full chain for a /version message and
// verifies the human-readable version string is produced.
func TestE2EVersionCommand(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "version")

	ctx := &Context{rawText: "/version"}
	buildFullChain(nil).Execute(ctx)

	if ctx.action != Version {
		t.Errorf("expected action %q, got %q", Version, ctx.action)
	}
	expected := version.GetHumanReadableVersion()
	if ctx.textResponse != expected {
		t.Errorf("expected textResponse %q, got %q", expected, ctx.textResponse)
	}
}

// TestE2EUnknownText verifies that plain text (no command prefix) does not
// trigger any action or produce a response.
func TestE2EUnknownText(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "ping")

	ctx := &Context{rawText: "hello world"}
	buildFullChain(nil).Execute(ctx)

	if ctx.action != "" {
		t.Errorf("expected no action, got %q", ctx.action)
	}
	if ctx.textResponse != "" {
		t.Errorf("expected no textResponse, got %q", ctx.textResponse)
	}
}

// TestE2EVideoDownloadMocked exercises the full chain for a /dl command with
// a mocked downloader.  Because no platform is configured the video cannot be
// delivered, which puts the chain on the "nag" path: shouldNagAboutOriginalMessage
// is set and textResponse warns the user about a bad link.
func TestE2EVideoDownloadMocked(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("YTDLP_TMP_DIR", tmpDir)
	t.Setenv("REPOST_DB_DIR", filepath.Join(tmpDir, "db"))
	t.Setenv("ENABLED_FEATURES", "dl")

	// Create a small stub file that stands in for a real downloaded video.
	mockVideoPath := filepath.Join(tmpDir, "mock_video.mp4")
	if err := os.WriteFile(mockVideoPath, []byte("mock video data"), 0o644); err != nil {
		t.Fatalf("failed to create mock video file: %v", err)
	}

	mockDownloader := func(_ string, _ uint64) string {
		return mockVideoPath
	}

	ctx := &Context{rawText: "/dl https://example.com/video.mp4"}
	buildFullChain(mockDownloader).Execute(ctx)

	if ctx.action != DownloadVideo {
		t.Errorf("expected action %q, got %q", DownloadVideo, ctx.action)
	}
	if ctx.url != "https://example.com/video.mp4" {
		t.Errorf("expected url %q, got %q", "https://example.com/video.mp4", ctx.url)
	}
	// Without a live platform the video send always "fails", so the chain
	// marks the message for nagging and sets the fallback text response.
	if !ctx.shouldNagAboutOriginalMessage {
		t.Error("expected shouldNagAboutOriginalMessage=true when no platform is configured")
	}
	if !strings.Contains(ctx.textResponse, "Hyvä linkki") {
		t.Errorf("expected nag textResponse to contain %q, got %q", "Hyvä linkki", ctx.textResponse)
	}
}

// TestE2EVideoDownloadNoURL runs the full chain for a /dl command that
// contains no URL.  The mock downloader returns an empty path (simulating a
// failed download), and the chain should arrive at the same nag path as any
// other failed video delivery.
func TestE2EVideoDownloadNoURL(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("YTDLP_TMP_DIR", tmpDir)
	t.Setenv("REPOST_DB_DIR", filepath.Join(tmpDir, "db"))
	t.Setenv("ENABLED_FEATURES", "dl")

	mockDownloader := func(_ string, _ uint64) string { return "" }

	ctx := &Context{rawText: "/dl"}
	buildFullChain(mockDownloader).Execute(ctx)

	if ctx.action != DownloadVideo {
		t.Errorf("expected action %q, got %q", DownloadVideo, ctx.action)
	}
	if ctx.url != "" {
		t.Errorf("expected empty url, got %q", ctx.url)
	}
	if ctx.finalVideoPath != "" {
		t.Errorf("expected empty finalVideoPath, got %q", ctx.finalVideoPath)
	}
	// No video → nag response expected.
	if !ctx.shouldNagAboutOriginalMessage {
		t.Error("expected shouldNagAboutOriginalMessage=true when no video was downloaded")
	}
}
