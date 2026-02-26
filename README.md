# GetBrowserHistory

Builds a self-contained HTML (and JSON) report of pages visited, search terms used, and downloads recorded across all Chromium-based browsers and Firefox installed on a Windows machine.

Useful for forensic review, parental oversight, or auditing browser activity on a Windows system. Runs as a single `.exe` with no installation required.

[![Go Version](https://img.shields.io/badge/Go-1.24.0-blue.svg)](https://golang.org/doc/devel/release.html)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Designed with OpenSpec](https://img.shields.io/badge/Designed%20with-OpenSpec-007ACC?logo=openlogo&logoColor=white)](https://github.com/Fission-AI/OpenSpec)
![Claude Used](https://img.shields.io/badge/Built%20with-Claude-4B5AEA)

## Features

- **Search History table** — extracts search queries from Google, Bing, DuckDuckGo, Yahoo, YouTube, Ecosia, Brave Search, and StartPage URLs, surfaced in a dedicated table per browser profile
- **Full history table** — all visited URLs with timestamp, title, visit count, and extracted search query column
- **Downloads table** — downloaded files with source URL, destination path, and timestamp
- **Multi-user** — enumerates all Windows user profiles under `C:\Users\` by default
- **Multi-browser** — Chrome, Edge, Brave, DuckDuckGo Browser, Firefox
- **Safe to run while browsers are open** — copies locked SQLite databases before reading
- **Self-contained output** — single `.html` file with all CSS and JS inline; no internet connection needed to view
- **JSON output** — machine-readable `report.json` alongside the HTML for scripting or further analysis
- **Sortable & filterable tables** — click column headers to sort; type to filter rows

## Usage

```
browser-report.exe [flags]
```

Run with no flags to extract data for all users and all browsers:

```
browser-report.exe
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--user <name>` | all users | Restrict extraction to a single Windows username (e.g. `alice`) |
| `--output <path>` | `C:\Windows\Temp\BrowserReport_<timestamp>\` | Directory to write report files |
| `--max-rows <n>` | `10000` | Maximum history rows per browser profile |
| `--no-downloads` | off | Skip download history extraction |
| `--version` | | Print version and exit |

### Examples

```
# All users, default output location
browser-report.exe

# Single user
browser-report.exe --user alice

# Custom output directory, row limit
browser-report.exe --output C:\Reports --max-rows 5000

# History only, no downloads
browser-report.exe --no-downloads
```

## Output

Two files are written to the output directory:

| File | Description |
|---|---|
| `BrowserReport.html` | Interactive HTML report — open in any browser |
| `report.json` | Full data in JSON format |

The HTML report is organised by Windows user → browser → profile. Each profile section contains:

1. **Search History** — queries extracted from search engine visits (shown only when searches exist)
2. **History** — all visited URLs, including a Search Query column
3. **Downloads** — downloaded files (unless `-no-downloads` was passed)

## Supported Browsers

| Browser | Engine |
|---|---|
| Google Chrome | Chromium |
| Microsoft Edge | Chromium |
| Brave | Chromium |
| DuckDuckGo Browser | Chromium |
| Firefox | Gecko |

## Supported Search Engines

Search queries are extracted from URLs belonging to: Google (all country domains), Bing, DuckDuckGo, Yahoo, YouTube, Ecosia, Brave Search, StartPage.

## Building from Source

Requires [Go 1.24+](https://go.dev/dl/).

```
git clone https://github.com/GetBrowserHistory/GetBrowserHistory.git
cd GetBrowserHistory
go build -o browser-report.exe ./cmd/browser-report/
```

To embed a version string:

```
go build -ldflags "-X main.version=1.0.0" -o browser-report.exe ./cmd/browser-report/
```

## Requirements

- Windows (reads from `C:\Users\` and Windows AppData paths)
- Must be run with sufficient privileges to read other users' AppData directories (e.g. as Administrator) for multi-user extraction
- No external dependencies at runtime
