package utils

import (
	"math/rand"
	"time"
)

const (
	dayVideoMinInterval         = 14 * 24 * time.Hour
	dayVideoMaxInterval         = 28 * 24 * time.Hour
	dayVideoMinMembers          = 3
	dayVideoMinVideos7d         = 7
	dayVideoActivityWindow      = 24 * time.Hour
	dayVideoVideoActivityWindow = 7 * 24 * time.Hour
	dayVideoPostWindowStartHour = 3
	dayVideoPostWindowEndHour   = 12
)

// DayVideoVideoLookback is the rolling window for counting bot video posts.
const DayVideoVideoLookback = dayVideoVideoActivityWindow

// ScheduleClock provides the current time for scheduling logic.
type ScheduleClock interface {
	Now() time.Time
}

// ScheduleRand provides randomness for scheduling logic.
type ScheduleRand interface {
	Float64() float64
}

// NewScheduleRand returns a seeded random source for production use.
func NewScheduleRand() ScheduleRand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GroupActivitySnapshot is input for active-group checks.
type GroupActivitySnapshot struct {
	IsGroupChat   bool
	MemberCount   int
	LastMessageAt time.Time
}

// InDayVideoPostWindow reports whether now is within 03:00–11:59 UTC.
func InDayVideoPostWindow(now time.Time) bool {
	hour := now.UTC().Hour()
	return hour >= dayVideoPostWindowStartHour && hour < dayVideoPostWindowEndHour
}

// IsGroupActive checks hardcoded eligibility gates for day meme posting.
func IsGroupActive(now time.Time, activity GroupActivitySnapshot, recentVideoCount int) bool {
	if !activity.IsGroupChat {
		return false
	}
	if activity.MemberCount < dayVideoMinMembers {
		return false
	}
	if now.Sub(activity.LastMessageAt) > dayVideoActivityWindow {
		return false
	}
	if recentVideoCount < dayVideoMinVideos7d {
		return false
	}
	return true
}

// PostProbability returns the chance of posting at now given the schedule window.
func PostProbability(now, eligibleFrom, dueBy time.Time) float64 {
	if now.Before(eligibleFrom) {
		return 0
	}
	if !now.Before(dueBy) {
		return 1
	}
	window := dueBy.Sub(eligibleFrom)
	if window <= 0 {
		return 1
	}
	progress := float64(now.Sub(eligibleFrom)) / float64(window)
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return 0.05 + progress*0.75
}

// ShouldPostDayVideo decides whether to post given schedule, activity, and a random roll.
func ShouldPostDayVideo(now time.Time, state DayVideoStateRow, activity GroupActivitySnapshot, recentVideoCount int, rng ScheduleRand) bool {
	if !IsGroupActive(now, activity, recentVideoCount) {
		return false
	}
	if !InDayVideoPostWindow(now) {
		return false
	}
	if now.Before(state.EligibleFrom) {
		return false
	}
	p := PostProbability(now, state.EligibleFrom, state.DueBy)
	return rng.Float64() < p
}

// NextSchedule computes the next eligible window after a reference time.
// due_by is uniform in [14d, 28d]; eligible_from is uniform in [14d, due_by].
func NextSchedule(from time.Time, rng ScheduleRand) (eligibleFrom, dueBy time.Time) {
	span := dayVideoMaxInterval - dayVideoMinInterval
	total := dayVideoMinInterval + time.Duration(rng.Float64()*float64(span))
	dueBy = from.Add(total)
	innerSpan := total - dayVideoMinInterval
	eligibleFrom = from.Add(dayVideoMinInterval + time.Duration(rng.Float64()*float64(innerSpan)))
	return eligibleFrom, dueBy
}

// InitialSchedule computes the first schedule for a newly seen group.
func InitialSchedule(now time.Time, rng ScheduleRand) (eligibleFrom, dueBy time.Time) {
	return NextSchedule(now, rng)
}
