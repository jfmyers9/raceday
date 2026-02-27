package cache

import (
	"testing"
	"time"
)

func TestTTLForProximity(t *testing.T) {
	baseTime := time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		now        time.Time
		eventStart time.Time
		want       time.Duration
	}{
		{
			name:       "event just started — live",
			now:        baseTime,
			eventStart: baseTime.Add(-10 * time.Minute),
			want:       0,
		},
		{
			name:       "event 3h ago — still in live window",
			now:        baseTime,
			eventStart: baseTime.Add(-3 * time.Hour),
			want:       0,
		},
		{
			name:       "event 5h59m ago — edge of live window",
			now:        baseTime,
			eventStart: baseTime.Add(-5*time.Hour - 59*time.Minute),
			want:       0,
		},
		{
			name:       "event 6h ago same day — past live window",
			now:        baseTime,
			eventStart: baseTime.Add(-6 * time.Hour),
			want:       2 * time.Minute,
		},
		{
			name:       "event yesterday — past live window",
			now:        baseTime,
			eventStart: baseTime.Add(-24 * time.Hour),
			want:       30 * time.Minute,
		},
		{
			name:       "event in 30 minutes",
			now:        baseTime,
			eventStart: baseTime.Add(30 * time.Minute),
			want:       30 * time.Second,
		},
		{
			name:       "event in exactly 1 hour",
			now:        baseTime,
			eventStart: baseTime.Add(time.Hour),
			want:       30 * time.Second,
		},
		{
			name:       "event later today (3h away)",
			now:        baseTime,
			eventStart: baseTime.Add(3 * time.Hour),
			want:       2 * time.Minute,
		},
		{
			name:       "event in 3 days",
			now:        baseTime,
			eventStart: baseTime.Add(3 * 24 * time.Hour),
			want:       10 * time.Minute,
		},
		{
			name:       "event in exactly 7 days",
			now:        baseTime,
			eventStart: baseTime.Add(7 * 24 * time.Hour),
			want:       10 * time.Minute,
		},
		{
			name:       "event in 2 weeks",
			now:        baseTime,
			eventStart: baseTime.Add(14 * 24 * time.Hour),
			want:       30 * time.Minute,
		},
		{
			name:       "event far in the past",
			now:        baseTime,
			eventStart: baseTime.Add(-30 * 24 * time.Hour),
			want:       30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := TimeNow
			TimeNow = func() time.Time { return tt.now }
			defer func() { TimeNow = orig }()

			got := TTLForProximity(tt.eventStart)
			if got != tt.want {
				t.Errorf("TTLForProximity() = %v, want %v", got, tt.want)
			}
		})
	}
}
