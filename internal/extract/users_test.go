package extract

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestEnumerateUsers(t *testing.T) {
	t.Run("returns real user dirs, skips system dirs and files", func(t *testing.T) {
		root := t.TempDir()

		// Real user directories that should be returned.
		realUsers := []string{"alice", "bob", "john.doe"}
		for _, u := range realUsers {
			if err := os.Mkdir(filepath.Join(root, u), 0755); err != nil {
				t.Fatal(err)
			}
		}

		// System directories that must be filtered out (case-insensitive).
		for _, sys := range []string{"Public", "Default", "Default User", "All Users"} {
			if err := os.Mkdir(filepath.Join(root, sys), 0755); err != nil {
				t.Fatal(err)
			}
		}

		// A plain file — must be ignored (not a directory).
		if err := os.WriteFile(filepath.Join(root, "somefile.txt"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := EnumerateUsers(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sort.Strings(got)
		sort.Strings(realUsers)
		if len(got) != len(realUsers) {
			t.Fatalf("got %v, want %v", got, realUsers)
		}
		for i := range realUsers {
			if got[i] != realUsers[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], realUsers[i])
			}
		}
	})

	t.Run("empty root returns empty slice without error", func(t *testing.T) {
		root := t.TempDir()
		got, err := EnumerateUsers(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})

	t.Run("non-existent root returns error", func(t *testing.T) {
		_, err := EnumerateUsers(filepath.Join(t.TempDir(), "does-not-exist"))
		if err == nil {
			t.Error("expected error for non-existent root, got nil")
		}
	})

	t.Run("system dir names are case-insensitively filtered", func(t *testing.T) {
		root := t.TempDir()
		// All-caps variants should still be filtered.
		for _, sys := range []string{"PUBLIC", "DEFAULT", "ALL USERS"} {
			if err := os.Mkdir(filepath.Join(root, sys), 0755); err != nil {
				t.Fatal(err)
			}
		}
		got, err := EnumerateUsers(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no users, got %v", got)
		}
	})
}

func TestResolveUser(t *testing.T) {
	t.Run("existing directory returns full path", func(t *testing.T) {
		root := t.TempDir()
		userDir := filepath.Join(root, "alice")
		if err := os.Mkdir(userDir, 0755); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveUser(root, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != userDir {
			t.Errorf("got %q, want %q", got, userDir)
		}
	})

	t.Run("non-existent user returns error", func(t *testing.T) {
		root := t.TempDir()
		_, err := ResolveUser(root, "nobody")
		if err == nil {
			t.Error("expected error for non-existent user, got nil")
		}
	})

	t.Run("path exists but is a file, not a directory", func(t *testing.T) {
		root := t.TempDir()
		filePath := filepath.Join(root, "notadir")
		if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := ResolveUser(root, "notadir")
		if err == nil {
			t.Error("expected error when path is a file, got nil")
		}
	})
}
