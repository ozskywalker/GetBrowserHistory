package extract

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractChromiumHistory(t *testing.T) {
	t.Run("missing History file returns error", func(t *testing.T) {
		_, err := ExtractChromiumHistory(t.TempDir(), 100)
		if err == nil {
			t.Fatal("expected error for missing History file")
		}
	})

	t.Run("returns records from valid database", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := webKitEpoch + int64(1704067200)*1_000_000 // 2024-01-01 00:00:00 UTC
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE urls (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER, visit_time INTEGER)`,
			`INSERT INTO urls VALUES (1, 'https://example.com', 'Example', 5)`,
			fmt.Sprintf(`INSERT INTO visits VALUES (1, 1, %d)`, ts),
		)

		records, err := ExtractChromiumHistory(profileDir, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		r := records[0]
		if r.URL != "https://example.com" {
			t.Errorf("URL: got %q, want %q", r.URL, "https://example.com")
		}
		if r.Title != "Example" {
			t.Errorf("Title: got %q, want %q", r.Title, "Example")
		}
		if r.VisitCount != 5 {
			t.Errorf("VisitCount: got %d, want 5", r.VisitCount)
		}
		want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if !r.LastVisitUTC.Equal(want) {
			t.Errorf("LastVisitUTC: got %v, want %v", r.LastVisitUTC, want)
		}
	})

	t.Run("NULL title becomes empty string", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := webKitEpoch + int64(1704067200)*1_000_000
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE urls (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER, visit_time INTEGER)`,
			`INSERT INTO urls VALUES (1, 'https://example.com', NULL, 1)`,
			fmt.Sprintf(`INSERT INTO visits VALUES (1, 1, %d)`, ts),
		)

		records, err := ExtractChromiumHistory(profileDir, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		if records[0].Title != "" {
			t.Errorf("NULL title: got %q, want empty string", records[0].Title)
		}
	})

	t.Run("search URL populates SearchQuery and SearchEngine", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := webKitEpoch + int64(1704067200)*1_000_000
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE urls (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER, visit_time INTEGER)`,
			`INSERT INTO urls VALUES (1, 'https://www.google.com/search?q=golang', 'Google', 1)`,
			fmt.Sprintf(`INSERT INTO visits VALUES (1, 1, %d)`, ts),
		)

		records, err := ExtractChromiumHistory(profileDir, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		if records[0].SearchQuery != "golang" {
			t.Errorf("SearchQuery: got %q, want %q", records[0].SearchQuery, "golang")
		}
		if records[0].SearchEngine != "Google" {
			t.Errorf("SearchEngine: got %q, want %q", records[0].SearchEngine, "Google")
		}
	})

	t.Run("maxRows limits result count", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := webKitEpoch + int64(1704067200)*1_000_000
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE urls (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER, visit_time INTEGER)`,
			`INSERT INTO urls VALUES (1,'https://a.com','A',1),(2,'https://b.com','B',1),(3,'https://c.com','C',1)`,
			fmt.Sprintf(`INSERT INTO visits VALUES (1,1,%d),(2,2,%d),(3,3,%d)`, ts, ts-1, ts-2),
		)

		records, err := ExtractChromiumHistory(profileDir, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Errorf("with maxRows=2 want 2 records, got %d", len(records))
		}
	})

	t.Run("empty history returns empty slice without error", func(t *testing.T) {
		profileDir := t.TempDir()
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE urls (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER, visit_time INTEGER)`,
		)

		records, err := ExtractChromiumHistory(profileDir, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}
	})
}

func TestExtractChromiumDownloads(t *testing.T) {
	t.Run("missing History file returns error", func(t *testing.T) {
		_, err := ExtractChromiumDownloads(t.TempDir())
		if err == nil {
			t.Fatal("expected error for missing History file")
		}
	})

	t.Run("returns download records from valid database", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := webKitEpoch + int64(1704067200)*1_000_000
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE downloads (id INTEGER PRIMARY KEY, target_path TEXT, start_time INTEGER, total_bytes INTEGER, mime_type TEXT)`,
			`CREATE TABLE downloads_url_chains (id INTEGER, url TEXT)`,
			fmt.Sprintf(`INSERT INTO downloads VALUES (1, 'C:\Users\alice\Downloads\file.zip', %d, 2048, 'application/zip')`, ts),
			`INSERT INTO downloads_url_chains VALUES (1, 'https://example.com/file.zip')`,
		)

		records, err := ExtractChromiumDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		r := records[0]
		if r.TargetPath != `C:\Users\alice\Downloads\file.zip` {
			t.Errorf("TargetPath: got %q", r.TargetPath)
		}
		if r.SourceURL != "https://example.com/file.zip" {
			t.Errorf("SourceURL: got %q", r.SourceURL)
		}
		if r.MimeType != "application/zip" {
			t.Errorf("MimeType: got %q", r.MimeType)
		}
		if r.TotalBytes != 2048 {
			t.Errorf("TotalBytes: got %d, want 2048", r.TotalBytes)
		}
		want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if !r.StartTimeUTC.Equal(want) {
			t.Errorf("StartTimeUTC: got %v, want %v", r.StartTimeUTC, want)
		}
	})

	t.Run("download without URL chain has empty SourceURL", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := webKitEpoch + int64(1704067200)*1_000_000
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE downloads (id INTEGER PRIMARY KEY, target_path TEXT, start_time INTEGER, total_bytes INTEGER, mime_type TEXT)`,
			`CREATE TABLE downloads_url_chains (id INTEGER, url TEXT)`,
			fmt.Sprintf(`INSERT INTO downloads VALUES (1, 'C:\file.exe', %d, 512, NULL)`, ts),
		)

		records, err := ExtractChromiumDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		if records[0].SourceURL != "" {
			t.Errorf("SourceURL: got %q, want empty (no URL chain)", records[0].SourceURL)
		}
		if records[0].MimeType != "" {
			t.Errorf("MimeType: got %q, want empty (was NULL)", records[0].MimeType)
		}
	})

	t.Run("no downloads returns empty slice without error", func(t *testing.T) {
		profileDir := t.TempDir()
		buildTestDB(t, filepath.Join(profileDir, "History"),
			`CREATE TABLE downloads (id INTEGER PRIMARY KEY, target_path TEXT, start_time INTEGER, total_bytes INTEGER, mime_type TEXT)`,
			`CREATE TABLE downloads_url_chains (id INTEGER, url TEXT)`,
		)

		records, err := ExtractChromiumDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}
	})
}
