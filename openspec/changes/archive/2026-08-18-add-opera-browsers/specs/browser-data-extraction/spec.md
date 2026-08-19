## MODIFIED Requirements

### Requirement: Locate browser profile directories
The binary SHALL discover all installed browser profile directories for a given Windows user by checking well-known AppData paths for Chrome, Edge, Opera, Opera GX, Firefox, and configurable Chromium-based browsers (e.g., DuckDuckGo, Brave).

#### Scenario: Chrome profiles found
- **WHEN** Chrome is installed and `C:\Users\<user>\AppData\Local\Google\Chrome\User Data\` exists
- **THEN** the binary enumerates all subdirectories containing a `History` file (e.g., `Default`, `Profile 1`, `Profile 2`) and includes each as a separate profile to query

#### Scenario: Opera profiles found
- **WHEN** Opera is installed and `C:\Users\<user>\AppData\Roaming\Opera Software\Opera Stable\` exists
- **THEN** the binary enumerates all subdirectories containing a `History` file (e.g., `Default`) and includes each as a separate profile to query

#### Scenario: Opera GX profiles found
- **WHEN** Opera GX is installed and `C:\Users\<user>\AppData\Roaming\Opera Software\Opera GX Stable\` exists
- **THEN** the binary enumerates all subdirectories containing a `History` file (e.g., `Default`) and includes each as a separate profile to query

#### Scenario: Firefox profiles found
- **WHEN** Firefox is installed and `C:\Users\<user>\AppData\Roaming\Mozilla\Firefox\Profiles\` exists
- **THEN** the binary enumerates all subdirectories containing a `places.sqlite` file and includes each as a separate profile to query

#### Scenario: Browser not installed
- **WHEN** a browser's expected AppData directory does not exist for a given user
- **THEN** the binary silently skips that browser for that user and continues

#### Scenario: Additional Chromium-based browser configured
- **WHEN** a `BrowserDef` entry is present in the `DefaultBrowsers` registry in `internal/browser/browsers.go`
- **THEN** the binary treats it as a Chromium-based browser and uses the same `History`/downloads schema
