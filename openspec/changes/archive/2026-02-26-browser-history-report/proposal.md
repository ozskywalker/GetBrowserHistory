## Why

IT administrators and security teams frequently need to audit browser activity across managed Windows endpoints — examining visited sites, navigation timelines, and file downloads — for incident response, policy enforcement, and compliance purposes. There is no native, multi-browser solution that works headlessly as SYSTEM and covers all major Chromium and Firefox-based browsers in a single unified report.

## What Changes

- Introduce a Go binary that extracts browser history and download records from one or more user profiles on a Windows machine
- Support Microsoft Edge, Google Chrome, Mozilla Firefox, and other Chromium-based browsers (e.g., DuckDuckGo browser)
- Run entirely without a GUI, suitable for execution via RMM tools or other headless orchestration running as SYSTEM
- Output a structured HTML (and optionally JSON/CSV) report showing visited-URL timelines and download history per user per browser
- Handle locked SQLite database files (browsers may be open) via safe copy-then-read strategy
- Enumerate all user profiles under `C:\Users\` automatically, or target a specific user

## Capabilities

### New Capabilities

- `browser-data-extraction`: Locate and read SQLite history/download databases for Chrome, Edge, Firefox, and Chromium-based browsers across all Windows user profiles
- `report-generation`: Produce a self-contained HTML report (with embedded CSS/JS) showing visited URLs with timestamps and downloaded files with source URLs, grouped by user and browser
- `system-execution`: Support headless execution as the SYSTEM account via RMM, with proper path resolution and privilege-aware file access
- `multi-user-enumeration`: Automatically enumerate local Windows user profiles or accept a target-user argument to scope extraction

### Modified Capabilities

<!-- None — this is a greenfield project -->

## Impact

- **New binary**: One primary executable (`browser-report.exe`), structured as a Go module with `cmd/` and `internal/` packages
- **Dependencies**: No runtime required on target; self-contained single Windows PE binary (compiled with `CGO_ENABLED=0`)
- **File system access**: Reads `%LOCALAPPDATA%`, `%APPDATA%`, and profile paths under `C:\Users\`; requires read access to those paths (SYSTEM account has this)
- **No network calls**: Fully offline; all data is local
- **Output**: Report file written to a caller-specified path (default: working directory)
