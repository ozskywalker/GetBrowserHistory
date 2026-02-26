package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ozskywalker/GetBrowserHistory/internal/extract"
)

func TestRenderJSON(t *testing.T) {
	t.Run("empty report serialises without error", func(t *testing.T) {
		r := Report{}
		b, err := RenderJSON(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(b) == 0 {
			t.Error("expected non-empty JSON output")
		}
		var out map[string]any
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
	})

	t.Run("meta fields serialise with correct JSON keys", func(t *testing.T) {
		ts := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		r := Report{
			Meta: ReportMeta{
				GeneratedAt:      ts,
				Hostname:         "test-host",
				ExecutingAccount: "test-account",
				Version:          "1.2.3",
			},
		}
		b, err := RenderJSON(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(b)
		for _, want := range []string{
			`"generatedAt"`,
			`"hostname": "test-host"`,
			`"executingAccount": "test-account"`,
			`"scriptVersion": "1.2.3"`,
		} {
			if !strings.Contains(s, want) {
				t.Errorf("output missing %q", want)
			}
		}
	})

	t.Run("time.Time serialises as RFC 3339 UTC", func(t *testing.T) {
		ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		r := Report{Meta: ReportMeta{GeneratedAt: ts}}
		b, err := RenderJSON(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(b), "2024-01-01T00:00:00Z") {
			t.Errorf("expected RFC 3339 timestamp in output, got: %s", b)
		}
	})

	t.Run("history and download records appear in output", func(t *testing.T) {
		ts := time.Date(2024, 3, 10, 8, 0, 0, 0, time.UTC)
		r := Report{
			Users: []UserData{
				{
					Username: "alice",
					Profiles: []ProfileData{
						{
							BrowserName: "Chrome",
							ProfilePath: `C:\Users\alice\AppData\Local\Google\Chrome\User Data\Default`,
							History: []extract.HistoryRecord{
								{
									URL:          "https://google.com/search?q=test",
									Title:        "test - Google Search",
									VisitCount:   3,
									LastVisitUTC: ts,
									SearchQuery:  "test",
									SearchEngine: "Google",
								},
							},
							Downloads: []extract.DownloadRecord{
								{
									TargetPath:   `C:\Users\alice\Downloads\file.zip`,
									SourceURL:    "https://example.com/file.zip",
									TotalBytes:   1024,
									StartTimeUTC: ts,
								},
							},
						},
					},
				},
			},
		}
		b, err := RenderJSON(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(b)
		for _, want := range []string{
			`"username": "alice"`,
			`"browserName": "Chrome"`,
			`"searchQuery": "test"`,
			`"searchEngine": "Google"`,
			`file.zip`,
		} {
			if !strings.Contains(s, want) {
				t.Errorf("output missing %q", want)
			}
		}
	})

	t.Run("Truncated omitempty: false and zero omitted", func(t *testing.T) {
		r := Report{
			Users: []UserData{
				{
					Profiles: []ProfileData{
						{Truncated: false, TruncatedAt: 0},
					},
				},
			},
		}
		b, err := RenderJSON(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(b), `"truncated"`) {
			t.Error("expected 'truncated' to be omitted when false")
		}
		if strings.Contains(string(b), `"truncatedAt"`) {
			t.Error("expected 'truncatedAt' to be omitted when zero")
		}
	})

	t.Run("Truncated fields present when set", func(t *testing.T) {
		r := Report{
			Users: []UserData{
				{
					Profiles: []ProfileData{
						{Truncated: true, TruncatedAt: 5000},
					},
				},
			},
		}
		b, err := RenderJSON(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(b)
		if !strings.Contains(s, `"truncated": true`) {
			t.Error("expected 'truncated' to be present when true")
		}
		if !strings.Contains(s, `"truncatedAt": 5000`) {
			t.Error("expected 'truncatedAt' to be present when non-zero")
		}
	})

	t.Run("warnings omitted when empty, present when set", func(t *testing.T) {
		b, _ := RenderJSON(Report{})
		if strings.Contains(string(b), `"warnings"`) {
			t.Error("expected 'warnings' to be omitted when nil")
		}

		b, _ = RenderJSON(Report{Warnings: []string{"something went wrong"}})
		if !strings.Contains(string(b), `"something went wrong"`) {
			t.Error("expected warning message in output")
		}
	})
}
