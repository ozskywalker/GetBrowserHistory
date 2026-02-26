## 1. Data Model

- [x] 1.1 Add `SearchQuery string` field (with `json:"searchQuery"` tag) to `HistoryRecord` in `internal/extract/records.go`

## 2. Search Query Extraction

- [x] 2.1 Create `internal/extract/search.go` with a static engine map (hostname → parameter name) covering Google (all `*.google.*` domains), Bing, DuckDuckGo, Yahoo, YouTube, Ecosia, Brave Search, and StartPage
- [x] 2.2 Implement `ExtractSearchQuery(rawURL string) (query, engine string)` in `search.go` using `net/url.Parse`, returning decoded query text and display engine name (or empty strings on no match / parse error)
- [x] 2.3 Write unit tests in `internal/extract/search_test.go` covering: Google, YouTube, DuckDuckGo, StartPage, non-search URL, malformed URL, Google country-code domain, percent-encoded query

## 3. Wire Extraction into History Records

- [x] 3.1 Call `ExtractSearchQuery` for each `HistoryRecord` assembled in `internal/extract/chromium.go` and populate `SearchQuery`
- [x] 3.2 Call `ExtractSearchQuery` for each `HistoryRecord` assembled in `internal/extract/firefox.go` and populate `SearchQuery`

## 4. HTML Report — History Table

- [x] 4.1 Add a **Search Query** `<th>` header (5th column, sortable) to the history table in `internal/report/template.html`
- [x] 4.2 Add a `<td>{{.SearchQuery}}</td>` cell to each history row in the template

## 5. HTML Report — Search History Table

- [x] 5.1 Add a helper to the template (or pre-compute in Go) that filters a profile's history records to those with non-empty `SearchQuery`
- [x] 5.2 Render a Search History `<table>` block above the History table in the template — columns: Timestamp (UTC), Engine, Query, URL — only when filtered list is non-empty
- [x] 5.3 Wire the Search History table into the existing `sortTable` JavaScript function (unique table ID per profile)
- [x] 5.4 Add an engine-name display helper: since `ExtractSearchQuery` now returns engine name, expose it on `HistoryRecord` or use a template map function so the Search History table can show "Google", "YouTube", etc.

## 6. Verification

- [x] 6.1 Build the binary (`go build ./...`) and confirm no compile errors
- [x] 6.2 Run all tests (`go test ./...`) and confirm they pass
- [x] 6.3 Generate a test report against a real or sample browser profile and verify: Search History table appears, History table has Search Query column populated, JSON output contains `searchQuery` field
