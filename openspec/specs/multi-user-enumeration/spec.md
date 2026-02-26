# multi-user-enumeration Specification

## Purpose
TBD - created by archiving change browser-history-report. Update Purpose after archive.
## Requirements
### Requirement: Enumerate all local user profiles
The binary SHALL enumerate all user profile directories under `C:\Users\` and process each one for browser data, skipping system accounts.

#### Scenario: Multiple user profiles present
- **WHEN** `C:\Users\` contains directories for multiple users (e.g., `alice`, `bob`, `Administrator`)
- **THEN** the binary processes each directory as a separate user and includes each in the report under its own section

#### Scenario: System account directories skipped
- **WHEN** `C:\Users\` contains well-known system directories (`Public`, `Default`, `Default User`, `All Users`)
- **THEN** the binary skips these directories and does not attempt to extract browser data from them

#### Scenario: No browser data for a user
- **WHEN** a user profile directory exists but contains no recognized browser data
- **THEN** the binary includes that user in the report with a "No browser data found" message and continues

---

### Requirement: Target a specific user via flag
The binary SHALL accept an optional `--user` flag to restrict extraction to a single user profile, overriding the default all-users enumeration.

#### Scenario: Single user targeted
- **WHEN** `--user alice` is provided
- **THEN** the binary processes only `C:\Users\alice\` and the report contains data only for that user

#### Scenario: Targeted user does not exist
- **WHEN** `--user nonexistent` is provided and `C:\Users\nonexistent\` does not exist
- **THEN** the binary writes an error to stderr and exits with code `1`

#### Scenario: Targeted user has no browser data
- **WHEN** `--user alice` is provided but Alice has no browser history
- **THEN** the binary produces a report with Alice's section showing "No browser data found" and exits with code `0`

---

### Requirement: Respect row limit per profile
The binary SHALL accept a `--max-rows` flag (default: `10000`) that limits the number of history rows returned per browser profile to prevent excessive memory usage and report size.

#### Scenario: History exceeds row limit
- **WHEN** a browser profile's history table contains more rows than `--max-rows`
- **THEN** the binary returns only the most recent `--max-rows` rows (ordered by visit timestamp descending) and notes the truncation in the report

#### Scenario: History within row limit
- **WHEN** a browser profile's history table contains fewer rows than `--max-rows`
- **THEN** all rows are returned without truncation

---

### Requirement: Downloads opt-out flag
The binary SHALL include downloads extraction by default. A `--no-downloads` flag SHALL disable downloads extraction entirely.

#### Scenario: Downloads extracted by default
- **WHEN** `--no-downloads` is not provided
- **THEN** the binary extracts and includes download records for all browser profiles

#### Scenario: Downloads skipped with --no-downloads
- **WHEN** `--no-downloads` is provided
- **THEN** the binary skips all download queries and the report omits all downloads sections

---

### Requirement: Binary usage documented via --help
The binary SHALL output usage documentation when invoked with `--help` or `-h`.

#### Scenario: --help displays flag documentation
- **WHEN** a user runs `browser-report.exe --help`
- **THEN** the output lists all flags with descriptions: `--user`, `--output`, `--max-rows`, `--no-downloads`, `--version`

