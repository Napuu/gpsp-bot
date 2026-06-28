package dayvideo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

const defaultTickerInterval = 30 * time.Minute

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type postAttempt struct {
	prevState utils.DayVideoStateRow
	at        time.Time
}

// Scheduler coordinates day meme posting checks and sends.
type Scheduler struct {
	dbPath         string
	telebot        *tele.Bot
	discord        *discordgo.Session
	clock          utils.ScheduleClock
	rand           utils.ScheduleRand
	poster         Poster
	tickerInterval time.Duration
	generateVideo  func(time.Time) (string, error)

	groupLocks sync.Map
}

// NewScheduler creates a scheduler with production dependencies.
func NewScheduler(dbPath string, telebot *tele.Bot, discord *discordgo.Session) *Scheduler {
	return &Scheduler{
		dbPath:         dbPath,
		telebot:        telebot,
		discord:        discord,
		clock:          realClock{},
		rand:           utils.NewScheduleRand(),
		poster:         NewPoster(),
		tickerInterval: defaultTickerInterval,
		generateVideo:  GenerateToTemp,
	}
}

// NewSchedulerForTest creates a scheduler with injectable dependencies.
func NewSchedulerForTest(dbPath string, telebot *tele.Bot, discord *discordgo.Session, clock utils.ScheduleClock, rng utils.ScheduleRand, poster Poster, tickerInterval time.Duration) *Scheduler {
	return &Scheduler{
		dbPath:         dbPath,
		telebot:        telebot,
		discord:        discord,
		clock:          clock,
		rand:           rng,
		poster:         poster,
		tickerInterval: tickerInterval,
		generateVideo:  GenerateToTemp,
	}
}

// Enabled reports whether the daymeme feature is active.
func Enabled() bool {
	return slices.Contains(config.EnabledFeatures(), "daymeme")
}

// Start launches the background scan loop until ctx is cancelled.
// Each iteration waits a jittered interval around tickerInterval (50–150%, averaging 100%).
func (s *Scheduler) Start(ctx context.Context) {
	if s == nil || !Enabled() {
		return
	}
	go func() {
		timer := time.NewTimer(s.nextTickDelay())
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.runBackgroundScan()
				timer.Reset(s.nextTickDelay())
			}
		}
	}()
}

func (s *Scheduler) nextTickDelay() time.Duration {
	base := float64(s.tickerInterval)
	return time.Duration(base * (0.5 + s.rand.Float64()))
}

func (s *Scheduler) runBackgroundScan() {
	db, err := utils.OpenStatsDB(s.dbPath)
	if err != nil {
		slog.Warn("day video background scan: open db", "error", err)
		return
	}

	now := s.clock.Now()
	groupIDs, err := utils.ListActiveVideoGroups(db, now)
	db.Close()
	if err != nil {
		slog.Warn("day video background scan: list active video groups", "error", err)
		return
	}

	for _, groupID := range groupIDs {
		platform, chatID, err := parseGroupID(groupID)
		if err != nil {
			slog.Warn("day video background scan: parse group", "groupId", groupID, "error", err)
			continue
		}
		s.tryPost(PostContext{
			GroupID:        groupID,
			Platform:       platform,
			ChatID:         chatID,
			Telebot:        s.telebot,
			DiscordSession: s.discord,
		})
	}
}

// reservePostIfDue reads schedule state and reserves the next window when a post should happen.
// A newly seen group is seeded with an initial schedule and skipped this round.
// The DB connection is closed before returning.
func (s *Scheduler) reservePostIfDue(ctx PostContext) (*postAttempt, bool) {
	db, err := utils.OpenStatsDB(s.dbPath)
	if err != nil {
		slog.Warn("day video: open db for post", "error", err)
		return nil, false
	}
	defer db.Close()

	now := s.clock.Now()
	state, err := utils.GetDayVideoState(db, ctx.GroupID)
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Warn("day video: load state", "error", err)
			return nil, false
		}
		eligibleFrom, dueBy := utils.InitialSchedule(now, s.rand)
		if err := utils.UpsertDayVideoState(db, utils.DayVideoStateRow{
			GroupID:      ctx.GroupID,
			EligibleFrom: eligibleFrom,
			DueBy:        dueBy,
		}); err != nil {
			slog.Warn("day video: seed state", "error", err)
		}
		return nil, false
	}

	since := now.Add(-utils.DayVideoVideoLookback)
	recentVideos, err := utils.CountRecentVideoPosts(db, ctx.GroupID, since)
	if err != nil {
		slog.Warn("day video: count recent videos", "error", err)
		return nil, false
	}

	if !utils.IsGroupActive(recentVideos) {
		return nil, false
	}
	if !utils.InDayVideoPostWindow(now) {
		return nil, false
	}
	if now.Before(state.EligibleFrom) {
		return nil, false
	}

	shouldPost := utils.EvaluateDayVideoPost(now, *state, s.rand)
	if !shouldPost {
		state.LastCheckedAt = sql.NullTime{Time: now, Valid: true}
		if err := utils.UpsertDayVideoState(db, *state); err != nil {
			slog.Warn("day video: record check", "error", err)
		}
		return nil, false
	}

	prevState := *state
	eligibleFrom, dueBy := utils.NextSchedule(now, s.rand)
	if err := utils.UpsertDayVideoState(db, utils.DayVideoStateRow{
		GroupID:      ctx.GroupID,
		LastPostedAt: sql.NullTime{Time: now, Valid: true},
		EligibleFrom: eligibleFrom,
		DueBy:        dueBy,
	}); err != nil {
		slog.Warn("day video: reserve state", "error", err)
		return nil, false
	}

	return &postAttempt{prevState: prevState, at: now}, true
}

func (s *Scheduler) tryPost(ctx PostContext) {
	lock := s.groupLock(ctx.GroupID)
	lock.Lock()
	defer lock.Unlock()

	attempt, ok := s.reservePostIfDue(ctx)
	if !ok {
		return
	}

	videoPath, err := s.generateVideo(attempt.at.UTC())
	if err != nil {
		s.restoreDayVideoState(attempt.prevState)
		slog.Warn("day video: generate", "error", err)
		return
	}
	defer os.Remove(videoPath)

	if err := s.poster.PostDayVideo(ctx, videoPath); err != nil {
		s.restoreDayVideoState(attempt.prevState)
		slog.Warn("day video: post", "error", err)
		return
	}

	slog.Info("day video posted", "groupId", ctx.GroupID)
}

func (s *Scheduler) restoreDayVideoState(state utils.DayVideoStateRow) {
	db, err := utils.OpenStatsDB(s.dbPath)
	if err != nil {
		slog.Error("day video: open db to restore state", "error", err, "groupId", state.GroupID)
		return
	}
	defer db.Close()

	if err := utils.UpsertDayVideoState(db, state); err != nil {
		slog.Error("day video: failed to restore state after aborted post", "error", err, "groupId", state.GroupID)
	}
}

func (s *Scheduler) groupLock(groupID string) *sync.Mutex {
	lock, _ := s.groupLocks.LoadOrStore(groupID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// SetGenerateVideoForTest overrides video generation in tests.
func (s *Scheduler) SetGenerateVideoForTest(fn func(time.Time) (string, error)) {
	s.generateVideo = fn
}

func parseGroupID(groupID string) (platform, chatID string, err error) {
	platform, chatID, ok := strings.Cut(groupID, ":")
	if !ok || platform == "" || chatID == "" {
		return "", "", fmt.Errorf("invalid group id %q", groupID)
	}
	return platform, chatID, nil
}
