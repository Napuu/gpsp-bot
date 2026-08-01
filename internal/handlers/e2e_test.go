package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

// findProjectRoot returns the project root by walking up from this test file.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	// filename is .../internal/handlers/e2e_test.go → go up 3 levels
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// newVideoFileServer starts an HTTP server that serves testdata/sample.mp4.
func newVideoFileServer(t *testing.T) *httptest.Server {
	t.Helper()
	root := findProjectRoot(t)
	samplePath := filepath.Join(root, "testdata", "sample.mp4")
	if _, err := os.Stat(samplePath); err != nil {
		t.Fatalf("sample.mp4 not found at %s: %v", samplePath, err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sample.mp4":
			http.ServeFile(w, r, samplePath)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// buildFullChain replicates chain.NewChainOfResponsibility() to avoid circular imports.
func buildFullChain() ContextHandler {
	onTextHandler := &OnTextHandler{}
	genericMessageHandler := &GenericMessageHandler{}
	urlParsingHandler := &URLParsingHandler{}
	typingHandler := &TypingHandler{}
	videoCutArgsHandler := &VideoCutArgsHandler{}
	videoDownloadHandler := &VideoDownloadHandler{}
	videoPostprocessingHandler := &VideoPostprocessingHandler{}
	repostDetectionHandler := &RepostDetectionHandler{}
	euriborHandler := &EuriborHandler{}
	statsHandler := &StatsHandler{}
	tuplillaResponseHandler := &TuplillaResponseHandler{}
	hyvaSuomiResponseHandler := &HyvaSuomiResponseHandler{}
	videoResponseHandler := &VideoResponseHandler{}
	videoStatsHandler := &VideoStatsHandler{}
	markForNaggingHandler := &MarkForNaggingHandler{}
	markForDeletionHandler := &MarkForDeletionHandler{}
	constructTextResponseHandler := &ConstructTextResponseHandler{}
	imageResponseHandler := &ImageResponseHandler{}
	deleteMessageHandler := &DeleteMessageHandler{}
	textResponseHandler := &TextResponseHandler{}
	endOfChainHandler := &EndOfChainHandler{}

	onTextHandler.SetNext(genericMessageHandler)
	genericMessageHandler.SetNext(urlParsingHandler)
	urlParsingHandler.SetNext(typingHandler)
	typingHandler.SetNext(videoCutArgsHandler)
	videoCutArgsHandler.SetNext(videoDownloadHandler)
	videoDownloadHandler.SetNext(videoPostprocessingHandler)
	videoPostprocessingHandler.SetNext(repostDetectionHandler)
	repostDetectionHandler.SetNext(euriborHandler)
	euriborHandler.SetNext(statsHandler)
	statsHandler.SetNext(tuplillaResponseHandler)
	tuplillaResponseHandler.SetNext(hyvaSuomiResponseHandler)
	hyvaSuomiResponseHandler.SetNext(videoResponseHandler)
	videoResponseHandler.SetNext(videoStatsHandler)
	videoStatsHandler.SetNext(markForNaggingHandler)
	markForNaggingHandler.SetNext(markForDeletionHandler)
	markForDeletionHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(imageResponseHandler)
	imageResponseHandler.SetNext(deleteMessageHandler)
	deleteMessageHandler.SetNext(textResponseHandler)
	textResponseHandler.SetNext(endOfChainHandler)

	return onTextHandler
}

// newTestBot creates a telebot connected to the mock Telegram server.
func newTestBot(t *testing.T, mockURL string) *tele.Bot {
	t.Helper()
	bot, err := tele.NewBot(tele.Settings{
		URL:         mockURL,
		Token:       "000000000:fake-test-token",
		Synchronous: true,
	})
	if err != nil {
		t.Fatalf("failed to create test bot: %v", err)
	}
	return bot
}

func TestPingE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	t.Setenv("ENABLED_FEATURES", "ping")

	mock := NewMockTelegramServer(t)
	bot := newTestBot(t, mock.Server.URL)

	chain := buildFullChain()
	bot.Handle(tele.OnText, func(c tele.Context) error {
		chain.Execute(&Context{
			TelebotContext: c,
			Telebot:        bot,
			Service:        Telegram,
		})
		return nil
	})

	bot.ProcessUpdate(tele.Update{
		Message: &tele.Message{
			ID:   1,
			Text: "/ping",
			Chat: &tele.Chat{ID: -1001234567890},
			Sender: &tele.User{
				ID:       999,
				Username: "testuser",
			},
		},
	})

	if len(mock.SentMessages) == 0 {
		t.Fatal("expected at least one sendMessage call, got none")
	}
	found := false
	for _, msg := range mock.SentMessages {
		if msg.Text == "pong" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected sendMessage with text 'pong', got: %+v", mock.SentMessages)
	}
}

func TestVideoDownloadE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found on PATH, skipping e2e test")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found on PATH, skipping e2e test")
	}

	tmpDir := t.TempDir()
	dbDir := t.TempDir()

	t.Setenv("ENABLED_FEATURES", "dl")
	t.Setenv("YTDLP_TMP_DIR", tmpDir)
	t.Setenv("REPOST_DB_DIR", dbDir)

	// Initialize the repost DB
	dbPath := filepath.Join(dbDir, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		t.Fatalf("failed to init repost DB: %v", err)
	}

	mock := NewMockTelegramServer(t)
	videoSrv := newVideoFileServer(t)
	bot := newTestBot(t, mock.Server.URL)

	chain := buildFullChain()
	bot.Handle(tele.OnText, func(c tele.Context) error {
		chain.Execute(&Context{
			TelebotContext: c,
			Telebot:        bot,
			Service:        Telegram,
		})
		return nil
	})

	bot.ProcessUpdate(tele.Update{
		Message: &tele.Message{
			ID:   2,
			Text: "/dl " + videoSrv.URL + "/sample.mp4",
			Chat: &tele.Chat{ID: -1001234567890},
			Sender: &tele.User{
				ID:       999,
				Username: "testuser",
			},
		},
	})

	if len(mock.SentVideos) == 0 {
		t.Fatal("expected at least one sendVideo call, got none")
	}
	if len(mock.SentVideos[0].Video) == 0 {
		t.Error("expected non-empty video data in sendVideo call")
	}
	if len(mock.ChatActionsSent) == 0 {
		t.Error("expected at least one sendChatAction call (typing indicator)")
	}
}

func TestVideoDownloadInvalidURL_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found on PATH, skipping e2e test")
	}

	tmpDir := t.TempDir()
	dbDir := t.TempDir()

	t.Setenv("ENABLED_FEATURES", "dl")
	t.Setenv("YTDLP_TMP_DIR", tmpDir)
	t.Setenv("REPOST_DB_DIR", dbDir)

	dbPath := filepath.Join(dbDir, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		t.Fatalf("failed to init repost DB: %v", err)
	}

	mock := NewMockTelegramServer(t)
	videoSrv := newVideoFileServer(t)
	bot := newTestBot(t, mock.Server.URL)

	chain := buildFullChain()
	bot.Handle(tele.OnText, func(c tele.Context) error {
		chain.Execute(&Context{
			TelebotContext: c,
			Telebot:        bot,
			Service:        Telegram,
		})
		return nil
	})

	// Request a URL that returns 404
	bot.ProcessUpdate(tele.Update{
		Message: &tele.Message{
			ID:   3,
			Text: "/dl " + videoSrv.URL + "/nonexistent.mp4",
			Chat: &tele.Chat{ID: -1001234567890},
			Sender: &tele.User{
				ID:       999,
				Username: "testuser",
			},
		},
	})

	if len(mock.SentVideos) != 0 {
		t.Errorf("expected no sendVideo calls for invalid URL, got %d", len(mock.SentVideos))
	}
}

func TestReactionsAndStatsE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found on PATH, skipping e2e test")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found on PATH, skipping e2e test")
	}

	tmpDir := t.TempDir()
	dbDir := t.TempDir()

	t.Setenv("ENABLED_FEATURES", "dl,stats")
	t.Setenv("YTDLP_TMP_DIR", tmpDir)
	t.Setenv("REPOST_DB_DIR", dbDir)

	dbPath := filepath.Join(dbDir, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		t.Fatalf("failed to init repost DB: %v", err)
	}

	mock := NewMockTelegramServer(t)
	videoSrv := newVideoFileServer(t)
	bot := newTestBot(t, mock.Server.URL)

	chain := buildFullChain()
	bot.Handle(tele.OnText, func(c tele.Context) error {
		chain.Execute(&Context{
			TelebotContext: c,
			Telebot:        bot,
			Service:        Telegram,
		})
		return nil
	})

	// Step 1: User 1 posts a video via /dl.
	bot.ProcessUpdate(tele.Update{
		Message: &tele.Message{
			ID:   10,
			Text: "/dl " + videoSrv.URL + "/sample.mp4",
			Chat: &tele.Chat{ID: -1001234567890},
			Sender: &tele.User{
				ID:       111,
				Username: "poster",
			},
		},
	})

	if len(mock.SentVideos) == 0 {
		t.Fatal("expected at least one sendVideo call after /dl, got none")
	}

	// Step 2: User 2 reacts with 👍. The mock returns message_id 42 for sendVideo,
	// so botMessageId stored in DB is "42". groupId = "telegram:-1001234567890".
	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open stats DB for reaction: %v", err)
	}
	if err := utils.UpdateReactionCount(db, "telegram", "telegram:-1001234567890", "42", "👍", 1); err != nil {
		db.Close()
		t.Fatalf("failed to update reaction count: %v", err)
	}
	db.Close()

	// Step 3: User 3 checks /stats.
	bot.ProcessUpdate(tele.Update{
		Message: &tele.Message{
			ID:   11,
			Text: "/stats",
			Chat: &tele.Chat{ID: -1001234567890},
			Sender: &tele.User{
				ID:       333,
				Username: "statsuser",
			},
		},
	})

	// Find the stats message among all sent messages.
	var statsText string
	for _, msg := range mock.SentMessages {
		if strings.Contains(msg.Text, "Top video posters:") {
			statsText = msg.Text
			break
		}
	}
	if statsText == "" {
		t.Fatalf("expected a sendMessage containing 'Top video posters:', got messages: %+v", mock.SentMessages)
	}

	checks := []struct {
		substr string
		desc   string
	}{
		{"poster", "poster username in leaderboard"},
		{"1 video", "video count for poster"},
		{"👍 Most liked:", "most liked section header"},
		{"poster — 1", "poster with 1 thumbs-up in most liked"},
	}
	for _, c := range checks {
		if !strings.Contains(statsText, c.substr) {
			t.Errorf("stats message missing %s: want %q in:\n%s", c.desc, c.substr, statsText)
		}
	}
}
