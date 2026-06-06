package utils

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRecordGroupActivityAndGet(t *testing.T) {
	db := setupTestDB(t)
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)
	memberCount := 10

	if err := RecordGroupActivity(db, "discord:123", "discord", &memberCount, now); err != nil {
		t.Fatalf("RecordGroupActivity() error: %v", err)
	}

	activity, err := GetGroupActivity(db, "discord:123")
	if err != nil {
		t.Fatalf("GetGroupActivity() error: %v", err)
	}
	if activity.Platform != "discord" {
		t.Fatalf("platform = %q, want discord", activity.Platform)
	}
	if !activity.MemberCount.Valid || activity.MemberCount.Int64 != 10 {
		t.Fatalf("member count = %v, want 10", activity.MemberCount)
	}
}

func TestDayVideoStateRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)
	eligible := now.Add(14 * 24 * time.Hour)
	dueBy := now.Add(21 * 24 * time.Hour)

	state := DayVideoStateRow{
		GroupID:      "telegram:999",
		LastPostedAt: sql.NullTime{Time: now, Valid: true},
		EligibleFrom: eligible,
		DueBy:        dueBy,
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
	if !got.EligibleFrom.Equal(eligible) || !got.DueBy.Equal(dueBy) {
		t.Fatalf("schedule mismatch: got (%v, %v)", got.EligibleFrom, got.DueBy)
	}
}

func TestListDueGroups(t *testing.T) {
	db := setupTestDB(t)
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)

	states := []DayVideoStateRow{
		{GroupID: "discord:1", EligibleFrom: now.Add(-time.Hour), DueBy: now.Add(7 * 24 * time.Hour)},
		{GroupID: "discord:2", EligibleFrom: now.Add(time.Hour), DueBy: now.Add(8 * 24 * time.Hour)},
	}
	for _, s := range states {
		if err := UpsertDayVideoState(db, s); err != nil {
			t.Fatalf("UpsertDayVideoState() error: %v", err)
		}
	}

	due, err := ListDueGroups(db, now)
	if err != nil {
		t.Fatalf("ListDueGroups() error: %v", err)
	}
	if len(due) != 1 || due[0] != "discord:1" {
		t.Fatalf("due groups = %v, want [discord:1]", due)
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

func seedActiveGroup(t *testing.T, db *sql.DB, groupID string, now time.Time, memberCount int) {
	t.Helper()
	platform := "discord"
	if idx := strings.Index(groupID, ":"); idx > 0 {
		platform = groupID[:idx]
	}
	count := memberCount
	if err := RecordGroupActivity(db, groupID, platform, &count, now); err != nil {
		t.Fatalf("RecordGroupActivity() error: %v", err)
	}
	if err := UpsertDayVideoState(db, DayVideoStateRow{
		GroupID:      groupID,
		EligibleFrom: now.Add(-time.Hour),
		DueBy:        now.Add(7 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("UpsertDayVideoState() error: %v", err)
	}
	for i := 0; i < 7; i++ {
		entry := VideoStatEntry{
			Platform:     platform,
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    fmt.Sprintf("https://example.com/%s/%d", groupID, i),
			BotMessageId: fmt.Sprintf("m%d", i),
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost() error: %v", err)
		}
	}
}

func TestListDueActiveGroups(t *testing.T) {
	db := setupTestDB(t)
	now := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)

	seedActiveGroup(t, db, "discord:active", now, 10)

	if err := UpsertDayVideoState(db, DayVideoStateRow{
		GroupID:      "discord:stale",
		EligibleFrom: now.Add(-time.Hour),
		DueBy:        now.Add(7 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("UpsertDayVideoState() error: %v", err)
	}
	staleCount := 10
	if err := RecordGroupActivity(db, "discord:stale", "discord", &staleCount, now.Add(-25*time.Hour)); err != nil {
		t.Fatalf("RecordGroupActivity() error: %v", err)
	}

	seedActiveGroup(t, db, "discord:low-members", now, 2)

	lowVideoCount := 10
	if err := RecordGroupActivity(db, "discord:low-videos", "discord", &lowVideoCount, now); err != nil {
		t.Fatalf("RecordGroupActivity() error: %v", err)
	}
	if err := UpsertDayVideoState(db, DayVideoStateRow{
		GroupID:      "discord:low-videos",
		EligibleFrom: now.Add(-time.Hour),
		DueBy:        now.Add(7 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("UpsertDayVideoState() error: %v", err)
	}
	for i := 0; i < 6; i++ {
		entry := VideoStatEntry{
			Platform:     "discord",
			GroupId:      "discord:low-videos",
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    fmt.Sprintf("https://example.com/low/%d", i),
			BotMessageId: fmt.Sprintf("lv%d", i),
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost() error: %v", err)
		}
	}

	due, err := ListDueActiveGroups(db, now)
	if err != nil {
		t.Fatalf("ListDueActiveGroups() error: %v", err)
	}
	if len(due) != 1 || due[0] != "discord:active" {
		t.Fatalf("due active groups = %v, want [discord:active]", due)
	}

	allDue, err := ListDueGroups(db, now)
	if err != nil {
		t.Fatalf("ListDueGroups() error: %v", err)
	}
	if len(allDue) < 4 {
		t.Fatalf("ListDueGroups() = %v, want at least 4 due rows", allDue)
	}
}
