package utils

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDayVideoStateRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)
	checked := now.Add(-2 * time.Hour)
	eligible := now.Add(14 * 24 * time.Hour)
	dueBy := now.Add(21 * 24 * time.Hour)

	state := DayVideoStateRow{
		GroupID:       "telegram:999",
		LastPostedAt:  sql.NullTime{Time: now, Valid: true},
		LastCheckedAt: sql.NullTime{Time: checked, Valid: true},
		EligibleFrom:  eligible,
		DueBy:         dueBy,
	}
	if err := UpsertDayVideoState(db, state); err != nil {
		t.Fatalf("UpsertDayVideoState() error: %v", err)
	}

	got, err := GetDayVideoState(db, "telegram:999")
	if err != nil {
		t.Fatalf("GetDayVideoState() error: %v", err)
	}
	if !got.LastPostedAt.Valid || !got.LastPostedAt.Time.Equal(now) {
		t.Fatalf("last posted = %v, want %v", got.LastPostedAt, now)
	}
	if !got.LastCheckedAt.Valid || !got.LastCheckedAt.Time.Equal(checked) {
		t.Fatalf("last checked = %v, want %v", got.LastCheckedAt, checked)
	}
	if !got.EligibleFrom.Equal(eligible) || !got.DueBy.Equal(dueBy) {
		t.Fatalf("schedule mismatch: got (%v, %v)", got.EligibleFrom, got.DueBy)
	}
}

func TestCountRecentVideoPosts(t *testing.T) {
	db := setupTestDB(t)
	groupID := "discord:555"
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 6; i++ {
		entry := VideoStatEntry{
			Platform:     "discord",
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    fmt.Sprintf("https://example.com/%d", i),
			BotMessageId: fmt.Sprintf("m%d", i),
			IsGroupChat:  true,
			PostedAt:     now.Add(-time.Duration(i) * 24 * time.Hour),
		}
		if err := RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost() error: %v", err)
		}
	}

	since := now.Add(-7 * 24 * time.Hour)
	count, err := CountRecentVideoPosts(db, groupID, since)
	if err != nil {
		t.Fatalf("CountRecentVideoPosts() error: %v", err)
	}
	if count != 6 {
		t.Fatalf("count = %d, want 6", count)
	}

	entry := VideoStatEntry{
		Platform:     "discord",
		GroupId:      groupID,
		UserId:       "u1",
		Username:     "alice",
		SourceUrl:    "https://example.com/seventh",
		BotMessageId: "m7",
		IsGroupChat:  true,
		PostedAt:     now.Add(-6 * time.Hour),
	}
	if err := RecordVideoPost(db, entry); err != nil {
		t.Fatalf("RecordVideoPost() error: %v", err)
	}

	count, err = CountRecentVideoPosts(db, groupID, since)
	if err != nil {
		t.Fatalf("CountRecentVideoPosts() error: %v", err)
	}
	if count != 7 {
		t.Fatalf("count = %d, want 7", count)
	}
}

func TestCountRecentVideoPostsExcludesDMs(t *testing.T) {
	db := setupTestDB(t)
	groupID := "telegram:dm"
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 7; i++ {
		entry := VideoStatEntry{
			Platform:     "telegram",
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    fmt.Sprintf("https://example.com/dm/%d", i),
			BotMessageId: fmt.Sprintf("dm%d", i),
			IsGroupChat:  false,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost() error: %v", err)
		}
	}

	count, err := CountRecentVideoPosts(db, groupID, now.Add(-7*24*time.Hour))
	if err != nil {
		t.Fatalf("CountRecentVideoPosts() error: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 for DM-only rows", count)
	}
}

func TestCountRecentVideoPostsExcludesReposts(t *testing.T) {
	db := setupTestDB(t)
	groupID := "discord:777"
	now := time.Now()

	for i := 0; i < 7; i++ {
		entry := VideoStatEntry{
			Platform:     "discord",
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    fmt.Sprintf("https://example.com/r%d", i),
			BotMessageId: fmt.Sprintf("r%d", i),
			IsRepost:     true,
			IsGroupChat:  true,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost() error: %v", err)
		}
	}

	count, err := CountRecentVideoPosts(db, groupID, now.Add(-7*24*time.Hour))
	if err != nil {
		t.Fatalf("CountRecentVideoPosts() error: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 for repost-only rows", count)
	}
}

func recordVideos(t *testing.T, db *sql.DB, groupID string, count int, isGroupChat, isRepost bool, now time.Time) {
	t.Helper()
	platform := "discord"
	if idx := strings.Index(groupID, ":"); idx > 0 {
		platform = groupID[:idx]
	}
	for i := 0; i < count; i++ {
		entry := VideoStatEntry{
			Platform:     platform,
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    fmt.Sprintf("https://example.com/%s/%d", groupID, i),
			BotMessageId: fmt.Sprintf("%s-m%d", groupID, i),
			IsRepost:     isRepost,
			IsGroupChat:  isGroupChat,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost() error: %v", err)
		}
	}
}

func TestListActiveVideoGroups(t *testing.T) {
	db := setupTestDB(t)
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)

	// Active group chat with enough recent videos.
	recordVideos(t, db, "discord:active", 7, true, false, now)
	// DM with plenty of videos must not qualify.
	recordVideos(t, db, "telegram:dm", 10, false, false, now)
	// Group chat with too few videos.
	recordVideos(t, db, "discord:low-videos", 6, true, false, now)
	// Group chat where activity is only reposts.
	recordVideos(t, db, "discord:reposts", 9, true, true, now)
	// Group chat whose videos are outside the lookback window.
	recordVideos(t, db, "discord:stale", 9, true, false, now.Add(-30*24*time.Hour))

	groups, err := ListActiveVideoGroups(db, now)
	if err != nil {
		t.Fatalf("ListActiveVideoGroups() error: %v", err)
	}
	if len(groups) != 1 || groups[0] != "discord:active" {
		t.Fatalf("active video groups = %v, want [discord:active]", groups)
	}
}
