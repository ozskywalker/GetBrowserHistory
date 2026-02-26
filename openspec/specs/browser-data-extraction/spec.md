# browser-data-extraction Specification

## Purpose
TBD - created by archiving change browser-history-report. Update Purpose after archive.
## Requirements
### Requirement: Locate browser profile directories
The binary SHALL discover all installed browser profile directories for a given Windows user by checking well-known AppData paths for Chrome, Edge, Firefox, and configurable Chromium-based browsers (e.g., DuckDuckGo, Brave).

#### Scenario: Chrome profiles found
- **WHEN** Chrome is installed and `C:\Users\<user>\AppData\Local\Google\Chrome\User Data\` exists
- **THEN** the binary enumerates all subdirectories containing a `History` file (e.g., `Default`, `Profile 1`, `Profile 2`) and includes each as a separate profile to query

#### Scenario: Firefox profiles found
- **WHEN** Firefox is installed and `C:\Users\<user>\AppData\Roaming\Mozilla\Firefox\Profiles\` exists
- **THEN** the binary enumerates all subdirectories containing a `places.sqlite` file and includes each as a separate profile to query

#### Scenario: Browser not installed
- **WHEN** a browser's expected AppData directory does not exist for a given user
- **THEN** the binary silently skips that browser for that user and continues

#### Scenario: Additional Chromium-based browser configured
- **WHEN** a `BrowserDef` entry is present in the `DefaultBrowsers` registry in `internal/browser/browsers.go`
- **THEN** the binary treats it as a Chromium-based browser and uses the same `History`/downloads schema

---

### Requirement: Copy locked database before querying
The binary SHALL copy each target SQLite database file to a temporary directory before opening it, to handle the case where the browser process holds a write lock on the file.

#### Scenario: Browser is open during extraction
- **WHEN** a browser is running and its History database is locked
- **THEN** the binary copies the file (plus any `-wal` and `-shm` sidecar files) to an `os.MkdirTemp()` directory, queries the copy, and deletes the temp directory after query completion

#### Scenario: Copy fails due to permissions
- **WHEN** the binary cannot copy the database file (e.g., ACL denies read)
- **THEN** the binary logs a warning for that profile and continues processing other profiles

#### Scenario: Temp directory cleanup
- **WHEN** database querying completes (success or error)
- **THEN** the temp directory created for that database copy SHALL be deleted via `defer os.RemoveAll(tempDir)`

---

### Requirement: Query Chromium browser history
The binary SHALL query the `urls` and `visits` tables from a Chromium-based browser's `History` SQLite database to extract visited URLs with timestamps, titles, and visit counts.

#### Scenario: Successful history extraction
- **WHEN** a valid Chromium `History` database is queried
- **THEN** the binary returns records containing: URL, page title, visit timestamp (converted from WebKit epoch to Go `time.Time` UTC), and visit count

#### Scenario: History table has zero rows
- **WHEN** the `urls` table exists but is empty
- **THEN** the binary returns an empty result set for that profile (no error)

#### Scenario: Unknown column in schema
- **WHEN** a column referenced in the query does not exist (schema version mismatch)
- **THEN** the binary logs a warning and returns an empty result set for that profile

---

### Requirement: Query Chromium browser downloads
The binary SHALL query the `downloads` and `downloads_url_chains` tables from a Chromium-based browser's `History` SQLite database to extract downloaded file records.

#### Scenario: Successful downloads extraction
- **WHEN** a valid Chromium `History` database with download records is queried
- **THEN** the binary returns records containing: target file path, source URL, start timestamp (converted from WebKit epoch to Go `time.Time` UTC), total bytes, and mime type

#### Scenario: No downloads present
- **WHEN** the `downloads` table is empty
- **THEN** the binary returns an empty result set and omits the downloads section from the report for that profile

---

### Requirement: Query Firefox history
The binary SHALL query the `moz_places` and `moz_historyvisits` tables from a Firefox `places.sqlite` database.

#### Scenario: Successful Firefox history extraction
- **WHEN** a valid Firefox `places.sqlite` database is queried
- **THEN** the binary returns records containing: URL, page title, visit timestamp (converted from PRTime microseconds to Go `time.Time` UTC), and visit count

#### Scenario: Firefox places.sqlite locked
- **WHEN** Firefox is running and `places.sqlite` is locked
- **THEN** the binary applies the same copy-then-query strategy as Chromium databases

---

### Requirement: Query Firefox downloads
The binary SHALL extract download records from Firefox by querying `moz_places` joined with `moz_annos` for the `downloads/destinationFileURI` annotation.

#### Scenario: Successful Firefox downloads extraction
- **WHEN** Firefox `places.sqlite` contains download annotation records
- **THEN** the binary returns records containing: destination file path, source URL, and timestamp

#### Scenario: No Firefox downloads
- **WHEN** no download annotations exist in `moz_annos`
- **THEN** the binary returns an empty result set for Firefox downloads

