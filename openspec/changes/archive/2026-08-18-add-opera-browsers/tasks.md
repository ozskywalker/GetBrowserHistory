## 1. Browser Registry

- [x] 1.1 Add an `Opera` `BrowserDef` to `DefaultBrowsers` in `internal/browser/browsers.go` with `RelativePath: Opera Software\Opera Stable`, `AppDataBase: AppDataRoaming`, `Type: BrowserChromium`
- [x] 1.2 Add an `Opera GX` `BrowserDef` to `DefaultBrowsers` with `RelativePath: Opera Software\Opera GX Stable`, `AppDataBase: AppDataRoaming`, `Type: BrowserChromium`

## 2. Unit Tests

- [x] 2.1 Add a `FindProfiles` test for a Roaming Chromium (Opera) def discovering a `Default` profile containing `History`
- [x] 2.2 Add a `FindProfiles` test for Opera GX discovering a `Default` profile containing `History`
- [x] 2.3 Add a `FindProfiles` test that Opera-not-installed returns an empty slice without error
- [x] 2.4 Add a `DefaultBrowsers` registry test asserting both Opera entries have `AppDataBase == AppDataRoaming`, `Type == BrowserChromium`, and the expected `RelativePath`

## 3. Documentation

- [x] 3.1 Update the README feature bullet to include Opera and Opera GX
- [x] 3.2 Add Opera and Opera GX rows to the README Supported Browsers table
- [x] 3.3 Update `openspec/specs/browser-data-extraction/spec.md` discovery requirement to mention Opera and Opera GX (via this change's delta spec, applied at archive)

## 4. Verification

- [x] 4.1 Build the binary (`go build ./...`) with no compile errors
- [x] 4.2 Run all tests (`go test ./...`) and confirm they pass with no regressions
- [x] 4.3 Run `go vet ./...` with no findings
