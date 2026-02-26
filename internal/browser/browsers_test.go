package browser

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// mkProfile creates a profile directory with the given marker file inside it,
// rooted at base/relPath/profileName/markerFile.
func mkProfile(t *testing.T, base, relPath, profileName, markerFile string) {
	t.Helper()
	dir := filepath.Join(base, relPath, profileName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if markerFile != "" {
		if err := os.WriteFile(filepath.Join(dir, markerFile), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFindProfiles(t *testing.T) {
	chromiumDef := BrowserDef{
		Name:         "Chrome",
		RelativePath: filepath.Join("Google", "Chrome", "User Data"),
		AppDataBase:  AppDataLocal,
		Type:         BrowserChromium,
	}

	firefoxDef := BrowserDef{
		Name:         "Firefox",
		RelativePath: filepath.Join("Mozilla", "Firefox", "Profiles"),
		AppDataBase:  AppDataRoaming,
		Type:         BrowserFirefox,
	}

	t.Run("chromium: single profile with History file found", func(t *testing.T) {
		base := t.TempDir()
		mkProfile(t, base, chromiumDef.RelativePath, "Default", "History")

		profiles, err := FindProfiles(base, chromiumDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 1 {
			t.Fatalf("expected 1 profile, got %d: %v", len(profiles), profiles)
		}
		if filepath.Base(profiles[0]) != "Default" {
			t.Errorf("expected profile named 'Default', got %q", filepath.Base(profiles[0]))
		}
	})

	t.Run("chromium: multiple profiles all found", func(t *testing.T) {
		base := t.TempDir()
		mkProfile(t, base, chromiumDef.RelativePath, "Default", "History")
		mkProfile(t, base, chromiumDef.RelativePath, "Profile 1", "History")
		mkProfile(t, base, chromiumDef.RelativePath, "Profile 2", "History")

		profiles, err := FindProfiles(base, chromiumDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 3 {
			t.Fatalf("expected 3 profiles, got %d: %v", len(profiles), profiles)
		}
	})

	t.Run("chromium: directory without History file not included", func(t *testing.T) {
		base := t.TempDir()
		mkProfile(t, base, chromiumDef.RelativePath, "Default", "History")
		// This directory exists but has no History file.
		mkProfile(t, base, chromiumDef.RelativePath, "Snapshots", "")

		profiles, err := FindProfiles(base, chromiumDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 1 {
			t.Fatalf("expected 1 profile, got %d: %v", len(profiles), profiles)
		}
		names := make([]string, len(profiles))
		for i, p := range profiles {
			names[i] = filepath.Base(p)
		}
		sort.Strings(names)
		if names[0] != "Default" {
			t.Errorf("unexpected profile: %v", names)
		}
	})

	t.Run("browser not installed returns empty slice without error", func(t *testing.T) {
		base := t.TempDir()
		// Do not create the browser directory at all.

		profiles, err := FindProfiles(base, chromiumDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 0 {
			t.Errorf("expected empty slice, got %v", profiles)
		}
	})

	t.Run("firefox: profile with places.sqlite found", func(t *testing.T) {
		base := t.TempDir()
		mkProfile(t, base, firefoxDef.RelativePath, "abc123.default", "places.sqlite")

		profiles, err := FindProfiles(base, firefoxDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 1 {
			t.Fatalf("expected 1 profile, got %d: %v", len(profiles), profiles)
		}
		if filepath.Base(profiles[0]) != "abc123.default" {
			t.Errorf("expected profile 'abc123.default', got %q", filepath.Base(profiles[0]))
		}
	})

	t.Run("firefox: directory without places.sqlite not included", func(t *testing.T) {
		base := t.TempDir()
		mkProfile(t, base, firefoxDef.RelativePath, "abc123.default", "places.sqlite")
		// This profile dir exists but has no places.sqlite.
		mkProfile(t, base, firefoxDef.RelativePath, "def456.recovery", "")

		profiles, err := FindProfiles(base, firefoxDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 1 {
			t.Fatalf("expected 1 profile, got %d: %v", len(profiles), profiles)
		}
	})
}
