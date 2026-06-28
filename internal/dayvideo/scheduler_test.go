package dayvideo

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/napuu/gpsp-bot/pkg/utils"
)

type mockClock struct {
	now time.Time
}

func (c *mockClock) Now() time.Time { return c.now }

type mockPoster struct {
	mu    sync.Mutex
	posts []PostContext
}

func (m *mockPoster) PostDayVideo(ctx PostContext, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.posts = append(m.posts, ctx)
	return nil
}

func (m *mockPoster) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.posts)
}

func setupSchedulerTest(t *testing.T) (*Scheduler, *mockPoster, *mockClock, string) {
	t.Helper()
	t.Setenv("ENABLED_FEATURES", "daymeme")

	dbPath := filepath.Join(t.TempDir(), "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		t.Fatalf("InitRepostDB: %v", err)
	}

	clock := &mockClock{now: time.Date(2026, time.June, 10, 8, 0, 0, 0, time.UTC)}
	poster := &mockPoster{}
	sched := NewSchedulerForTest(dbPath, nil, nil, clock, rand.New(rand.NewSource(1)), poster, time.Hour)
	sched.SetGenerateVideoForTest(func(_ time.Time) (string, error) {
		f, err := os.CreateTemp(t.TempDir(), "day-*.mp4")
		if err != nil {
			return "", err
		}
		f.Close()
		return f.Name(), nil
	})
	return sched, poster, clock, dbPath
}

func seedEligibleGroup(t *testing.T, dbPath, groupID string, now time.Time) {
	t.Helper()
	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	platform, _, _ := parseGroupID(groupID)

	dueBy := now.Add(-time.Hour)
	if err := utils.UpsertDayVideoState(db, utils.DayVideoStateRow{
		GroupID:      groupID,
		EligibleFrom: dueBy.Add(-14 * 24 * time.Hour),
		DueBy:        dueBy,
	}); err != nil {
		t.Fatalf("UpsertDayVideoState: %v", err)
	}

	for i := 0; i < 7; i++ {
		entry := utils.VideoStatEntry{
			Platform:     platform,
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    "https://example.com/v",
			BotMessageId: fmt.Sprintf("m%d", i),
			IsGroupChat:  true,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := utils.RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost: %v", err)
		}
	}
}

func TestSchedulerRecordsLastCheckedWhenNotPosting(t *testing.T) {
	sched, _, clock, dbPath := setupSchedulerTest(t)
	groupID := "discord:check"
	now := clock.now

	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	eligible := now.Add(-2 * 24 * time.Hour)
	dueBy := now.Add(10 * 24 * time.Hour)
	if err := utils.UpsertDayVideoState(db, utils.DayVideoStateRow{
		GroupID:      groupID,
		EligibleFrom: eligible,
		DueBy:        dueBy,
	}); err != nil {
		t.Fatalf("UpsertDayVideoState: %v", err)
	}
	for i := 0; i < 7; i++ {
		entry := utils.VideoStatEntry{
			Platform:     "discord",
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    "https://example.com/v",
			BotMessageId: fmt.Sprintf("check-m%d", i),
			IsGroupChat:  true,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := utils.RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost: %v", err)
		}
	}
	db.Close()

	sched.tryPost(PostContext{GroupID: groupID, Platform: "discord", ChatID: "check"})

	db, err = utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	state, err := utils.GetDayVideoState(db, groupID)
	if err != nil {
		t.Fatalf("GetDayVideoState: %v", err)
	}
	if !state.LastCheckedAt.Valid || !state.LastCheckedAt.Time.Equal(now) {
		t.Fatalf("last_checked_at = %v, want %v", state.LastCheckedAt, now)
	}
	if state.LastPostedAt.Valid {
		t.Fatal("expected no post for early-window check")
	}
}

func TestSchedulerTryPostWhenDue(t *testing.T) {
	sched, poster, clock, dbPath := setupSchedulerTest(t)
	groupID := "discord:123"
	seedEligibleGroup(t, dbPath, groupID, clock.now)

	sched.tryPost(PostContext{GroupID: groupID, Platform: "discord", ChatID: "123"})
	if poster.count() != 1 {
		t.Fatalf("poster calls = %d, want 1", poster.count())
	}

	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	state, err := utils.GetDayVideoState(db, groupID)
	if err != nil {
		t.Fatalf("GetDayVideoState: %v", err)
	}
	if !state.LastPostedAt.Valid {
		t.Fatal("expected last_posted_at to be set")
	}
}

func TestSchedulerTryPostNoDoublePost(t *testing.T) {
	sched, poster, clock, dbPath := setupSchedulerTest(t)
	groupID := "discord:456"
	seedEligibleGroup(t, dbPath, groupID, clock.now)

	ctx := PostContext{GroupID: groupID, Platform: "discord", ChatID: "456"}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sched.tryPost(ctx)
	}()
	go func() {
		defer wg.Done()
		sched.tryPost(ctx)
	}()
	wg.Wait()

	if poster.count() != 1 {
		t.Fatalf("poster calls = %d, want 1", poster.count())
	}
}

func TestSchedulerSkipsInactiveGroup(t *testing.T) {
	sched, poster, clock, dbPath := setupSchedulerTest(t)
	groupID := "discord:789"
	now := clock.now

	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	// Schedule is due, but the group lacks enough recent group-chat videos.
	if err := utils.UpsertDayVideoState(db, utils.DayVideoStateRow{
		GroupID:      groupID,
		EligibleFrom: now.Add(-time.Hour),
		DueBy:        now.Add(-30 * time.Minute),
	}); err != nil {
		t.Fatalf("UpsertDayVideoState: %v", err)
	}
	for i := 0; i < 6; i++ {
		entry := utils.VideoStatEntry{
			Platform:     "discord",
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    "https://example.com/v",
			BotMessageId: fmt.Sprintf("inactive-m%d", i),
			IsGroupChat:  true,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := utils.RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost: %v", err)
		}
	}

	sched.tryPost(PostContext{GroupID: groupID, Platform: "discord", ChatID: "789"})
	if poster.count() != 0 {
		t.Fatalf("poster calls = %d, want 0 for inactive group", poster.count())
	}
}

func TestSchedulerScanSeedsNewActiveGroupWithoutPosting(t *testing.T) {
	sched, poster, clock, dbPath := setupSchedulerTest(t)
	groupID := "discord:fresh"
	now := clock.now

	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	for i := 0; i < 7; i++ {
		entry := utils.VideoStatEntry{
			Platform:     "discord",
			GroupId:      groupID,
			UserId:       "u1",
			Username:     "alice",
			SourceUrl:    "https://example.com/v",
			BotMessageId: fmt.Sprintf("fresh-m%d", i),
			IsGroupChat:  true,
			PostedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := utils.RecordVideoPost(db, entry); err != nil {
			t.Fatalf("RecordVideoPost: %v", err)
		}
	}
	db.Close()

	sched.runBackgroundScan()

	if poster.count() != 0 {
		t.Fatalf("poster calls = %d, want 0 for newly seeded group", poster.count())
	}

	db, err = utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	state, err := utils.GetDayVideoState(db, groupID)
	if err != nil {
		t.Fatalf("GetDayVideoState: %v", err)
	}
	if !state.EligibleFrom.After(now) {
		t.Fatalf("expected eligible_from %v to be in the future relative to %v", state.EligibleFrom, now)
	}
	if state.LastPostedAt.Valid {
		t.Fatal("expected last_posted_at to be unset for a freshly seeded group")
	}
}

func TestSchedulerRestoresStateOnPostFailure(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "daymeme")

	dbPath := filepath.Join(t.TempDir(), "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		t.Fatalf("InitRepostDB: %v", err)
	}

	clock := &mockClock{now: time.Date(2026, time.June, 10, 8, 0, 0, 0, time.UTC)}
	sched := NewSchedulerForTest(dbPath, nil, nil, clock, rand.New(rand.NewSource(1)), failPoster{}, time.Hour)
	sched.SetGenerateVideoForTest(func(_ time.Time) (string, error) {
		f, err := os.CreateTemp(t.TempDir(), "day-*.mp4")
		if err != nil {
			return "", err
		}
		f.Close()
		return f.Name(), nil
	})

	groupID := "discord:321"
	seedEligibleGroup(t, dbPath, groupID, clock.now)

	db, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	before, err := utils.GetDayVideoState(db, groupID)
	if err != nil {
		t.Fatalf("GetDayVideoState: %v", err)
	}
	db.Close()

	sched.tryPost(PostContext{GroupID: groupID, Platform: "discord", ChatID: "321"})

	db, err = utils.OpenStatsDB(dbPath)
	if err != nil {
		t.Fatalf("OpenStatsDB: %v", err)
	}
	defer db.Close()

	after, err := utils.GetDayVideoState(db, groupID)
	if err != nil {
		t.Fatalf("GetDayVideoState: %v", err)
	}
	if after.LastPostedAt.Valid {
		t.Fatal("expected last_posted_at to remain unset after failed post")
	}
	if !after.EligibleFrom.Equal(before.EligibleFrom) || !after.DueBy.Equal(before.DueBy) {
		t.Fatalf("schedule changed after failed post: before (%v, %v), after (%v, %v)",
			before.EligibleFrom, before.DueBy, after.EligibleFrom, after.DueBy)
	}
}

func TestEnabledRequiresExactFeatureName(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "antidaymeme")
	if Enabled() {
		t.Fatal("Enabled() = true for antidaymeme, want false")
	}

	t.Setenv("ENABLED_FEATURES", "ping;daymeme")
	if !Enabled() {
		t.Fatal("Enabled() = false for ping;daymeme, want true")
	}
}

type failPoster struct{}

func (failPoster) PostDayVideo(PostContext, string) error {
	return fmt.Errorf("send failed")
}

func TestParseGroupID(t *testing.T) {
	platform, chatID, err := parseGroupID("telegram:-100123")
	if err != nil || platform != "telegram" || chatID != "-100123" {
		t.Fatalf("parseGroupID() = (%q, %q, %v)", platform, chatID, err)
	}
	if _, _, err := parseGroupID("invalid"); err == nil {
		t.Fatal("expected error for invalid group id")
	}
}

var _ utils.ScheduleClock = (*mockClock)(nil)
