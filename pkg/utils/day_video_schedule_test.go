package utils

import (
	"math/rand"
	"testing"
	"time"
)

type fixedRand struct {
	values []float64
	idx    int
}

func (r *fixedRand) Float64() float64 {
	if len(r.values) == 0 {
		return 0
	}
	v := r.values[r.idx%len(r.values)]
	r.idx++
	return v
}

func utcAt(year int, month time.Month, day, hour, min int) time.Time {
	return time.Date(year, month, day, hour, min, 0, 0, time.UTC)
}

func TestInDayVideoPostWindow(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
		want bool
	}{
		{"before window", utcAt(2026, time.June, 6, 2, 59), false},
		{"start", utcAt(2026, time.June, 6, 3, 0), true},
		{"end", utcAt(2026, time.June, 6, 11, 59), true},
		{"after window", utcAt(2026, time.June, 6, 12, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InDayVideoPostWindow(tt.at); got != tt.want {
				t.Fatalf("InDayVideoPostWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGroupActive(t *testing.T) {
	now := utcAt(2026, time.June, 6, 10, 0)
	active := GroupActivitySnapshot{
		IsGroupChat:   true,
		MemberCount:   10,
		LastMessageAt: now.Add(-1 * time.Hour),
	}

	tests := []struct {
		name   string
		act    GroupActivitySnapshot
		videos int
		want   bool
	}{
		{"all gates pass", active, 7, true},
		{"dm", GroupActivitySnapshot{IsGroupChat: false, MemberCount: 10, LastMessageAt: now}, 7, false},
		{"too few members", GroupActivitySnapshot{IsGroupChat: true, MemberCount: 2, LastMessageAt: now}, 7, false},
		{"stale messages", GroupActivitySnapshot{IsGroupChat: true, MemberCount: 10, LastMessageAt: now.Add(-25 * time.Hour)}, 7, false},
		{"too few videos", active, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGroupActive(now, tt.act, tt.videos); got != tt.want {
				t.Fatalf("IsGroupActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostProbability(t *testing.T) {
	eligible := utcAt(2026, time.June, 1, 0, 0)
	dueBy := eligible.Add(14 * 24 * time.Hour)

	if p := PostProbability(eligible.Add(-time.Hour), eligible, dueBy); p != 0 {
		t.Fatalf("before eligible = %v, want 0", p)
	}
	if p := PostProbability(dueBy, eligible, dueBy); p != 1 {
		t.Fatalf("at dueBy = %v, want 1", p)
	}
	mid := eligible.Add(7 * 24 * time.Hour)
	if p := PostProbability(mid, eligible, dueBy); p < 0.4 || p > 0.5 {
		t.Fatalf("mid progress probability = %v, want ~0.425", p)
	}
}

func TestShouldPostDayVideoForcedAfterDueBy(t *testing.T) {
	now := utcAt(2026, time.June, 10, 8, 0)
	state := DayVideoStateRow{
		EligibleFrom: utcAt(2026, time.May, 20, 0, 0),
		DueBy:        utcAt(2026, time.June, 1, 0, 0),
	}
	activity := GroupActivitySnapshot{
		IsGroupChat:   true,
		MemberCount:   10,
		LastMessageAt: now.Add(-time.Hour),
	}
	rng := &fixedRand{values: []float64{0.99}}

	if !ShouldPostDayVideo(now, state, activity, 7, rng) {
		t.Fatal("expected forced post after due_by")
	}
}

func TestShouldPostDayVideoLowProbabilityAtStart(t *testing.T) {
	eligible := utcAt(2026, time.June, 1, 3, 0)
	dueBy := eligible.Add(14 * 24 * time.Hour)
	now := eligible
	state := DayVideoStateRow{EligibleFrom: eligible, DueBy: dueBy}
	activity := GroupActivitySnapshot{
		IsGroupChat:   true,
		MemberCount:   10,
		LastMessageAt: now,
	}
	rng := &fixedRand{values: []float64{0.5}}

	if ShouldPostDayVideo(now, state, activity, 7, rng) {
		t.Fatal("expected no post with high roll at start of window")
	}
}

func TestNextScheduleBounds(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	from := utcAt(2026, time.January, 1, 0, 0)

	for i := 0; i < 100; i++ {
		eligible, dueBy := NextSchedule(from, rng)
		if eligible.Before(from.Add(dayVideoMinInterval)) {
			t.Fatalf("eligibleFrom %v before minimum interval from %v", eligible, from)
		}
		if dueBy.Before(eligible) {
			t.Fatalf("dueBy %v before eligibleFrom %v", dueBy, eligible)
		}
		span := dueBy.Sub(from)
		if span < dayVideoMinInterval || span > dayVideoMaxInterval {
			t.Fatalf("dueBy span %v outside [%v, %v]", span, dayVideoMinInterval, dayVideoMaxInterval)
		}
	}
}
