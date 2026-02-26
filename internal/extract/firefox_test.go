package extract

import (
	"testing"
	"time"
)

func TestPRTimeToTime(t *testing.T) {
	// PRTime is microseconds since the Unix epoch (unlike WebKit which uses 1601).
	tests := []struct {
		name  string
		input int64
		want  time.Time
	}{
		{
			name:  "zero returns zero time",
			input: 0,
			want:  time.Time{},
		},
		{
			// 1 second past the Unix epoch.
			name:  "1 second past unix epoch",
			input: 1_000_000,
			want:  time.Unix(1, 0).UTC(),
		},
		{
			// 2024-01-01 00:00:00 UTC = Unix 1704067200 seconds.
			name:  "2024-01-01 00:00:00 UTC",
			input: 1704067200 * 1_000_000,
			want:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			// Sub-second precision: 500 ms past the Unix epoch.
			name:  "sub-second precision",
			input: 500_000, // 0.5 seconds in µs
			want:  time.Unix(0, 500_000*1_000).UTC(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PRTimeToTime(tc.input)
			if !got.Equal(tc.want) {
				t.Errorf("PRTimeToTime(%d) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
