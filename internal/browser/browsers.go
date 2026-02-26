package browser

import (
	"os"
	"path/filepath"
)

// BrowserType identifies the browser engine family.
type BrowserType string

const (
	BrowserChromium BrowserType = "Chromium"
	BrowserFirefox  BrowserType = "Firefox"
)

// AppDataType identifies which AppData subdirectory to use.
type AppDataType string

const (
	AppDataLocal   AppDataType = "Local"
	AppDataRoaming AppDataType = "Roaming"
)

// BrowserDef describes how to locate a browser's profile data for a given user.
type BrowserDef struct {
	// Name is the human-readable browser name shown in the report.
	Name string
	// RelativePath is the path relative to the AppData base (Local or Roaming)
	// that contains the browser's profile root directory (e.g. "User Data" for
	// Chromium, or "Mozilla\Firefox\Profiles" for Firefox).
	RelativePath string
	// AppDataBase indicates whether RelativePath is under AppData\Local or
	// AppData\Roaming.
	AppDataBase AppDataType
	// Type is the browser engine family, which determines the SQLite schema.
	Type BrowserType
}

// DefaultBrowsers is the registry of supported browsers. To add a new
// Chromium-based browser, append a BrowserDef with Type: BrowserChromium and
// the appropriate RelativePath under AppData\Local.
var DefaultBrowsers = []BrowserDef{
	{
		Name:         "Chrome",
		RelativePath: `Google\Chrome\User Data`,
		AppDataBase:  AppDataLocal,
		Type:         BrowserChromium,
	},
	{
		Name:         "Edge",
		RelativePath: `Microsoft\Edge\User Data`,
		AppDataBase:  AppDataLocal,
		Type:         BrowserChromium,
	},
	{
		Name:         "Brave",
		RelativePath: `BraveSoftware\Brave-Browser\User Data`,
		AppDataBase:  AppDataLocal,
		Type:         BrowserChromium,
	},
	{
		Name:         "DuckDuckGo",
		RelativePath: `DuckDuckGo\Browser\User Data`,
		AppDataBase:  AppDataLocal,
		Type:         BrowserChromium,
	},
	{
		Name:         "Firefox",
		RelativePath: `Mozilla\Firefox\Profiles`,
		AppDataBase:  AppDataRoaming,
		Type:         BrowserFirefox,
	},
}

// FindProfiles locates profile directories for the given browser under
// appDataBase (the user's full Local or Roaming AppData path).
//
// For Chromium browsers, it looks for subdirectories of RelativePath that
// contain a "History" file.
//
// For Firefox, it looks for subdirectories of RelativePath that contain a
// "places.sqlite" file.
//
// Returns an empty slice (not an error) if the browser is not installed.
func FindProfiles(appDataBase string, b BrowserDef) ([]string, error) {
	root := filepath.Join(appDataBase, b.RelativePath)

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // browser not installed for this user
		}
		return nil, err
	}

	var markerFile string
	switch b.Type {
	case BrowserChromium:
		markerFile = "History"
	case BrowserFirefox:
		markerFile = "places.sqlite"
	}

	var profiles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate := filepath.Join(root, e.Name(), markerFile)
		if _, err := os.Stat(candidate); err == nil {
			profiles = append(profiles, filepath.Join(root, e.Name()))
		}
	}
	return profiles, nil
}
