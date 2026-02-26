## Context

Windows endpoints managed via RMM tools accumulate browser activity across multiple users and browser families. Security and IT teams need to forensically review this activity during incidents, compliance audits, or HR investigations. The binary must run as SYSTEM (the account most RMM agents use), without a desktop session, and must handle the common case where browsers are currently open and have their SQLite databases locked.

There is no existing code in this repository. This is a greenfield implementation.

## Goals / Non-Goals

**Goals:**
- Extract visited-URL history and downloaded-file records from Chrome, Edge, Firefox, and Chromium-based browsers (e.g., DuckDuckGo, Brave)
- Enumerate all local user profiles under `C:\Users\` or accept a targeted single user
- Produce a self-contained HTML report (inline CSS + JS, no external dependencies) and a JSON sidecar for programmatic consumption
- Execute cleanly as SYSTEM with no interactive prompts or GUI
- Handle locked databases by copying to a temp location before querying
- Work on Windows 10 / Windows 11 with no runtime, DLL, or installer required on the target

**Non-Goals:**
- Browser data from network drives or roaming profiles on remote shares
- Real-time monitoring or continuous collection (single-run snapshot only)
- Browser extension data, cookies, saved passwords, or form-fill data
- macOS or Linux support
- Automatic upload/exfiltration of results to any remote system
- Code signing infrastructure or GitHub Releases pipeline (deferred to future work)

## Decisions

### Decision 1: Go as implementation language

**Chosen**: Go 1.21+, compiled to a single Windows PE executable (`browser-report.exe`)

**Rationale**:
- Produces a single self-contained binary with no runtime dependencies — no PowerShell engine version concerns, no .NET requirements, no module installation
- `CGO_ENABLED=0` build with `modernc.org/sqlite` (pure Go SQLite) means the binary has zero external DLL dependencies
- Strong typing catches epoch conversion bugs and schema mismatches at compile time rather than at runtime on a target machine
- `html/template` provides automatic HTML escaping — correct by default, unlike string concatenation
- `defer` + `os.RemoveAll` is a cleaner and more reliable cleanup pattern than PS `try/finally`
- Cross-compiles from any platform: `GOOS=windows GOARCH=amd64 go build`

**Alternatives considered**:
- PowerShell 5.1: Native to Windows, but requires embedding `sqlite3.exe` as a base64 blob and executing it as a subprocess — which adds a binary-on-disk step anyway, creates a suspicious process tree, and is flagged by behavioral EDR rules. No meaningful deployment advantage over a Go binary delivered via the same RMM.
- Python: Not guaranteed present on managed endpoints.

---

### Decision 2: SQLite access via `modernc.org/sqlite` (pure Go)

**Chosen**: `modernc.org/sqlite` — a pure Go port of SQLite, no CGO required

**Rationale**:
- In-process API via Go's standard `database/sql` interface — no subprocess, no CSV parsing, no temp binary extraction
- Supports WAL mode natively
- `CGO_ENABLED=0` compatible: cross-compiles cleanly from macOS or Linux to a Windows binary
- Row streaming is possible for large result sets (avoiding full buffering)
- Typed column access: `int64` for timestamps, `string` for URLs — no string-to-epoch parsing risk

**Alternatives considered**:
- `mattn/go-sqlite3`: Most popular Go SQLite library, but requires CGO — breaks cross-compilation and links against platform libsqlite3
- Subprocess `sqlite3.exe`: Eliminated — see Decision 1
- `crawshaw/sqlite`: Pure Go but less actively maintained

---

### Decision 3: Package structure

**Chosen**: Structured Go module layout

```
browser-report/
  cmd/
    browser-report/
      main.go           ← CLI entry point, flag parsing, orchestration
  internal/
    browser/
      browsers.go       ← BrowserDef registry (paths, types, names)
    extract/
      users.go          ← User profile enumeration
      copy.go           ← Safe SQLite copy (locked file handling)
      sqlite.go         ← Generic SQLite query executor
      chromium.go       ← Chromium history + downloads extraction
      firefox.go        ← Firefox history + downloads extraction
    report/
      schema.go         ← Shared data structs + JSON tags
      html.go           ← HTML rendering via html/template
      json.go           ← JSON rendering via encoding/json
      template.html     ← Embedded HTML template (//go:embed)
  go.mod
  go.sum
```

**Rationale**: Separates CLI concerns from extraction logic from rendering. `internal/` prevents accidental external imports. Clean enough to extend (new browsers, new output formats) without touching the entry point.

---

### Decision 4: CLI flag parsing via stdlib `flag` package

**Chosen**: Go stdlib `flag` package with GNU long flag style (`--user`, `--output`, `--max-rows`, `--no-downloads`)

**Rationale**:
- Go's `flag` package natively accepts both `-flag` and `--flag` syntax — GNU double-dash style works with no additional dependencies
- No cobra/pflag dependency for a single-command tool
- Simple, auditable, no transitive deps
- Flag names are chosen to be intuitive for RMM operators: `--user alice`, `--output C:\Reports`, `--max-rows 5000`

---

### Decision 5: HTML templating via `html/template`

**Chosen**: Go stdlib `html/template` with template file embedded via `//go:embed`

**Rationale**:
- Automatic context-aware HTML escaping — URLs and page titles containing `<`, `>`, `&`, `"` are escaped correctly without explicit handling
- Template is a real `.html` file (not string concatenation) — readable and maintainable
- `//go:embed` compiles the template into the binary at build time, keeping the single-binary guarantee
- Inline CSS and JS remain in the template file; no external CDN references

---

### Decision 6: Database copy strategy for locked files

**Chosen**: Copy the SQLite file plus any `-wal` and `-shm` sidecar files to `os.MkdirTemp()` before opening; clean up with `defer os.RemoveAll(tempDir)`

**Rationale**: Chrome/Edge/Brave hold a write lock on their History database while running. The OS permits reads of locked files; copying the file bypasses the exclusive lock. WAL sidecar files must be copied too to ensure a consistent read.

**Risk**: Temp copy contains sensitive data momentarily. Mitigation: `defer os.RemoveAll` runs even on panic; GUID-named temp directory prevents collisions.

---

### Decision 7: Profile path discovery

**Chosen**: Well-known path patterns per browser family, resolved from `C:\Users\<username>\AppData\` directly (not from environment variables):

- **Chrome**: `AppData\Local\Google\Chrome\User Data\<profile>\`
- **Edge**: `AppData\Local\Microsoft\Edge\User Data\<profile>\`
- **Brave**: `AppData\Local\BraveSoftware\Brave-Browser\User Data\<profile>\`
- **DuckDuckGo**: `AppData\Local\DuckDuckGo\Browser\User Data\<profile>\`
- **Firefox**: `AppData\Roaming\Mozilla\Firefox\Profiles\<profile>\`

Chromium browsers share the same `History` SQLite schema. Firefox uses `places.sqlite` with a different schema. Additional Chromium-based browsers can be added to the `DefaultBrowsers` registry in `internal/browser/browsers.go`.

**Rationale for avoiding env vars**: When running as SYSTEM, `%LOCALAPPDATA%` and `%APPDATA%` point to the SYSTEM account's own profile, not the target user's. Paths must be constructed directly from `C:\Users\<username>\AppData\`.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Unsigned binary triggers SmartScreen on unmanaged machines | Tool is designed exclusively for RMM-managed deployment; SmartScreen is suppressed by policy on managed endpoints. Not intended for manual execution on unmanaged machines. Document this constraint. |
| SYSTEM account lacks read access to user AppData | SYSTEM has read access to all `C:\Users\*\AppData` on Windows 10/11 by default. Document as a prerequisite. |
| Browser DB schema changes between versions | Query only stable, long-lived columns; wrap column access in error handling and log warnings on mismatch |
| Very large history databases (> 1 GB) | `--max-rows` flag (default 10,000) applies `LIMIT` to all history queries |
| Temp files not cleaned up on panic | `defer os.RemoveAll(tempDir)` in each database accessor; top-level `defer` in `main` covers orchestration-level panics |
| Multiple Chromium profiles (`Default`, `Profile 1`, etc.) | Enumerate all subdirectories under `User Data\` containing a `History` file |
| DuckDuckGo browser path varies by version | Browser registry is defined in `internal/browser/browsers.go` — easy to patch without touching extraction logic |

## Migration Plan

Not applicable — greenfield implementation, no existing deployment.

**Delivery**: `browser-report.exe` is pushed to the target machine by the RMM tool, executed, and the output directory is retrieved via the RMM's file-collection mechanism. No persistent installation.

## Open Questions

1. **DuckDuckGo exact path**: DDG's Windows installer may place its profile under a different `AppData\Local\` subdirectory depending on the version. Verify against an installed copy before finalizing the path in `browsers.go`.
2. **Output path on non-standard Windows installs**: Assumes `C:\Windows\Temp\` exists and is writable by SYSTEM. If the Windows directory is on a different drive, the default output path may need adjustment. Consider falling back to `os.TempDir()` (which returns `%TEMP%` for the executing account — `C:\Windows\Temp` for SYSTEM).
