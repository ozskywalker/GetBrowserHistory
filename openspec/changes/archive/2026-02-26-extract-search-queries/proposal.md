## Why

Browser history records contain search engine visits that reveal what users were actively searching for, but this intent is buried inside raw URLs. Surfacing extracted search queries as first-class data makes forensic review significantly faster.

## What Changes

- A new URL-parsing step enriches each `HistoryRecord` with the search query string extracted from known search engine URL patterns (`q=`, `query=`, `search_query=` parameters on Google, Bing, DuckDuckGo, Yahoo, YouTube, and similar engines).
- The HTML report history table gains a **Search Query** column (empty for non-search URLs).
- A new **Search History** summary table is rendered above the per-profile History table, listing only the records where a query was extracted (timestamp, engine, query, URL).
- The JSON output includes the `searchQuery` field on each history record.

## Capabilities

### New Capabilities
- `search-query-extraction`: Parse history record URLs to identify known search engine patterns and extract the human-readable query string from query parameters (`q`, `query`, `search_query`).

### Modified Capabilities
- `report-generation`: Add a Search Query column to the history table and render a Search History summary table above the History table for each browser profile.

## Impact

- `internal/extract/records.go` — `HistoryRecord` gains a `SearchQuery string` field.
- `internal/extract/` — new `search.go` file implementing URL parsing and engine detection.
- `internal/report/schema.go` — no struct changes needed (field is on `HistoryRecord`).
- `internal/report/template.html` — two template changes: new Search History table block, new column in History table.
- `internal/report/json.go` — zero changes needed; `SearchQuery` serialises automatically via the struct tag.
- No new dependencies; uses standard library `net/url`.
