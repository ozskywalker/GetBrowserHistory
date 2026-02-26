# report-generation Specification

## Purpose
TBD - created by archiving change browser-history-report. Update Purpose after archive.
## Requirements
### Requirement: Generate self-contained HTML report
The binary SHALL produce a single `.html` file with all CSS and JavaScript embedded inline (no external dependencies) that displays browser history and download data organized by user and browser.

#### Scenario: Report created at output path
- **WHEN** data extraction completes successfully
- **THEN** the binary writes a single `.html` file to the specified output directory (default: `C:\Windows\Temp\BrowserReport_<timestamp>\`) and displays the full file path to stdout

#### Scenario: Report with multiple users and browsers
- **WHEN** data from multiple users and multiple browsers is collected
- **THEN** the HTML report presents a tabbed interface with one top-level tab per user, and within each user tab, one collapsible section per browser, each containing a history table and a downloads table

#### Scenario: Empty report
- **WHEN** no browser data is found for any user
- **THEN** the binary still produces a valid HTML report with a "No data found" message and exits with code `0`

---

### Requirement: History table display
The HTML report SHALL render visited URL records in a sortable, searchable table showing timestamp, URL (as a clickable link), page title, visit count, and search query.

#### Scenario: Timestamp display
- **WHEN** a history record is rendered
- **THEN** the visit timestamp SHALL be displayed in ISO 8601 format (YYYY-MM-DD HH:MM:SS UTC)

#### Scenario: Table sorting
- **WHEN** a user clicks a column header in the history table
- **THEN** the table rows sort by that column (ascending on first click, descending on second click)

#### Scenario: Table filtering
- **WHEN** a user types in the search/filter box above a history table
- **THEN** rows whose URL or title do not match the filter text are hidden

#### Scenario: Search query column populated for search URLs
- **WHEN** a history record has a non-empty SearchQuery field
- **THEN** the Search Query column displays the extracted query text

#### Scenario: Search query column empty for non-search URLs
- **WHEN** a history record has an empty SearchQuery field
- **THEN** the Search Query column cell is empty

### Requirement: Downloads table display
The HTML report SHALL render downloaded file records in a sortable table showing timestamp, destination file path, source URL, and file size.

#### Scenario: Downloads section rendered
- **WHEN** download records exist for a browser profile
- **THEN** a "Downloads" table is rendered beneath the history table for that profile

#### Scenario: No downloads
- **WHEN** no download records exist for a browser profile
- **THEN** the downloads section for that profile displays "No downloads recorded"

---

### Requirement: JSON output
The binary SHALL also write a `report.json` file alongside the HTML report containing all extracted data in machine-readable form.

#### Scenario: JSON file created
- **WHEN** the report generation completes
- **THEN** a `report.json` file is written to the same output directory as the HTML file, containing an array of user objects, each with a `browsers` array, each with `history` and `downloads` arrays

#### Scenario: JSON is valid
- **WHEN** the JSON file is opened by any standard JSON parser
- **THEN** it SHALL parse without errors and conform to the documented schema

---

### Requirement: Report includes generation metadata
The HTML and JSON reports SHALL include a metadata section showing binary version, execution timestamp, executing account, and hostname.

#### Scenario: Metadata in HTML report header
- **WHEN** the HTML report is opened
- **THEN** the page header SHALL display: report generation timestamp (UTC), hostname, executing Windows account, and binary version

#### Scenario: Metadata in JSON
- **WHEN** the JSON report is parsed
- **THEN** a top-level `meta` object SHALL contain: `generatedAt`, `hostname`, `executingAccount`, `scriptVersion`

### Requirement: Search History summary table
The HTML report SHALL render a Search History table above the History table for each browser profile, listing only records where a search query was extracted. The table SHALL be omitted when no search records exist for that profile.

#### Scenario: Search History table shown when searches exist
- **WHEN** at least one history record for a profile has a non-empty SearchQuery
- **THEN** a Search History table is rendered above the History table for that profile, with columns: Timestamp (UTC), Engine, Query, URL

#### Scenario: Search History table omitted when no searches
- **WHEN** no history records for a profile have a non-empty SearchQuery
- **THEN** no Search History table is rendered for that profile

#### Scenario: Search History table is sortable
- **WHEN** a user clicks a column header in the Search History table
- **THEN** the table rows sort by that column (ascending on first click, descending on second click)

#### Scenario: Engine name derived from host
- **WHEN** a search record is rendered in the Search History table
- **THEN** the Engine column displays the recognisable engine name (e.g. "Google", "YouTube", "DuckDuckGo") derived from the URL host

#### Scenario: Search History query is a clickable link
- **WHEN** a search record is rendered
- **THEN** the URL column contains a clickable link to the original search URL

### Requirement: JSON includes search query field
The JSON report SHALL include a `searchQuery` field on each history record object, containing the extracted query string or empty string.

#### Scenario: JSON search query populated
- **WHEN** the JSON report is parsed and a history record had a search query extracted
- **THEN** that record's `searchQuery` field contains the human-readable query string

#### Scenario: JSON search query empty for non-search records
- **WHEN** the JSON report is parsed and a history record had no search query
- **THEN** that record's `searchQuery` field is present and set to `""`

