package handlers

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/napuu/gpsp-bot/pkg/utils"
)

func setupStatsTestDB(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		t.Fatalf("InitRepostDB failed: %v", err)
	}
	return dbPath
}

func TestStatsHandlerSetsTextResponse(t *testing.T) {
	dbPath := setupStatsTestDB(t)

	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB failed: %v", err)
	}
	if err := utils.RecordVideoPost(db, utils.VideoStatEntry{
		Platform: "discord", GroupId: "discord:999", UserId: "u1", Username: "alice",
		SourceUrl: "https://youtube.com/watch?v=1", BotMessageId: "m1", PostedAt: time.Now(),
	}); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}
	db.Close()

	t.Setenv("ENABLED_FEATURES", "stats")
	t.Setenv("REPOST_DB_DIR", filepath.Dir(dbPath))

	ctx := &Context{
		Service: Discord,
		chatId:  "999",
		action:  Stats,
	}

	statsHandler := &StatsHandler{}
	end := &mockHandler{}
	statsHandler.SetNext(end)

	statsHandler.Execute(ctx)

	if ctx.textResponse == "" {
		t.Error("expected textResponse to be set, got empty string")
	}
	if !strings.Contains(ctx.textResponse, "alice") {
		t.Errorf("expected textResponse to contain 'alice', got: %q", ctx.textResponse)
	}
	if !end.executed {
		t.Error("expected next handler to be called")
	}
}

func TestStatsHandlerSkipsWhenActionIsNotStats(t *testing.T) {
	ctx := &Context{
		Service: Discord,
		chatId:  "999",
		action:  Ping,
	}

	statsHandler := &StatsHandler{}
	end := &mockHandler{}
	statsHandler.SetNext(end)

	statsHandler.Execute(ctx)

	if ctx.textResponse != "" {
		t.Errorf("expected empty textResponse for non-stats action, got: %q", ctx.textResponse)
	}
	if !end.executed {
		t.Error("expected next handler to be called")
	}
}

func TestBuildStatsTextWithData(t *testing.T) {
	posters := []utils.PosterStat{
		{UserId: "u1", Username: "alice", PostCount: 5},
		{UserId: "u2", Username: "bob", PostCount: 2},
	}
	thumbsUp := []utils.ReactionStat{
		{BotMessageId: "m1", Username: "alice", SourceUrl: "https://discord.com/channels/123/999/m1", ReactionCount: 7, PostedAt: time.Now()},
	}
	thumbsDown := []utils.ReactionStat{
		{BotMessageId: "m2", Username: "bob", SourceUrl: "https://discord.com/channels/123/999/m2", ReactionCount: 3, PostedAt: time.Now()},
	}
	reposters := []utils.RepostStat{
		{UserId: "u2", Username: "bob", RepostCount: 4},
	}

	ctx := &Context{
		Service: Discord,
		chatId:  "999",
		guildId: "123",
		action:  Stats,
	}

	result := buildStatsText(ctx, posters, nil, thumbsUp, nil, thumbsDown, nil, reposters, nil)

	if !strings.Contains(result, "alice") {
		t.Errorf("expected result to contain 'alice', got: %q", result)
	}
	if !strings.Contains(result, "5 videos") {
		t.Errorf("expected result to contain '5 videos', got: %q", result)
	}
	if !strings.Contains(result, "👍") {
		t.Errorf("expected result to contain thumbs up section, got: %q", result)
	}
	if !strings.Contains(result, "👎") {
		t.Errorf("expected result to contain thumbs down section, got: %q", result)
	}
	if !strings.Contains(result, "4 reposts") {
		t.Errorf("expected result to contain '4 reposts', got: %q", result)
	}
}

func TestBuildStatsTextEmpty(t *testing.T) {
	ctx := &Context{Service: Discord, chatId: "999", action: Stats}
	result := buildStatsText(ctx, nil, nil, nil, nil, nil, nil, nil, nil)

	if !strings.Contains(result, "No videos posted yet") {
		t.Errorf("expected empty-state message for posters, got: %q", result)
	}
	if !strings.Contains(result, "None yet") {
		t.Errorf("expected empty-state message for reactions, got: %q", result)
	}
	if !strings.Contains(result, "No reposts detected") {
		t.Errorf("expected empty-state message for reposters, got: %q", result)
	}
}

func TestBuildStatsTextSingularPlural(t *testing.T) {
	ctx := &Context{Service: Discord, chatId: "999", action: Stats}
	posters := []utils.PosterStat{{UserId: "u1", Username: "alice", PostCount: 1}}
	reposters := []utils.RepostStat{{UserId: "u2", Username: "bob", RepostCount: 1}}

	result := buildStatsText(ctx, posters, nil, nil, nil, nil, nil, reposters, nil)

	if strings.Contains(result, "1 videos") {
		t.Errorf("expected singular '1 video', got: %q", result)
	}
	if !strings.Contains(result, "1 video") {
		t.Errorf("expected '1 video' in result, got: %q", result)
	}
	if strings.Contains(result, "1 reposts") {
		t.Errorf("expected singular '1 repost', got: %q", result)
	}
	if !strings.Contains(result, "1 repost") {
		t.Errorf("expected '1 repost' in result, got: %q", result)
	}
}
