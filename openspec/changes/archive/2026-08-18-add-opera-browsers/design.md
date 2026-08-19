## Context

The report utility discovers installed browsers via the `DefaultBrowsers` registry in `internal/browser/browsers.go`. Each `BrowserDef` describes a `Name`, a `RelativePath` under AppData, an `AppDataBase` (`Local` or `Roaming`), and a `Type` (`Chromium` or `Firefox`). Discovery (`FindProfiles`) enumerates subdirectories of `RelativePath` that contain an engine-specific marker file (`History` for Chromium, `places.sqlite` for Firefox). Extraction is dispatched in `cmd/browser-report/main.go` by engine `Type`, selecting the base directory from `AppDataBase`.

All existing Chromium entries (Chrome, Edge, Brave, DuckDuckGo) use `AppDataBase: AppDataLocal`. Opera is different: its profile data lives under `AppData\Roaming`, with the `History` DB at `%APPDATA%\Opera Software\Opera Stable\Default\History` (confirmed against a real install). Opera GX follows the identical pattern under `Opera GX Stable`.

## Goals / Non-Goals

**Goals:**
- Discover Opera (`Opera Software\Opera Stable`) and Opera GX (`Opera Software\Opera GX Stable`) Chromium profiles.
- Reuse the existing Chromium history/downloads extraction without modification.
- Preserve existing browser behaviour (no regressions).

**Non-Goals:**
- Opera Beta / Developer / GX Beta channels.
- Generic auto-detection of any `Opera Software\*` folder.
- Any change to extraction, report schema, or output format.

## Decisions

- **Add two `BrowserDef` entries** to `DefaultBrowsers` rather than a new discovery mechanism. This matches the documented, existing extension point ("To add a new Chromium-based browser, append a BrowserDef…") and requires no new code paths. Alternative (generic Opera-family scan) was rejected as out of scope.
- **Use `AppDataBase: AppDataRoaming`** for both entries. This is the deviation from other Chromium browsers and is the crux of the fix. `main.go` already selects the base from `AppDataBase`, so no dispatch change is needed.
- **`RelativePath` points at the product root** (`Opera Software\Opera Stable`), not the `Default` subfolder, because `FindProfiles` scans subdirectories for the `History` marker and the confirmed layout places `History` under `...\Opera Stable\Default\`. This keeps multi-profile support (additional named profiles containing `History`) working for free.
- **Extraction reuse:** because `Type: BrowserChromium`, existing `ExtractChromiumHistory`/`ExtractChromiumDownloads` handle Opera unchanged (verified identical `History` schema).

## Risks / Trade-offs

- [Opera is the first Roaming Chromium entry] → `main.go` resolves the base from `AppDataBase` (verified this session); a registry test asserting `AppDataBase == AppDataRoaming` guards against a regression to `Local`.
- [Opera GX path not confirmed on the user's machine] → follows the confirmed Opera layout and is well-documented under `Opera GX Stable`; covered by integration verification and, if needed, runtime E2E on a machine with Opera GX installed.
- [Opera splits data across Local (cache) and Roaming (profile)] → discovery points at Roaming only, where the `History` DB lives; the Local `Opera Software` folder is intentionally ignored.
