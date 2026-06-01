package report

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestLoadJSON(t *testing.T) {
	t.Run("round-trips through RenderJSON", func(t *testing.T) {
		ts := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		original := Report{
			Meta: ReportMeta{
				GeneratedAt:      ts,
				Hostname:         "test-host",
				ExecutingAccount: "alice",
				Version:          "1.0.0",
			},
			Users: []UserData{
				{
					Username: "alice",
					Profiles: []ProfileData{
						{
							BrowserName: "Chrome",
							ProfilePath: `C:\Users\alice\AppData\Local\Google\Chrome\User Data\Default`,
							History: []extract.HistoryRecord{
								{URL: "https://example.com", Title: "Example", VisitCount: 1, LastVisitUTC: ts},
							},
						},
					},
				},
			},
			Warnings: []string{"a warning"},
		}

		b, err := RenderJSON(original)
		if err != nil {
			t.Fatalf("RenderJSON: %v", err)
		}

		tmp := filepath.Join(t.TempDir(), "report.json")
		if err := os.WriteFile(tmp, b, 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		loaded, err := LoadJSON(tmp)
		if err != nil {
			t.Fatalf("LoadJSON: %v", err)
		}

		if loaded.Meta.Hostname != original.Meta.Hostname {
			t.Errorf("hostname: got %q, want %q", loaded.Meta.Hostname, original.Meta.Hostname)
		}
		if loaded.Meta.Version != original.Meta.Version {
			t.Errorf("version: got %q, want %q", loaded.Meta.Version, original.Meta.Version)
		}
		if !loaded.Meta.GeneratedAt.Equal(original.Meta.GeneratedAt) {
			t.Errorf("generatedAt: got %v, want %v", loaded.Meta.GeneratedAt, original.Meta.GeneratedAt)
		}
		if len(loaded.Users) != 1 || loaded.Users[0].Username != "alice" {
			t.Errorf("unexpected users: %+v", loaded.Users)
		}
		if len(loaded.Users[0].Profiles[0].History) != 1 {
			t.Error("expected 1 history record")
		}
		if len(loaded.Warnings) != 1 || loaded.Warnings[0] != "a warning" {
			t.Errorf("unexpected warnings: %v", loaded.Warnings)
		}
	})

	t.Run("error on missing file", func(t *testing.T) {
		_, err := LoadJSON(filepath.Join(t.TempDir(), "nonexistent.json"))
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("error on invalid JSON", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(tmp, []byte(`{not valid json`), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		_, err := LoadJSON(tmp)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
