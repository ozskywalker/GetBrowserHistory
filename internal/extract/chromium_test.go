package extract

import (
	"testing"
	"time"
)

func TestWebKitEpochToTime(t *testing.T) {
	// webKitEpoch = 11644473600 * 1e6 microseconds between 1601-01-01 and 1970-01-01.
	const webKitEpochMicros = int64(11644473600 * 1e6)

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
			// WebKit epoch offset itself maps exactly to 1970-01-01 00:00:00 UTC.
			name:  "webkit epoch offset = unix epoch",
			input: webKitEpochMicros,
			want:  time.Unix(0, 0).UTC(),
		},
		{
			// 2024-01-01 00:00:00 UTC = Unix 1704067200 seconds.
			// WebKit value = 1704067200*1e6 + webKitEpoch.
			name:  "2024-01-01 00:00:00 UTC",
			input: webKitEpochMicros + 1704067200*1e6,
			want:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			// Sub-second precision: 500 ms past the Unix epoch.
			name:  "sub-second precision",
			input: webKitEpochMicros + 500_000, // 0.5 seconds in µs
			want:  time.Unix(0, 500_000*1_000).UTC(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := WebKitEpochToTime(tc.input)
			if !got.Equal(tc.want) {
				t.Errorf("WebKitEpochToTime(%d) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
