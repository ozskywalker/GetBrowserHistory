package report

import (
	"strings"
	"testing"
	"time"

	"github.com/ozskywalker/GetBrowserHistory/internal/extract"
)

func TestRenderHTML(t *testing.T) {
	t.Run("empty report renders without error", func(t *testing.T) {
		b, err := RenderHTML(Report{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(b), "Browser History Report") {
			t.Error("expected report title in output")
		}
	})

	t.Run("meta fields appear in header", func(t *testing.T) {
		r := Report{
			Meta: ReportMeta{
				Hostname:         "forensic-workstation",
				ExecutingAccount: "investigator",
				Version:          "1.0.0",
				GeneratedAt:      time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
			},
		}
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(b)
		for _, want := range []string{
			"forensic-workstation",
			"investigator",
			"1.0.0",
			"2024-06-01",
		} {
			if !strings.Contains(s, want) {
				t.Errorf("expected %q in HTML output", want)
			}
		}
	})

	t.Run("search history section shown when searches present", func(t *testing.T) {
		r := reportWithSearchHit()
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(b)
		if !strings.Contains(s, "Search History") {
			t.Error("expected 'Search History' section when search records exist")
		}
		if !strings.Contains(s, "golang tutorial") {
			t.Error("expected search query text in output")
		}
	})

	t.Run("search history section absent when no searches", func(t *testing.T) {
		r := Report{
			Users: []UserData{
				{
					Username: "alice",
					Profiles: []ProfileData{
						{
							BrowserName: "Chrome",
							ProfilePath: `/home/alice/chrome/Default`,
							History: []extract.HistoryRecord{
								{
									URL:   "https://github.com",
									Title: "GitHub",
								},
							},
						},
					},
				},
			},
		}
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(b), "Search History") {
			t.Error("expected no 'Search History' section when no search records")
		}
	})

	t.Run("profile path uses base component via base() template func", func(t *testing.T) {
		r := Report{
			Users: []UserData{
				{
					Username: "alice",
					Profiles: []ProfileData{
						{
							BrowserName: "Chrome",
							ProfilePath: `/Users/alice/AppData/Local/Google/Chrome/User Data/Default`,
						},
					},
				},
			},
		}
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The template renders "Chrome — Default" (base of the ProfilePath).
		if !strings.Contains(string(b), "Chrome") {
			t.Error("expected browser name in profile header")
		}
		if !strings.Contains(string(b), "Default") {
			t.Error("expected profile path base name in header")
		}
	})

	t.Run("truncation notice present when Truncated is true", func(t *testing.T) {
		r := Report{
			Users: []UserData{
				{
					Username: "alice",
					Profiles: []ProfileData{
						{
							BrowserName: "Chrome",
							ProfilePath: `/chrome/Default`,
							Truncated:   true,
							TruncatedAt: 10000,
							History:     []extract.HistoryRecord{{URL: "https://example.com"}},
						},
					},
				},
			},
		}
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(b), "10000") {
			t.Error("expected truncation row count in output")
		}
		if !strings.Contains(string(b), "truncation-notice") {
			t.Error("expected truncation-notice CSS class in output")
		}
	})

	t.Run("download records appear in output", func(t *testing.T) {
		r := Report{
			Users: []UserData{
				{
					Username: "alice",
					Profiles: []ProfileData{
						{
							BrowserName: "Firefox",
							ProfilePath: `/firefox/default`,
							Downloads: []extract.DownloadRecord{
								{
									TargetPath: `/Users/alice/Downloads/report.pdf`,
									SourceURL:  "https://example.com/report.pdf",
									TotalBytes: 2048,
								},
							},
						},
					},
				},
			},
		}
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(b)
		if !strings.Contains(s, "report.pdf") {
			t.Error("expected download filename in output")
		}
		if !strings.Contains(s, "example.com") {
			t.Error("expected download source URL in output")
		}
	})

	t.Run("warnings section present when warnings set", func(t *testing.T) {
		r := Report{Warnings: []string{"could not read profile: access denied"}}
		b, err := RenderHTML(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(b), "access denied") {
			t.Error("expected warning text in output")
		}
	})
}

func reportWithSearchHit() Report {
	return Report{
		Users: []UserData{
			{
				Username: "alice",
				Profiles: []ProfileData{
					{
						BrowserName: "Chrome",
						ProfilePath: `/chrome/Default`,
						History: []extract.HistoryRecord{
							{
								URL:          "https://www.youtube.com/results?search_query=golang+tutorial",
								Title:        "golang tutorial - YouTube",
								VisitCount:   1,
								LastVisitUTC: time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC),
								SearchQuery:  "golang tutorial",
								SearchEngine: "YouTube",
							},
						},
					},
				},
			},
		},
	}
}
