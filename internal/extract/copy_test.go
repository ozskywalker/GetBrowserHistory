package extract

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDatabase(t *testing.T) {
	t.Run("copies main database file correctly", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "History")
		content := []byte("SQLite format 3\x00fake-db-content")
		if err := os.WriteFile(src, content, 0644); err != nil {
			t.Fatal(err)
		}

		tempDir, destPath, err := CopyDatabase(src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer os.RemoveAll(tempDir)

		got, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("cannot read dest file: %v", err)
		}
		if !bytes.Equal(got, content) {
			t.Errorf("dest content does not match source: got %q, want %q", got, content)
		}
	})

	t.Run("destination filename matches source filename", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "places.sqlite")
		if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}

		tempDir, destPath, err := CopyDatabase(src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer os.RemoveAll(tempDir)

		if filepath.Base(destPath) != "places.sqlite" {
			t.Errorf("expected dest filename 'places.sqlite', got %q", filepath.Base(destPath))
		}
	})

	t.Run("WAL sidecar is copied when present", func(t *testing.T) {
		srcDir := t.TempDir()
		src := filepath.Join(srcDir, "History")
		wal := src + "-wal"
		walContent := []byte("wal-data")

		if err := os.WriteFile(src, []byte("db"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(wal, walContent, 0644); err != nil {
			t.Fatal(err)
		}

		tempDir, destPath, err := CopyDatabase(src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer os.RemoveAll(tempDir)

		copiedWAL := destPath + "-wal"
		got, err := os.ReadFile(copiedWAL)
		if err != nil {
			t.Fatalf("WAL sidecar not copied: %v", err)
		}
		if !bytes.Equal(got, walContent) {
			t.Errorf("WAL content mismatch: got %q, want %q", got, walContent)
		}
	})

	t.Run("missing WAL sidecar does not cause error", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "History")
		if err := os.WriteFile(src, []byte("db"), 0644); err != nil {
			t.Fatal(err)
		}

		// No -wal or -shm files exist.
		tempDir, _, err := CopyDatabase(src)
		if err != nil {
			t.Fatalf("expected no error when sidecars absent, got: %v", err)
		}
		defer os.RemoveAll(tempDir)
	})

	t.Run("non-existent source returns error and does not leak temp dir", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "nonexistent.db")

		tempDir, _, err := CopyDatabase(src)
		if err == nil {
			t.Error("expected error for non-existent source, got nil")
			// Clean up just in case.
			if tempDir != "" {
				os.RemoveAll(tempDir)
			}
			return
		}
		// The function must clean up its temp dir on error.
		if tempDir != "" {
			if _, statErr := os.Stat(tempDir); statErr == nil {
				t.Error("temp dir still exists after failed CopyDatabase")
			}
		}
	})
}
