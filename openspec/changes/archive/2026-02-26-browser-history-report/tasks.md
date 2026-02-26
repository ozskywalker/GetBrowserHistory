## 1. Project Setup

- [x] 1.1 Initialize Go module: `go mod init github.com/<org>/browser-report` targeting Go 1.21 minimum
- [x] 1.2 Create directory structure: `cmd/browser-report/`, `internal/browser/`, `internal/extract/`, `internal/report/`
- [x] 1.3 Add `modernc.org/sqlite` dependency via `go get modernc.org/sqlite`
- [x] 1.4 Run `go mod tidy` to generate `go.sum` and verify no CGO dependency is introduced

## 2. CLI Entry Point & Flags (`cmd/browser-report/main.go`)

- [x] 2.1 Define GNU long flags using stdlib `flag` package: `--user` (string, default empty), `--output` (string, default empty), `--max-rows` (int, default 10000), `--no-downloads` (bool), `--version` (bool)
- [x] 2.2 Implement `--version` flag: print version string (injected via `-ldflags "-X main.version=<ver>"`) and exit 0
- [x] 2.3 Implement `--help` output: `flag.Usage` function listing all flags with descriptions
- [x] 2.4 Implement timestamped progress logger: `logProgress(format string, args ...any)` writes `[YYYY-MM-DD HH:MM:SS UTC] <message>` to stdout
- [x] 2.5 Implement warning logger: `logWarn(format string, args ...any)` writes `[WARN] <message>` to stderr and appends to a `[]string` warnings accumulator
- [x] 2.6 Implement `os.Exit(0)` on success / `os.Exit(1)` on fatal error with error message written to stderr

## 3. SQLite Layer (`internal/extract/`)

- [x] 3.1 Create `sqlite.go`: implement `QueryDB(dbPath, sql string, args ...any) ([]map[string]string, error)` using `database/sql` with `_ "modernc.org/sqlite"` driver registration
- [x] 3.2 Handle schema-mismatch errors in `QueryDB`: catch column-not-found errors, return `(nil, err)` so callers can log and skip
- [x] 3.3 Create `copy.go`: implement `CopyDatabase(src string) (tempDir string, destPath string, err error)` that calls `os.MkdirTemp`, copies the `.sqlite` file, and also copies any `-wal` and `-shm` sidecar files if they exist
- [x] 3.4 Document the expected caller pattern: `tempDir, dbCopy, err := CopyDatabase(src); defer os.RemoveAll(tempDir)` — ensure all callers follow this

## 4. User Profile Enumeration (`internal/extract/users.go`)

- [x] 4.1 Implement `EnumerateUsers(usersRoot string) ([]string, error)` that reads directory entries under `C:\Users\` and returns user directory names
- [x] 4.2 Implement skip list for system directories: `Public`, `Default`, `Default User`, `All Users`
- [x] 4.3 Implement `ResolveUser(usersRoot, userName string) (string, error)` for `--user` override — validates existence of `C:\Users\<userName>`, returns error if not found
- [x] 4.4 Construct `LocalAppData` and `RoamingAppData` paths as `filepath.Join(usersRoot, username, "AppData", "Local")` and `..., "Roaming"` — no env var lookups

## 5. Browser Registry (`internal/browser/browsers.go`)

- [x] 5.1 Define `BrowserType` as a string enum: `Chromium`, `Firefox`
- [x] 5.2 Define `AppDataType` as a string enum: `Local`, `Roaming`
- [x] 5.3 Define `BrowserDef` struct: `Name string`, `RelativePath string` (path under Local or Roaming AppData to the `User Data` or `Profiles` root), `AppDataBase AppDataType`, `Type BrowserType`
- [x] 5.4 Define `DefaultBrowsers []BrowserDef` with entries for: Chrome (`Local\Google\Chrome\User Data`), Edge (`Local\Microsoft\Edge\User Data`), Brave (`Local\BraveSoftware\Brave-Browser\User Data`), DuckDuckGo (`Local\DuckDuckGo\Browser\User Data`), Firefox (`Roaming\Mozilla\Firefox\Profiles`)
- [x] 5.5 Implement `FindProfiles(appDataBase string, b BrowserDef) ([]string, error)` — for Chromium browsers, enumerate subdirectories of `User Data` containing a `History` file; for Firefox, enumerate subdirectories of `Profiles` containing a `places.sqlite` file

## 6. Chromium Extraction (`internal/extract/chromium.go`)

- [x] 6.1 Define `HistoryRecord` struct: `URL string`, `Title string`, `VisitCount int`, `LastVisitUTC time.Time`
- [x] 6.2 Define `DownloadRecord` struct: `TargetPath string`, `SourceURL string`, `MimeType string`, `TotalBytes int64`, `StartTimeUTC time.Time`
- [x] 6.3 Implement `WebKitEpochToTime(microseconds int64) time.Time` — WebKit epoch is microseconds since 1601-01-01 00:00:00 UTC; subtract the delta from Unix epoch and use `time.Unix`
- [x] 6.4 Implement `ExtractChromiumHistory(profilePath string, maxRows int) ([]HistoryRecord, error)` — calls `CopyDatabase` on the `History` file, queries `SELECT u.url, u.title, u.visit_count, v.visit_time FROM urls u JOIN visits v ON u.id = v.url ORDER BY v.visit_time DESC LIMIT ?`
- [x] 6.5 Implement `ExtractChromiumDownloads(profilePath string) ([]DownloadRecord, error)` — copies `History` DB, queries `downloads LEFT JOIN downloads_url_chains` for target path, URL chain, start time, total bytes, mime type

## 7. Firefox Extraction (`internal/extract/firefox.go`)

- [x] 7.1 Implement `PRTimeToTime(microseconds int64) time.Time` — PRTime is microseconds since Unix epoch; divide by 1,000,000 and use `time.Unix`
- [x] 7.2 Implement `ExtractFirefoxHistory(profilePath string, maxRows int) ([]HistoryRecord, error)` — calls `CopyDatabase` on `places.sqlite`, queries `SELECT p.url, p.title, p.visit_count, h.visit_date FROM moz_places p JOIN moz_historyvisits h ON p.id = h.place_id ORDER BY h.visit_date DESC LIMIT ?`
- [x] 7.3 Implement `ExtractFirefoxDownloads(profilePath string) ([]DownloadRecord, error)` — copies `places.sqlite`, queries `moz_places JOIN moz_annos` filtering for the `downloads/destinationFileURI` annotation attribute to extract destination path, source URL, and timestamp

## 8. Report Data Model (`internal/report/schema.go`)

- [x] 8.1 Define `ProfileData` struct: `BrowserName string`, `ProfilePath string`, `History []extract.HistoryRecord`, `Downloads []extract.DownloadRecord`, `Truncated bool`, `TruncatedAt int`
- [x] 8.2 Define `UserData` struct: `Username string`, `Profiles []ProfileData`
- [x] 8.3 Define `ReportMeta` struct with `json` tags: `GeneratedAt time.Time`, `Hostname string`, `ExecutingAccount string`, `Version string`
- [x] 8.4 Define `Report` struct: `Meta ReportMeta`, `Users []UserData`, `Warnings []string`

## 9. HTML Report (`internal/report/html.go` + `template.html`)

- [x] 9.1 Create `internal/report/template.html` using Go `html/template` syntax (`{{range .Users}}`, `{{.Username}}`, etc.)
- [x] 9.2 Implement `//go:embed template.html` in `html.go` to compile template into binary at build time
- [x] 9.3 Add inline CSS to template: table styling, user tab bar (one tab per user), collapsible browser/profile sections, responsive layout — no external CDN references
- [x] 9.4 Add inline JS to template: column sort on `<th>` click (toggle asc/desc using data attribute), per-table text filter input that hides non-matching rows on `input` event
- [x] 9.5 Implement `RenderHTML(r Report) ([]byte, error)` using `html/template.Execute()` — auto-escaping handles all URL and title content
- [x] 9.6 Add truncation notice to template: when `ProfileData.Truncated` is true, render a banner above the history table noting row count was capped at `TruncatedAt`
- [x] 9.7 Add warnings section to template: when `Report.Warnings` is non-empty, render a collapsible warnings panel below the main content

## 10. JSON Report (`internal/report/json.go`)

- [x] 10.1 Implement `RenderJSON(r Report) ([]byte, error)` using `encoding/json.MarshalIndent` with 2-space indent
- [x] 10.2 Verify `time.Time` fields serialize as RFC 3339 / ISO 8601 UTC strings (Go's default JSON marshaling for `time.Time`)
- [x] 10.3 Verify output schema matches spec: top-level `meta` object + `users` array, each user with `profiles` array, each profile with `browserName`, `profilePath`, `history` array, `downloads` array

## 11. Orchestration (`cmd/browser-report/main.go`)

- [x] 11.1 Implement `createOutputDir(basePath string) (string, error)` — if `basePath` is empty, default to `C:\Windows\Temp\BrowserReport_<20060102_150405>\`; create directory and verify write access
- [x] 11.2 Wire main loop: enumerate users → for each user, call `browser.FindProfiles` per `DefaultBrowsers` entry → call appropriate `Extract*History` and `Extract*Downloads` → accumulate into `Report` struct
- [x] 11.3 Call `report.RenderHTML` and `report.RenderJSON`, write results to `<outputDir>\BrowserReport.html` and `<outputDir>\report.json`
- [x] 11.4 Print final summary to stdout: output directory path, users processed, total history rows, total download rows, warning count

## 12. Build Verification

- [x] 12.1 Confirm `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o browser-report.exe ./cmd/browser-report/` succeeds with no CGO requirement
- [x] 12.2 Confirm resulting `.exe` has no unexpected DLL dependencies (`dumpbin /dependents browser-report.exe` should show only standard Windows system DLLs)
- [x] 12.3 Verify version injection: `go build -ldflags "-X main.version=0.1.0" ...` and confirm `browser-report.exe --version` outputs `0.1.0`

## 13. Testing & Validation

- [ ] 13.1 Test on machine with Chrome installed and browser open (locked DB) — confirm report produced without error and history data present
- [ ] 13.2 Test on machine with Firefox installed — confirm Firefox history and downloads appear in report
- [ ] 13.3 Test as SYSTEM via `PsExec64.exe -s browser-report.exe` — confirm all user profiles discovered, output written to `C:\Windows\Temp\`
- [ ] 13.4 Test `--user alice` (valid user) — confirm only that user appears in report
- [ ] 13.5 Test `--user nonexistent` — confirm exit code 1 and error message to stderr
- [ ] 13.6 Test `--max-rows 50` — confirm truncation notice appears in HTML report for profiles with more than 50 history rows
- [ ] 13.7 Test `--no-downloads` — confirm downloads tables are absent from HTML report and `downloads` arrays are empty in JSON
- [ ] 13.8 Validate HTML report opens in Edge, Chrome, and Firefox — no console errors, sort and filter interactions work correctly
- [ ] 13.9 Validate `report.json` parses cleanly (`python -m json.tool report.json` or equivalent) — schema matches spec, timestamps are ISO 8601
- [ ] 13.10 Test on machine with no browsers installed — confirm exit code 0 and HTML report contains "No browser data found"
