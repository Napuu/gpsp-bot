package utils

import (
	"math/rand"
	"testing"
	"time"
)

func TestSimulateGroupPostsMeanInterval(t *testing.T) {
	const cycles = 500
	seeds := []int64{1, 2, 3}

	const day = 24 * time.Hour
	const hour = time.Hour

	for _, seed := range seeds {
		seed := seed
		t.Run(time.Unix(seed, 0).Format("seed"), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			now := utcAt(2020, time.January, 1, 3, 0)
			lastPosted := now
			eligibleFrom, dueBy := NextSchedule(lastPosted, rng)

			var intervals []time.Duration
			checksWithoutPost := 0

			end := now.Add(time.Duration(cycles) * 35 * day)
			for now.Before(end) {
				if InDayVideoPostWindow(now) && !now.Before(eligibleFrom) {
					state := DayVideoStateRow{EligibleFrom: eligibleFrom, DueBy: dueBy}
					activity := GroupActivitySnapshot{
						IsGroupChat:   true,
						MemberCount:   10,
						LastMessageAt: now,
					}
					if ShouldPostDayVideo(now, state, activity, 7, rng) {
						intervals = append(intervals, now.Sub(lastPosted))
						lastPosted = now
						eligibleFrom, dueBy = NextSchedule(lastPosted, rng)
						checksWithoutPost = 0
					} else if now.After(dueBy.Add(24 * time.Hour)) {
						t.Fatalf("missed forced post by %v after dueBy %v", now, dueBy)
					} else if now.After(dueBy) {
						checksWithoutPost++
						if checksWithoutPost > 24 {
							t.Fatalf("expected forced post within 24h after dueBy %v at now %v", dueBy, now)
						}
					}
				}
				now = now.Add(hour)
			}

			if len(intervals) < 10 {
				t.Fatalf("too few simulated posts: %d", len(intervals))
			}

			var total time.Duration
			for _, iv := range intervals {
				if iv < dayVideoMinInterval-hour {
					t.Fatalf("interval %v shorter than min interval", iv)
				}
				if iv > dayVideoMaxInterval+25*hour {
					t.Fatalf("interval %v longer than max interval + grace", iv)
				}
				total += iv
			}
			mean := total / time.Duration(len(intervals))
			if mean < 18*day || mean > 24*day {
				t.Fatalf("mean interval %v, want ~21d (18-24d)", mean)
			}
		})
	}
}
