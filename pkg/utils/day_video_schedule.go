package utils

import (
	"math/rand"
	"time"
)

const (
	dayVideoMinInterval         = 14 * 24 * time.Hour
	dayVideoMaxInterval         = 42 * 24 * time.Hour
	dayVideoMinVideos7d         = 7
	dayVideoPostWindowStartHour = 3
	dayVideoPostWindowEndHour   = 12
)

// DayVideoVideoLookback is the rolling window for counting bot video posts.
const DayVideoVideoLookback = 7 * 24 * time.Hour

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

// InDayVideoPostWindow reports whether now is within 03:00–11:59 UTC.
func InDayVideoPostWindow(now time.Time) bool {
	hour := now.UTC().Hour()
	return hour >= dayVideoPostWindowStartHour && hour < dayVideoPostWindowEndHour
}

// IsGroupActive reports whether a group has enough recent bot video activity to
// be eligible for the day meme easter egg. Bot video posts (recorded for
// reaction tracking) are the sole activity signal; only group-chat posts count.
func IsGroupActive(recentVideoCount int) bool {
	return recentVideoCount >= dayVideoMinVideos7d
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// dayVideoWindowProgress returns how far t is through [eligibleFrom, dueBy], in [0, 1].
func dayVideoWindowProgress(t, eligibleFrom, dueBy time.Time) float64 {
	window := dueBy.Sub(eligibleFrom)
	if window <= 0 {
		return 1
	}
	return clamp01(float64(t.Sub(eligibleFrom)) / float64(window))
}

// dayVideoTargetCDF shapes when posts land across the eligible window.
// Linear (uniform) => posts spread evenly across [eligible_from, due_by].
func dayVideoTargetCDF(u float64) float64 {
	return clamp01(u)
}

// DayVideoCheckReference returns the previous in-window check time for survival-ratio math.
func DayVideoCheckReference(state DayVideoStateRow) time.Time {
	if state.LastCheckedAt.Valid && !state.LastCheckedAt.Time.Before(state.EligibleFrom) {
		return state.LastCheckedAt.Time
	}
	return state.EligibleFrom
}

// PostProbability returns the per-check chance of posting so the posting-time
// distribution matches dayVideoTargetCDF regardless of check cadence.
func PostProbability(prev, now, eligibleFrom, dueBy time.Time) float64 {
	if now.Before(eligibleFrom) {
		return 0
	}
	if !now.Before(dueBy) {
		return 1
	}
	uPrev := dayVideoWindowProgress(prev, eligibleFrom, dueBy)
	uNow := dayVideoWindowProgress(now, eligibleFrom, dueBy)
	sPrev := 1 - dayVideoTargetCDF(uPrev)
	sNow := 1 - dayVideoTargetCDF(uNow)
	if sPrev <= 0 {
		return 1
	}
	return clamp01(1 - sNow/sPrev)
}

// EvaluateDayVideoPost rolls whether to post when schedule gates already pass.
func EvaluateDayVideoPost(now time.Time, state DayVideoStateRow, rng ScheduleRand) bool {
	prev := DayVideoCheckReference(state)
	p := PostProbability(prev, now, state.EligibleFrom, state.DueBy)
	return rng.Float64() < p
}

// NextSchedule computes the next eligible window after a reference time.
// due_by is uniform in [14d, 42d]; eligible_from is uniform in [14d, due_by].
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
