# system-execution Specification

## Purpose
TBD - created by archiving change browser-history-report. Update Purpose after archive.
## Requirements
### Requirement: Headless execution with no interactive prompts
The binary SHALL execute completely non-interactively, producing no GUI dialogs, stdin reads, or blocking calls, so it can be dispatched by an RMM agent or scheduled task running as SYSTEM.

#### Scenario: Binary runs without user input
- **WHEN** `browser-report.exe` is launched with all required flags supplied on the command line
- **THEN** it runs to completion and exits without waiting for stdin

#### Scenario: Missing optional flags use defaults
- **WHEN** optional flags (`--output`, `--max-rows`, `--no-downloads`) are omitted
- **THEN** the binary uses documented defaults and continues without prompting

---

### Requirement: Exit codes
The binary SHALL return a meaningful exit code to the calling process.

#### Scenario: Successful execution
- **WHEN** the binary completes and produces a report (even if some browsers had no data)
- **THEN** the binary exits with code `0`

#### Scenario: Fatal error
- **WHEN** the binary encounters an unrecoverable error (e.g., output directory cannot be created)
- **THEN** the binary exits with code `1` and writes the error message to stderr

#### Scenario: Partial failure
- **WHEN** some browser profiles fail to extract but others succeed
- **THEN** the binary exits with code `0`, logs warnings for failed profiles to stderr, and includes a warnings section in the report

---

### Requirement: Structured stdout/stderr logging
The binary SHALL write progress messages to stdout and error/warning messages to stderr so RMM tools can capture them separately.

#### Scenario: Progress output
- **WHEN** the binary processes each user and browser profile
- **THEN** a timestamped progress line is written to stdout (e.g., `[2025-01-15 10:23:45 UTC] Processing user: jsmith | Browser: Chrome | Profile: Default`)

#### Scenario: Error output
- **WHEN** a non-fatal error occurs (e.g., a single profile fails)
- **THEN** a warning line prefixed with `[WARN]` is written to stderr and execution continues

---

### Requirement: No runtime dependencies on target machine
The binary SHALL function on a stock Windows 10 / Windows 11 system with no pre-installed runtimes, DLLs, modules, or packages.

#### Scenario: Self-contained binary
- **WHEN** `browser-report.exe` is placed on a machine with no Go runtime, no Visual C++ redistributable, and no SQLite DLL installed
- **THEN** it executes successfully — SQLite support is compiled in via `modernc.org/sqlite` (pure Go, `CGO_ENABLED=0`)

#### Scenario: No files extracted to disk at runtime
- **WHEN** the binary runs
- **THEN** it does NOT extract any embedded executables, DLLs, or helper binaries to disk — all functionality runs in-process

---

### Requirement: SYSTEM account privilege compatibility
The binary SHALL run correctly when executed as the Windows SYSTEM account, which has no interactive user profile and different environment variable values than interactive users.

#### Scenario: SYSTEM account resolves user profile paths correctly
- **WHEN** running as SYSTEM
- **THEN** the binary enumerates `C:\Users\*` directory entries directly (not via `%USERPROFILE%`, `%LOCALAPPDATA%`, or `%APPDATA%` env vars) to discover user profiles

#### Scenario: Output written to SYSTEM-writable location
- **WHEN** no `--output` flag is specified
- **THEN** the default output path is `C:\Windows\Temp\BrowserReport_<timestamp>\`, which SYSTEM has write access to

---

### Requirement: Version flag
The binary SHALL support a `--version` flag that prints the build version and exits.

#### Scenario: Version output
- **WHEN** `browser-report.exe --version` is run
- **THEN** the binary prints the version string (injected at build time via `-ldflags "-X main.version=<ver>"`) and exits with code `0`

---

### Requirement: RMM-managed deployment
The binary SHALL be delivered and executed via an RMM tool or other managed deployment mechanism. It is not intended for manual download and execution on unmanaged endpoints.

#### Scenario: RMM delivery establishes trust
- **WHEN** the binary is pushed and executed via an RMM agent running as SYSTEM
- **THEN** execution proceeds normally — managed endpoints suppress SmartScreen for RMM-delivered files via group policy

#### Scenario: Manual execution on unmanaged machine
- **WHEN** a user downloads and double-clicks `browser-report.exe` on an unmanaged Windows machine without a code-signing certificate
- **THEN** Windows SmartScreen may display a warning — this is expected behavior and out of scope for this tool

