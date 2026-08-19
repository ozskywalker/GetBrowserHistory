## Why

Opera and Opera GX (modern Chromium-based browsers) are widely deployed but currently invisible to the history report — the utility only discovers Chrome, Edge, Brave, DuckDuckGo, and Firefox. Unlike other Chromium browsers, Opera keeps its profile under `AppData\Roaming`, so it is not picked up by the existing Local-based discovery.

## What Changes

- Add **Opera** and **Opera GX** to the `DefaultBrowsers` registry as Chromium-based browsers.
- Discover their profiles under `%APPDATA%\Opera Software\Opera Stable` and `%APPDATA%\Opera Software\Opera GX Stable` respectively (the first Chromium entries that use Roaming AppData rather than Local).
- Reuse the existing Chromium `History`/downloads extraction unchanged — no new extraction code.
- Update the README supported-browsers table.

No BREAKING changes.

## Capabilities

### New Capabilities

<!-- None — no new capability is introduced. -->

### Modified Capabilities

- `browser-data-extraction`: Extend browser profile discovery to include the Opera and Opera GX Chromium-based browsers, whose profiles are located under Roaming AppData (a deviation from the Local AppData location used by other Chromium browsers).

## Impact

- `internal/browser/browsers.go` — two new `BrowserDef` entries (`Opera`, `Opera GX`) in `DefaultBrowsers` with `AppDataBase: AppDataRoaming` and `Type: BrowserChromium`.
- `internal/browser/browsers_test.go` — new unit tests for Roaming Chromium discovery, not-installed behaviour, and registry field correctness.
- `README.md` — supported-browsers table and feature bullet list.
- `openspec/specs/browser-data-extraction/spec.md` — update the discovery requirement to mention Opera.
- No changes to `internal/extract/*`, `cmd/browser-report/main.go`, dependencies, or output format.
