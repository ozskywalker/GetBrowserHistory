package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogProgress(t *testing.T) {
	// logProgress writes a timestamped line to stdout; verify it runs without panic.
	logProgress("test %s %d", "message", 42)
}

func TestLogWarn(t *testing.T) {
	orig := warnings
	warnings = nil
	defer func() { warnings = orig }()

	logWarn("test %s %d", "warning", 7)

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0] != "test warning 7" {
		t.Errorf("got %q, want %q", warnings[0], "test warning 7")
	}
}

func TestCreateOutputDir(t *testing.T) {
	t.Run("explicit path is created and returned", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "reports", "sub")
		got, err := createOutputDir(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != dir {
			t.Errorf("got %q, want %q", got, dir)
		}
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("directory not created: %v", err)
		}
	})

	t.Run("existing directory succeeds", func(t *testing.T) {
		dir := t.TempDir()
		got, err := createOutputDir(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != dir {
			t.Errorf("got %q, want %q", got, dir)
		}
	})

	t.Run("empty path generates timestamped default under Windows Temp", func(t *testing.T) {
		got, err := createOutputDir("")
		if err != nil {
			t.Skipf("default path not writable (skipping): %v", err)
		}
		defer func() { _ = os.RemoveAll(got) }()
		if !strings.Contains(got, "BrowserReport_") {
			t.Errorf("default path should contain BrowserReport_, got %q", got)
		}
	})
}

func TestCurrentUser(t *testing.T) {
	t.Run("returns empty string when USERNAME not set", func(t *testing.T) {
		t.Setenv("USERNAME", "")
		got := currentUser()
		if got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	t.Run("returns USERNAME when USERDOMAIN is empty", func(t *testing.T) {
		t.Setenv("USERNAME", "alice")
		t.Setenv("USERDOMAIN", "")
		got := currentUser()
		if got != "alice" {
			t.Errorf("got %q, want %q", got, "alice")
		}
	})

	t.Run("returns DOMAIN\\USERNAME when domain differs from computer name", func(t *testing.T) {
		t.Setenv("USERNAME", "bob")
		t.Setenv("USERDOMAIN", "CORP")
		t.Setenv("COMPUTERNAME", "WORKSTATION")
		got := currentUser()
		if got != `CORP\bob` {
			t.Errorf("got %q, want %q", got, `CORP\bob`)
		}
	})

	t.Run("returns USERNAME when domain matches computer name", func(t *testing.T) {
		t.Setenv("USERNAME", "carol")
		t.Setenv("USERDOMAIN", "MYPC")
		t.Setenv("COMPUTERNAME", "MYPC")
		got := currentUser()
		if got != "carol" {
			t.Errorf("got %q, want %q", got, "carol")
		}
	})

	t.Run("domain comparison is case-insensitive", func(t *testing.T) {
		t.Setenv("USERNAME", "dave")
		t.Setenv("USERDOMAIN", "mypc")
		t.Setenv("COMPUTERNAME", "MYPC")
		got := currentUser()
		if got != "dave" {
			t.Errorf("case-insensitive domain: got %q, want %q", got, "dave")
		}
	})
}
