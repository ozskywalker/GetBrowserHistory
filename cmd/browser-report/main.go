package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ozskywalker/GetBrowserHistory/internal/browser"
	"github.com/ozskywalker/GetBrowserHistory/internal/extract"
	"github.com/ozskywalker/GetBrowserHistory/internal/report"
)

// version is injected at build time via -ldflags "-X main.version=<ver>"
var version = "dev"

// warnings accumulates non-fatal error messages for inclusion in the final report.
var warnings []string

func logProgress(format string, args ...any) {
	ts := time.Now().UTC().Format("2006-01-02 15:04:05")
	fmt.Fprintf(os.Stdout, "[%s UTC] %s\n", ts, fmt.Sprintf(format, args...))
}

func logWarn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
	warnings = append(warnings, msg)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func main() {
	var (
		flagUser        = flag.String("user", "", "Restrict extraction to a single Windows user profile name (e.g. alice). Default: all users.")
		flagOutput      = flag.String("output", "", "Directory to write report files. Default: C:\\Windows\\Temp\\BrowserReport_<timestamp>\\")
		flagMaxRows     = flag.Int("max-rows", 10000, "Maximum history rows to retrieve per browser profile.")
		flagNoDownloads = flag.Bool("no-downloads", false, "Skip download history extraction.")
		flagVersion     = flag.Bool("version", false, "Print version and exit.")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "browser-report %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n  browser-report.exe [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		// Capture PrintDefaults and rewrite single-dash prefixes to double-dash.
		// Go's flag package accepts both "-flag" and "--flag" at runtime.
		var buf strings.Builder
		flag.CommandLine.SetOutput(&buf)
		flag.PrintDefaults()
		flag.CommandLine.SetOutput(os.Stderr)
		out := strings.ReplaceAll("\n"+buf.String(), "\n  -", "\n  --")
		fmt.Fprint(os.Stderr, out[1:])
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  browser-report.exe\n")
		fmt.Fprintf(os.Stderr, "  browser-report.exe --user alice\n")
		fmt.Fprintf(os.Stderr, "  browser-report.exe --output C:\\Reports --max-rows 5000\n")
		fmt.Fprintf(os.Stderr, "  browser-report.exe --no-downloads\n")
	}

	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// Determine output directory.
	outputDir, err := createOutputDir(*flagOutput)
	if err != nil {
		fatal("cannot create output directory: %v", err)
	}
	logProgress("Output directory: %s", outputDir)

	// Determine users to process.
	const usersRoot = `C:\Users`
	var userNames []string
	if *flagUser != "" {
		resolved, err := extract.ResolveUser(usersRoot, *flagUser)
		if err != nil {
			fatal("%v", err)
		}
		userNames = []string{filepath.Base(resolved)}
	} else {
		userNames, err = extract.EnumerateUsers(usersRoot)
		if err != nil {
			fatal("cannot enumerate user profiles: %v", err)
		}
	}

	if len(userNames) == 0 {
		logWarn("no user profiles found under %s", usersRoot)
	}

	// Build report.
	rpt := report.Report{
		Meta: report.ReportMeta{
			GeneratedAt: time.Now().UTC(),
			Version:     version,
		},
	}

	// Populate hostname and executing account.
	if host, err := os.Hostname(); err == nil {
		rpt.Meta.Hostname = host
	}
	if user := currentUser(); user != "" {
		rpt.Meta.ExecutingAccount = user
	}

	totalHistory := 0
	totalDownloads := 0

	for _, userName := range userNames {
		logProgress("Processing user: %s", userName)

		localAppData := filepath.Join(usersRoot, userName, "AppData", "Local")
		roamingAppData := filepath.Join(usersRoot, userName, "AppData", "Roaming")

		userData := report.UserData{Username: userName}

		for _, bd := range browser.DefaultBrowsers {
			var appDataBase string
			if bd.AppDataBase == browser.AppDataLocal {
				appDataBase = localAppData
			} else {
				appDataBase = roamingAppData
			}

			profiles, err := browser.FindProfiles(appDataBase, bd)
			if err != nil {
				logWarn("user %s | browser %s: cannot find profiles: %v", userName, bd.Name, err)
				continue
			}

			for _, profilePath := range profiles {
				logProgress("  Browser: %s | Profile: %s", bd.Name, filepath.Base(profilePath))

				pd := report.ProfileData{
					BrowserName: bd.Name,
					ProfilePath: profilePath,
				}

				// Extract history.
				var history []extract.HistoryRecord
				var histErr error
				if bd.Type == browser.BrowserChromium {
					history, histErr = extract.ExtractChromiumHistory(profilePath, *flagMaxRows)
				} else {
					history, histErr = extract.ExtractFirefoxHistory(profilePath, *flagMaxRows)
				}
				if histErr != nil {
					logWarn("user %s | browser %s | profile %s: history error: %v", userName, bd.Name, filepath.Base(profilePath), histErr)
				} else {
					pd.History = history
					if len(history) == *flagMaxRows {
						pd.Truncated = true
						pd.TruncatedAt = *flagMaxRows
					}
					totalHistory += len(history)
				}

				// Extract downloads (unless --no-downloads).
				if !*flagNoDownloads {
					var downloads []extract.DownloadRecord
					var dlErr error
					if bd.Type == browser.BrowserChromium {
						downloads, dlErr = extract.ExtractChromiumDownloads(profilePath)
					} else {
						downloads, dlErr = extract.ExtractFirefoxDownloads(profilePath)
					}
					if dlErr != nil {
						logWarn("user %s | browser %s | profile %s: downloads error: %v", userName, bd.Name, filepath.Base(profilePath), dlErr)
					} else {
						pd.Downloads = downloads
						totalDownloads += len(downloads)
					}
				}

				userData.Profiles = append(userData.Profiles, pd)
			}
		}

		rpt.Users = append(rpt.Users, userData)
	}

	rpt.Warnings = warnings

	// Render and write HTML report.
	htmlBytes, err := report.RenderHTML(rpt)
	if err != nil {
		fatal("failed to render HTML report: %v", err)
	}
	htmlPath := filepath.Join(outputDir, "BrowserReport.html")
	if err := os.WriteFile(htmlPath, htmlBytes, 0644); err != nil {
		fatal("failed to write HTML report: %v", err)
	}

	// Render and write JSON report.
	jsonBytes, err := report.RenderJSON(rpt)
	if err != nil {
		fatal("failed to render JSON report: %v", err)
	}
	jsonPath := filepath.Join(outputDir, "report.json")
	if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
		fatal("failed to write JSON report: %v", err)
	}

	// Final summary.
	fmt.Fprintf(os.Stdout, "\n=== Report Complete ===\n")
	fmt.Fprintf(os.Stdout, "Output:     %s\n", outputDir)
	fmt.Fprintf(os.Stdout, "Users:      %d\n", len(userNames))
	fmt.Fprintf(os.Stdout, "History:    %d rows\n", totalHistory)
	fmt.Fprintf(os.Stdout, "Downloads:  %d rows\n", totalDownloads)
	fmt.Fprintf(os.Stdout, "Warnings:   %d\n", len(warnings))
	fmt.Fprintf(os.Stdout, "HTML:       %s\n", htmlPath)
	fmt.Fprintf(os.Stdout, "JSON:       %s\n", jsonPath)

	os.Exit(0)
}

// createOutputDir creates and returns the output directory path.
// If basePath is empty, defaults to C:\Windows\Temp\BrowserReport_<timestamp>\.
func createOutputDir(basePath string) (string, error) {
	if basePath == "" {
		ts := time.Now().UTC().Format("20060102_150405")
		basePath = filepath.Join(`C:\Windows\Temp`, "BrowserReport_"+ts)
	}
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", err
	}
	// Verify write access by creating and removing a test file.
	testFile := filepath.Join(basePath, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return "", fmt.Errorf("output directory not writable: %w", err)
	}
	f.Close()
	os.Remove(testFile)
	return basePath, nil
}

// currentUser returns the currently executing Windows username via environment.
// When running as SYSTEM, returns "NT AUTHORITY\\SYSTEM".
func currentUser() string {
	if u := os.Getenv("USERNAME"); u != "" {
		domain := os.Getenv("USERDOMAIN")
		if domain != "" && !strings.EqualFold(domain, os.Getenv("COMPUTERNAME")) {
			return domain + "\\" + u
		}
		return u
	}
	return ""
}
