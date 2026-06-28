package utils

import (
	"database/sql"
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
	tests := []struct {
		name   string
		videos int
		want   bool
	}{
		{"enough videos", 7, true},
		{"more than enough", 20, true},
		{"too few videos", 6, false},
		{"no videos", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGroupActive(tt.videos); got != tt.want {
				t.Fatalf("IsGroupActive(%d) = %v, want %v", tt.videos, got, tt.want)
			}
		})
	}
}

func TestDayVideoCheckReference(t *testing.T) {
	eligible := utcAt(2026, time.June, 1, 0, 0)
	checked := eligible.Add(2 * 24 * time.Hour)

	state := DayVideoStateRow{EligibleFrom: eligible}
	if got := DayVideoCheckReference(state); !got.Equal(eligible) {
		t.Fatalf("without last_checked = %v, want eligible %v", got, eligible)
	}

	state.LastCheckedAt = sql.NullTime{Time: checked, Valid: true}
	if got := DayVideoCheckReference(state); !got.Equal(checked) {
		t.Fatalf("with last_checked = %v, want %v", got, checked)
	}

	stale := eligible.Add(-time.Hour)
	state.LastCheckedAt = sql.NullTime{Time: stale, Valid: true}
	if got := DayVideoCheckReference(state); !got.Equal(eligible) {
		t.Fatalf("stale last_checked = %v, want eligible %v", got, eligible)
	}
}

func TestPostProbability(t *testing.T) {
	eligible := utcAt(2026, time.June, 1, 0, 0)
	dueBy := eligible.Add(14 * 24 * time.Hour)

	if p := PostProbability(eligible.Add(-time.Hour), eligible, eligible, dueBy); p != 0 {
		t.Fatalf("before eligible = %v, want 0", p)
	}
	if p := PostProbability(dueBy, eligible, eligible, dueBy); p != 1 {
		t.Fatalf("at dueBy = %v, want 1", p)
	}
	if p := PostProbability(eligible, eligible, eligible, dueBy); p != 0 {
		t.Fatalf("first check at eligible = %v, want 0", p)
	}

	mid := eligible.Add(7 * 24 * time.Hour)
	if p := PostProbability(eligible, mid, eligible, dueBy); p < 0.49 || p > 0.51 {
		t.Fatalf("half-window probability = %v, want ~0.5", p)
	}

	latePrev := eligible.Add(6 * 24 * time.Hour)
	if p := PostProbability(latePrev, mid, eligible, dueBy); p < 0.12 || p > 0.13 {
		t.Fatalf("incremental probability = %v, want ~0.125", p)
	}
}

// TestEvaluateDayVideoPostForcedAfterDueBy verifies that the production decision
// core forces a post once now is past due_by, regardless of the random roll.
func TestEvaluateDayVideoPostForcedAfterDueBy(t *testing.T) {
	now := utcAt(2026, time.June, 10, 8, 0)
	state := DayVideoStateRow{
		EligibleFrom: utcAt(2026, time.May, 20, 0, 0),
		DueBy:        utcAt(2026, time.June, 1, 0, 0),
	}
	rng := &fixedRand{values: []float64{0.99}}

	// Preconditions the production gates enforce before EvaluateDayVideoPost runs.
	if !InDayVideoPostWindow(now) {
		t.Fatal("precondition: now should be in the post window")
	}
	if now.Before(state.EligibleFrom) {
		t.Fatal("precondition: now should be at or after eligible_from")
	}

	if !EvaluateDayVideoPost(now, state, rng) {
		t.Fatal("expected forced post after due_by")
	}
}

// TestIsGroupActiveThreshold verifies the real activity gate production calls:
// a recent video count below the threshold makes a group ineligible.
func TestIsGroupActiveThreshold(t *testing.T) {
	if IsGroupActive(6) {
		t.Fatal("expected group with 6 recent videos to be inactive")
	}
	if !IsGroupActive(7) {
		t.Fatal("expected group with 7 recent videos to be active")
	}
}

// TestEvaluateDayVideoPostLowProbabilityAtStart verifies the production decision
// core does not post near the very start of the window with a moderate roll.
func TestEvaluateDayVideoPostLowProbabilityAtStart(t *testing.T) {
	eligible := utcAt(2026, time.June, 1, 3, 0)
	dueBy := eligible.Add(14 * 24 * time.Hour)
	now := eligible.Add(time.Hour)
	state := DayVideoStateRow{EligibleFrom: eligible, DueBy: dueBy}
	rng := &fixedRand{values: []float64{0.5}}

	// Preconditions the production gates enforce before EvaluateDayVideoPost runs.
	if !InDayVideoPostWindow(now) {
		t.Fatal("precondition: now should be in the post window")
	}
	if now.Before(state.EligibleFrom) {
		t.Fatal("precondition: now should be at or after eligible_from")
	}

	if EvaluateDayVideoPost(now, state, rng) {
		t.Fatal("expected no post with high roll near start of window")
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
