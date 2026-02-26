package extract

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractFirefoxHistory(t *testing.T) {
	t.Run("missing places.sqlite returns error", func(t *testing.T) {
		_, err := ExtractFirefoxHistory(t.TempDir(), 100)
		if err == nil {
			t.Fatal("expected error for missing places.sqlite")
		}
	})

	t.Run("returns records from valid database", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000 // 2024-01-01 00:00:00 UTC in PRTime
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE moz_historyvisits (id INTEGER PRIMARY KEY, place_id INTEGER, visit_date INTEGER)`,
			`INSERT INTO moz_places VALUES (1, 'https://example.com', 'Example', 4)`,
			fmt.Sprintf(`INSERT INTO moz_historyvisits VALUES (1, 1, %d)`, ts),
		)

		records, err := ExtractFirefoxHistory(profileDir, 100)
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
		if r.VisitCount != 4 {
			t.Errorf("VisitCount: got %d, want 4", r.VisitCount)
		}
		want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if !r.LastVisitUTC.Equal(want) {
			t.Errorf("LastVisitUTC: got %v, want %v", r.LastVisitUTC, want)
		}
	})

	t.Run("NULL title becomes empty string", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE moz_historyvisits (id INTEGER PRIMARY KEY, place_id INTEGER, visit_date INTEGER)`,
			`INSERT INTO moz_places VALUES (1, 'https://example.com', NULL, 1)`,
			fmt.Sprintf(`INSERT INTO moz_historyvisits VALUES (1, 1, %d)`, ts),
		)

		records, err := ExtractFirefoxHistory(profileDir, 100)
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
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE moz_historyvisits (id INTEGER PRIMARY KEY, place_id INTEGER, visit_date INTEGER)`,
			`INSERT INTO moz_places VALUES (1, 'https://www.google.com/search?q=firefox', 'Search', 1)`,
			fmt.Sprintf(`INSERT INTO moz_historyvisits VALUES (1, 1, %d)`, ts),
		)

		records, err := ExtractFirefoxHistory(profileDir, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		if records[0].SearchQuery != "firefox" {
			t.Errorf("SearchQuery: got %q, want %q", records[0].SearchQuery, "firefox")
		}
		if records[0].SearchEngine != "Google" {
			t.Errorf("SearchEngine: got %q, want %q", records[0].SearchEngine, "Google")
		}
	})

	t.Run("maxRows limits result count", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE moz_historyvisits (id INTEGER PRIMARY KEY, place_id INTEGER, visit_date INTEGER)`,
			`INSERT INTO moz_places VALUES (1,'https://a.com','A',1),(2,'https://b.com','B',1),(3,'https://c.com','C',1)`,
			fmt.Sprintf(`INSERT INTO moz_historyvisits VALUES (1,1,%d),(2,2,%d),(3,3,%d)`, ts, ts-1, ts-2),
		)

		records, err := ExtractFirefoxHistory(profileDir, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Errorf("with maxRows=2 want 2 records, got %d", len(records))
		}
	})

	t.Run("empty history returns empty slice without error", func(t *testing.T) {
		profileDir := t.TempDir()
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT, title TEXT, visit_count INTEGER)`,
			`CREATE TABLE moz_historyvisits (id INTEGER PRIMARY KEY, place_id INTEGER, visit_date INTEGER)`,
		)

		records, err := ExtractFirefoxHistory(profileDir, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}
	})
}

func TestExtractFirefoxDownloads(t *testing.T) {
	t.Run("missing places.sqlite returns error", func(t *testing.T) {
		_, err := ExtractFirefoxDownloads(t.TempDir())
		if err == nil {
			t.Fatal("expected error for missing places.sqlite")
		}
	})

	t.Run("returns download records from valid database", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT)`,
			`CREATE TABLE moz_annos (id INTEGER PRIMARY KEY, place_id INTEGER, anno_attribute_id INTEGER, content TEXT, dateAdded INTEGER)`,
			`CREATE TABLE moz_anno_attributes (id INTEGER PRIMARY KEY, name TEXT)`,
			`INSERT INTO moz_places VALUES (1, 'https://example.com/file.pdf')`,
			`INSERT INTO moz_anno_attributes VALUES (1, 'downloads/destinationFileURI')`,
			fmt.Sprintf(`INSERT INTO moz_annos VALUES (1, 1, 1, 'file:///C:/Users/alice/Downloads/file.pdf', %d)`, ts),
		)

		records, err := ExtractFirefoxDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		r := records[0]
		if r.SourceURL != "https://example.com/file.pdf" {
			t.Errorf("SourceURL: got %q", r.SourceURL)
		}
		// "file://" (7 chars) is stripped, leaving the remainder of the URI.
		if r.TargetPath != "/C:/Users/alice/Downloads/file.pdf" {
			t.Errorf("TargetPath: got %q, want %q", r.TargetPath, "/C:/Users/alice/Downloads/file.pdf")
		}
		want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if !r.StartTimeUTC.Equal(want) {
			t.Errorf("StartTimeUTC: got %v, want %v", r.StartTimeUTC, want)
		}
	})

	t.Run("file:// prefix is stripped from destination path", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT)`,
			`CREATE TABLE moz_annos (id INTEGER PRIMARY KEY, place_id INTEGER, anno_attribute_id INTEGER, content TEXT, dateAdded INTEGER)`,
			`CREATE TABLE moz_anno_attributes (id INTEGER PRIMARY KEY, name TEXT)`,
			`INSERT INTO moz_places VALUES (1, 'https://example.com/doc.txt')`,
			`INSERT INTO moz_anno_attributes VALUES (1, 'downloads/destinationFileURI')`,
			fmt.Sprintf(`INSERT INTO moz_annos VALUES (1, 1, 1, 'file:///home/user/doc.txt', %d)`, ts),
		)

		records, err := ExtractFirefoxDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		if records[0].TargetPath != "/home/user/doc.txt" {
			t.Errorf("TargetPath: got %q, want %q", records[0].TargetPath, "/home/user/doc.txt")
		}
	})

	t.Run("path longer than 7 chars without file:// prefix is kept as-is", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT)`,
			`CREATE TABLE moz_annos (id INTEGER PRIMARY KEY, place_id INTEGER, anno_attribute_id INTEGER, content TEXT, dateAdded INTEGER)`,
			`CREATE TABLE moz_anno_attributes (id INTEGER PRIMARY KEY, name TEXT)`,
			`INSERT INTO moz_places VALUES (1, 'https://example.com/')`,
			`INSERT INTO moz_anno_attributes VALUES (1, 'downloads/destinationFileURI')`,
			fmt.Sprintf(`INSERT INTO moz_annos VALUES (1, 1, 1, 'not_a_file_uri_path.txt', %d)`, ts),
		)

		records, err := ExtractFirefoxDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("want 1 record, got %d", len(records))
		}
		if records[0].TargetPath != "not_a_file_uri_path.txt" {
			t.Errorf("TargetPath: got %q, want %q", records[0].TargetPath, "not_a_file_uri_path.txt")
		}
	})

	t.Run("only downloads/destinationFileURI annotations are returned", func(t *testing.T) {
		profileDir := t.TempDir()
		ts := int64(1704067200) * 1_000_000
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT)`,
			`CREATE TABLE moz_annos (id INTEGER PRIMARY KEY, place_id INTEGER, anno_attribute_id INTEGER, content TEXT, dateAdded INTEGER)`,
			`CREATE TABLE moz_anno_attributes (id INTEGER PRIMARY KEY, name TEXT)`,
			`INSERT INTO moz_places VALUES (1, 'https://example.com/file.zip')`,
			`INSERT INTO moz_anno_attributes VALUES (1, 'downloads/destinationFileURI'), (2, 'other/annotation')`,
			fmt.Sprintf(`INSERT INTO moz_annos VALUES (1, 1, 1, 'file:///tmp/file.zip', %d), (2, 1, 2, 'ignored', %d)`, ts, ts),
		)

		records, err := ExtractFirefoxDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Errorf("want 1 record (other annotation filtered), got %d", len(records))
		}
	})

	t.Run("no downloads returns empty slice without error", func(t *testing.T) {
		profileDir := t.TempDir()
		buildTestDB(t, filepath.Join(profileDir, "places.sqlite"),
			`CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT)`,
			`CREATE TABLE moz_annos (id INTEGER PRIMARY KEY, place_id INTEGER, anno_attribute_id INTEGER, content TEXT, dateAdded INTEGER)`,
			`CREATE TABLE moz_anno_attributes (id INTEGER PRIMARY KEY, name TEXT)`,
		)

		records, err := ExtractFirefoxDownloads(profileDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}
	})
}
