package utils

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "repost_fingerprints.duckdb")
	if err := InitRepostDB(dbPath); err != nil {
		t.Fatalf("InitRepostDB failed: %v", err)
	}
	db, err := OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRecordVideoPost(t *testing.T) {
	db := setupTestDB(t)

	entry := VideoStatEntry{
		Platform:     "discord",
		GroupId:      "discord:123",
		UserId:       "user1",
		Username:     "testuser",
		SourceUrl:    "https://youtube.com/watch?v=abc",
		BotMessageId: "msg1",
		PostedAt:     time.Now(),
	}

	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}

	posters, err := GetGroupLeaderboard(db, "discord:123", 10)
	if err != nil {
		t.Fatalf("GetGroupLeaderboard failed: %v", err)
	}
	if len(posters) != 1 {
		t.Fatalf("expected 1 poster, got %d", len(posters))
	}
	if posters[0].Username != "testuser" {
		t.Errorf("expected username %q, got %q", "testuser", posters[0].Username)
	}
	if posters[0].PostCount != 1 {
		t.Errorf("expected post count 1, got %d", posters[0].PostCount)
	}
}

func TestGetGroupLeaderboardOrdering(t *testing.T) {
	db := setupTestDB(t)

	entries := []VideoStatEntry{
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/1", BotMessageId: "m1", PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u2", Username: "bob", SourceUrl: "https://example.com/2", BotMessageId: "m2", PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/3", BotMessageId: "m3", PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/4", BotMessageId: "m4", PostedAt: time.Now()},
	}

	for _, e := range entries {
		if err := RecordVideoPost(db, e); err != nil {
			t.Fatalf("RecordVideoPost failed: %v", err)
		}
	}

	posters, err := GetGroupLeaderboard(db, "discord:123", 10)
	if err != nil {
		t.Fatalf("GetGroupLeaderboard failed: %v", err)
	}
	if len(posters) != 2 {
		t.Fatalf("expected 2 posters, got %d", len(posters))
	}
	if posters[0].Username != "alice" || posters[0].PostCount != 3 {
		t.Errorf("expected alice with 3 posts first, got %q with %d", posters[0].Username, posters[0].PostCount)
	}
	if posters[1].Username != "bob" || posters[1].PostCount != 1 {
		t.Errorf("expected bob with 1 post second, got %q with %d", posters[1].Username, posters[1].PostCount)
	}
}

func TestGetGroupLeaderboardGroupsByUserId(t *testing.T) {
	db := setupTestDB(t)

	// Same user_id, different usernames (user renamed themselves).
	entries := []VideoStatEntry{
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/1", BotMessageId: "m1", PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice_renamed", SourceUrl: "https://example.com/2", BotMessageId: "m2", PostedAt: time.Now()},
	}
	for _, e := range entries {
		if err := RecordVideoPost(db, e); err != nil {
			t.Fatalf("RecordVideoPost failed: %v", err)
		}
	}

	posters, err := GetGroupLeaderboard(db, "discord:123", 10)
	if err != nil {
		t.Fatalf("GetGroupLeaderboard failed: %v", err)
	}
	if len(posters) != 1 || posters[0].PostCount != 2 {
		t.Errorf("expected renamed user grouped into single row with 2 posts, got %v", posters)
	}
}

func TestGetGroupLeaderboardExcludesReposts(t *testing.T) {
	db := setupTestDB(t)

	entries := []VideoStatEntry{
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/1", BotMessageId: "m1", IsRepost: false, PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/2", BotMessageId: "m2", IsRepost: true, PostedAt: time.Now()},
	}
	for _, e := range entries {
		if err := RecordVideoPost(db, e); err != nil {
			t.Fatalf("RecordVideoPost failed: %v", err)
		}
	}

	posters, err := GetGroupLeaderboard(db, "discord:123", 10)
	if err != nil {
		t.Fatalf("GetGroupLeaderboard failed: %v", err)
	}
	if len(posters) != 1 || posters[0].PostCount != 1 {
		t.Errorf("leaderboard should only count non-reposts, got %v", posters)
	}
}

func TestGetGroupLeaderboardIsolatesGroups(t *testing.T) {
	db := setupTestDB(t)

	entries := []VideoStatEntry{
		{Platform: "discord", GroupId: "discord:111", UserId: "u1", Username: "alice", SourceUrl: "https://example.com/1", BotMessageId: "m1", PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:222", UserId: "u2", Username: "bob", SourceUrl: "https://example.com/2", BotMessageId: "m2", PostedAt: time.Now()},
	}
	for _, e := range entries {
		if err := RecordVideoPost(db, e); err != nil {
			t.Fatalf("RecordVideoPost failed: %v", err)
		}
	}

	posters, err := GetGroupLeaderboard(db, "discord:111", 10)
	if err != nil {
		t.Fatalf("GetGroupLeaderboard failed: %v", err)
	}
	if len(posters) != 1 || posters[0].Username != "alice" {
		t.Errorf("expected only alice in group 111, got %v", posters)
	}
}

func TestUpdateReactionCount(t *testing.T) {
	db := setupTestDB(t)

	entry := VideoStatEntry{
		Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
		SourceUrl: "https://example.com/1", BotMessageId: "msg1", PostedAt: time.Now(),
	}
	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}

	if err := UpdateReactionCount(db, "discord", "discord:123", "msg1", "👍", +1); err != nil {
		t.Fatalf("UpdateReactionCount +1 failed: %v", err)
	}
	if err := UpdateReactionCount(db, "discord", "discord:123", "msg1", "👍", +1); err != nil {
		t.Fatalf("UpdateReactionCount +1 failed: %v", err)
	}

	videos, err := GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}
	if videos[0].ReactionCount != 2 {
		t.Errorf("expected thumbs_up count 2, got %d", videos[0].ReactionCount)
	}

	if err := UpdateReactionCount(db, "discord", "discord:123", "msg1", "👍", -1); err != nil {
		t.Fatalf("UpdateReactionCount -1 failed: %v", err)
	}

	videos, err = GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if videos[0].ReactionCount != 1 {
		t.Errorf("expected thumbs_up count 1 after removal, got %d", videos[0].ReactionCount)
	}
}

func TestUpdateReactionCountThumbsDown(t *testing.T) {
	db := setupTestDB(t)

	entry := VideoStatEntry{
		Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
		SourceUrl: "https://example.com/1", BotMessageId: "msg1", PostedAt: time.Now(),
	}
	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}

	UpdateReactionCount(db, "discord", "discord:123", "msg1", "👎", +1)
	UpdateReactionCount(db, "discord", "discord:123", "msg1", "👎", +1)
	UpdateReactionCount(db, "discord", "discord:123", "msg1", "👍", +1)

	downVideos, err := GetTopThumbsDown(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsDown failed: %v", err)
	}
	if len(downVideos) != 1 || downVideos[0].ReactionCount != 2 {
		t.Errorf("expected 2 thumbs down, got %v", downVideos)
	}

	upVideos, err := GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if len(upVideos) != 1 || upVideos[0].ReactionCount != 1 {
		t.Errorf("expected 1 thumbs up, got %v", upVideos)
	}
}

func TestUpdateReactionCountFire(t *testing.T) {
	db := setupTestDB(t)

	entry := VideoStatEntry{
		Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
		SourceUrl: "https://example.com/1", BotMessageId: "msg1", PostedAt: time.Now(),
	}
	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}

	UpdateReactionCount(db, "discord", "discord:123", "msg1", "🔥", +1)
	UpdateReactionCount(db, "discord", "discord:123", "msg1", "🔥", +1)
	UpdateReactionCount(db, "discord", "discord:123", "msg1", "👎", +1)

	upVideos, err := GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if len(upVideos) != 1 || upVideos[0].ReactionCount != 2 {
		t.Errorf("expected 2 thumbs_up_count from 🔥, got %v", upVideos)
	}

	downVideos, err := GetTopThumbsDown(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsDown failed: %v", err)
	}
	if len(downVideos) != 1 || downVideos[0].ReactionCount != 1 {
		t.Errorf("expected 1 thumbs_down_count, got %v", downVideos)
	}
}

func TestUpdateReactionCountNauseated(t *testing.T) {
	db := setupTestDB(t)

	entry := VideoStatEntry{
		Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
		SourceUrl: "https://example.com/1", BotMessageId: "msg1", PostedAt: time.Now(),
	}
	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}

	UpdateReactionCount(db, "discord", "discord:123", "msg1", "🤮", +1)
	UpdateReactionCount(db, "discord", "discord:123", "msg1", "🤮", +1)
	UpdateReactionCount(db, "discord", "discord:123", "msg1", "👍", +1)

	downVideos, err := GetTopThumbsDown(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsDown failed: %v", err)
	}
	if len(downVideos) != 1 || downVideos[0].ReactionCount != 2 {
		t.Errorf("expected 2 thumbs_down_count from 🤮, got %v", downVideos)
	}

	upVideos, err := GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if len(upVideos) != 1 || upVideos[0].ReactionCount != 1 {
		t.Errorf("expected 1 thumbs_up_count, got %v", upVideos)
	}
}

func TestUpdateReactionCountUnknownMessage(t *testing.T) {
	db := setupTestDB(t)

	if err := UpdateReactionCount(db, "discord", "discord:123", "nonexistent-msg", "👍", +1); err != nil {
		t.Errorf("UpdateReactionCount on unknown message should not error, got: %v", err)
	}
}

func TestGetTopThumbsUpOrdering(t *testing.T) {
	db := setupTestDB(t)

	entries := []VideoStatEntry{
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://a.com", BotMessageId: "m1", PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u2", Username: "bob", SourceUrl: "https://b.com", BotMessageId: "m2", PostedAt: time.Now()},
	}
	for _, e := range entries {
		if err := RecordVideoPost(db, e); err != nil {
			t.Fatalf("RecordVideoPost failed: %v", err)
		}
	}

	for range 3 {
		UpdateReactionCount(db, "discord", "discord:123", "m2", "👍", +1)
	}
	UpdateReactionCount(db, "discord", "discord:123", "m1", "👍", +1)

	videos, err := GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if len(videos) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(videos))
	}
	if videos[0].Username != "bob" || videos[0].ReactionCount != 3 {
		t.Errorf("expected bob with 3 thumbs up first, got %q with %d", videos[0].Username, videos[0].ReactionCount)
	}
}

func TestGetTopThumbsUpExcludesZero(t *testing.T) {
	db := setupTestDB(t)

	entry := VideoStatEntry{
		Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
		SourceUrl: "https://example.com/1", BotMessageId: "m1", PostedAt: time.Now(),
	}
	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost failed: %v", err)
	}

	videos, err := GetTopThumbsUp(db, "discord:123", 5)
	if err != nil {
		t.Fatalf("GetTopThumbsUp failed: %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("expected 0 videos with no thumbs up, got %d", len(videos))
	}
}

func TestGetTopReposters(t *testing.T) {
	db := setupTestDB(t)

	entries := []VideoStatEntry{
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://a.com", BotMessageId: "m1", IsRepost: true, PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice", SourceUrl: "https://b.com", BotMessageId: "m2", IsRepost: true, PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u2", Username: "bob", SourceUrl: "https://c.com", BotMessageId: "m3", IsRepost: true, PostedAt: time.Now()},
		{Platform: "discord", GroupId: "discord:123", UserId: "u3", Username: "carol", SourceUrl: "https://d.com", BotMessageId: "m4", IsRepost: false, PostedAt: time.Now()},
	}
	for _, e := range entries {
		if err := RecordVideoPost(db, e); err != nil {
			t.Fatalf("RecordVideoPost failed: %v", err)
		}
	}

	reposters, err := GetTopReposters(db, "discord:123", 10)
	if err != nil {
		t.Fatalf("GetTopReposters failed: %v", err)
	}
	if len(reposters) != 2 {
		t.Fatalf("expected 2 reposters (carol posted original, not repost), got %d", len(reposters))
	}
	if reposters[0].Username != "alice" || reposters[0].RepostCount != 2 {
		t.Errorf("expected alice with 2 reposts first, got %q with %d", reposters[0].Username, reposters[0].RepostCount)
	}
}

func TestUpdateReactionCountPositiveEmojis(t *testing.T) {
	for _, emoji := range positiveEmojis {
		t.Run(emoji, func(t *testing.T) {
			db := setupTestDB(t)

			entry := VideoStatEntry{
				Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
				SourceUrl: "https://example.com/1", BotMessageId: "msg1", PostedAt: time.Now(),
			}
			if err := RecordVideoPost(db, entry); err != nil {
				t.Fatalf("RecordVideoPost failed: %v", err)
			}

			if err := UpdateReactionCount(db, "discord", "discord:123", "msg1", emoji, +1); err != nil {
				t.Fatalf("UpdateReactionCount(%q) failed: %v", emoji, err)
			}

			videos, err := GetTopThumbsUp(db, "discord:123", 5)
			if err != nil {
				t.Fatalf("GetTopThumbsUp failed: %v", err)
			}
			if len(videos) != 1 || videos[0].ReactionCount != 1 {
				t.Errorf("emoji %q: expected 1 thumbs_up_count, got %v", emoji, videos)
			}
		})
	}
}

func TestUpdateReactionCountNegativeEmojis(t *testing.T) {
	for _, emoji := range negativeEmojis {
		t.Run(emoji, func(t *testing.T) {
			db := setupTestDB(t)

			entry := VideoStatEntry{
				Platform: "discord", GroupId: "discord:123", UserId: "u1", Username: "alice",
				SourceUrl: "https://example.com/1", BotMessageId: "msg1", PostedAt: time.Now(),
			}
			if err := RecordVideoPost(db, entry); err != nil {
				t.Fatalf("RecordVideoPost failed: %v", err)
			}

			if err := UpdateReactionCount(db, "discord", "discord:123", "msg1", emoji, +1); err != nil {
				t.Fatalf("UpdateReactionCount(%q) failed: %v", emoji, err)
			}

			videos, err := GetTopThumbsDown(db, "discord:123", 5)
			if err != nil {
				t.Fatalf("GetTopThumbsDown failed: %v", err)
			}
			if len(videos) != 1 || videos[0].ReactionCount != 1 {
				t.Errorf("emoji %q: expected 1 thumbs_down_count, got %v", emoji, videos)
			}
		})
	}
}

func TestQueryTopByReactionColumnRejectsUnknownColumn(t *testing.T) {
	db := setupTestDB(t)
	if _, err := queryTopByReactionColumn(db, "discord:123", "id; DROP TABLE video_stats;--", 5); err == nil {
		t.Error("expected error on unknown reaction column, got nil")
	}
}
