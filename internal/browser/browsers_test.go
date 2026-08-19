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

	t.Run("opera: roaming chromium profile with History found", func(t *testing.T) {
		// Opera stores its profile under AppData\Roaming (unlike Chrome/Edge/Brave)
		// with the History DB nested under a per-profile subdirectory, e.g.
		// ...\Opera Stable\Default\History.
		operaDef := BrowserDef{
			Name:         "Opera",
			RelativePath: filepath.Join("Opera Software", "Opera Stable"),
			AppDataBase:  AppDataRoaming,
			Type:         BrowserChromium,
		}

		base := t.TempDir() // simulates the Roaming AppData root
		mkProfile(t, base, operaDef.RelativePath, "Default", "History")

		profiles, err := FindProfiles(base, operaDef)
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

	t.Run("opera gx: roaming chromium profile with History found", func(t *testing.T) {
		operaGxDef := BrowserDef{
			Name:         "Opera GX",
			RelativePath: filepath.Join("Opera Software", "Opera GX Stable"),
			AppDataBase:  AppDataRoaming,
			Type:         BrowserChromium,
		}

		base := t.TempDir()
		mkProfile(t, base, operaGxDef.RelativePath, "Default", "History")

		profiles, err := FindProfiles(base, operaGxDef)
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

	t.Run("opera: not installed returns empty slice without error", func(t *testing.T) {
		operaDef := BrowserDef{
			Name:         "Opera",
			RelativePath: filepath.Join("Opera Software", "Opera Stable"),
			AppDataBase:  AppDataRoaming,
			Type:         BrowserChromium,
		}

		base := t.TempDir()
		profiles, err := FindProfiles(base, operaDef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 0 {
			t.Errorf("expected empty slice, got %v", profiles)
		}
	})
}

// TestDefaultBrowsersOperaRegistry asserts that the Opera and Opera GX entries
// in DefaultBrowsers carry the correct fields: Chromium engine type, profile
// under Roaming AppData (Opera's unique behaviour vs other Chromium browsers),
// and the expected RelativePath.
func TestDefaultBrowsersOperaRegistry(t *testing.T) {
	var opera, operaGX *BrowserDef
	for i := range DefaultBrowsers {
		switch DefaultBrowsers[i].Name {
		case "Opera":
			opera = &DefaultBrowsers[i]
		case "Opera GX":
			operaGX = &DefaultBrowsers[i]
		}
	}

	if opera == nil {
		t.Fatal("expected an 'Opera' entry in DefaultBrowsers")
	}
	if opera.AppDataBase != AppDataRoaming {
		t.Errorf("Opera AppDataBase = %q, want %q (Opera uses Roaming, not Local)", opera.AppDataBase, AppDataRoaming)
	}
	if opera.Type != BrowserChromium {
		t.Errorf("Opera Type = %q, want %q", opera.Type, BrowserChromium)
	}
	if want := "Opera Software\\Opera Stable"; opera.RelativePath != want {
		t.Errorf("Opera RelativePath = %q, want %q", opera.RelativePath, want)
	}

	if operaGX == nil {
		t.Fatal("expected an 'Opera GX' entry in DefaultBrowsers")
	}
	if operaGX.AppDataBase != AppDataRoaming {
		t.Errorf("Opera GX AppDataBase = %q, want %q (Opera GX uses Roaming, not Local)", operaGX.AppDataBase, AppDataRoaming)
	}
	if operaGX.Type != BrowserChromium {
		t.Errorf("Opera GX Type = %q, want %q", operaGX.Type, BrowserChromium)
	}
	if want := "Opera Software\\Opera GX Stable"; operaGX.RelativePath != want {
		t.Errorf("Opera GX RelativePath = %q, want %q", operaGX.RelativePath, want)
	}
}
