## Context

The tool extracts Chromium and Firefox history records from SQLite databases and renders them into an HTML/JSON report. Each `HistoryRecord` currently carries raw URL, title, visit count, and timestamp. No post-processing of URLs is done. Users reviewing the report must manually inspect URLs to identify what was searched, which is slow when tables contain thousands of rows.

## Goals / Non-Goals

**Goals:**
- Enrich `HistoryRecord` with an optional `SearchQuery` string populated by parsing the URL at extraction time.
- Support the most common search engines via `q=`, `query=`, and `search_query=` query parameters.
- Surface extracted queries in a per-profile **Search History** table (timestamp, engine, query, URL) rendered above the History table in the HTML report.
- Add a **Search Query** column to the existing History table (empty for non-search rows).
- Include `searchQuery` in JSON output automatically via the existing struct serialisation path.

**Non-Goals:**
- Detecting search intent from arbitrary URLs (only well-known parameter names on recognised hostnames).
- Tracking which specific engine variant was used (e.g. Google Images vs Google Web) — engine name is derived from the registered domain only.
- Modifying the downloads table or download records.
- Adding any new CLI flags or configuration options.

## Decisions

### 1. Parsing at extraction time vs. report time

**Decision**: Parse during extraction, storing `SearchQuery` on `HistoryRecord`.

**Rationale**: Keeps report rendering logic simple (template just reads a field). The JSON output automatically includes the field. Centralises URL logic in the `extract` package where data originates.

**Alternative considered**: Parse in the report template or in `report/html.go`. Rejected because it would duplicate logic across HTML and JSON paths and make the template harder to read.

---

### 2. Engine detection strategy

**Decision**: Match the registered domain of the URL against a static allowlist, then read the appropriate parameter name for that engine.

```
google.com, google.<cc>  → q=
bing.com                 → q=
duckduckgo.com           → q=
yahoo.com                → p= (also q= for some paths)
youtube.com              → search_query=
ecosia.org               → q=
brave.com (search)       → q=
startpage.com            → query=
```

**Rationale**: Simple, zero-dependency, fast. Covers the engines that appear in practice. False negatives (unrecognised engines) are silent — the column is just empty, which is safe.

**Alternative considered**: Try all three parameter names on every URL regardless of host. Rejected because it would produce false positives (e.g. any URL with `?q=` would appear as a search result).

---

### 3. New file vs. extending existing files

**Decision**: Add a new `internal/extract/search.go` containing the engine map and `ExtractSearchQuery(rawURL string) string` function. Call it from wherever history records are assembled (currently in `chromium.go` and `firefox.go`).

**Rationale**: Keeps the extraction concern isolated and unit-testable without touching existing files beyond the call sites and the struct definition.

---

### 4. Search History table placement

**Decision**: Render the Search History table immediately above the History table, within the same per-profile collapsible section. Only show the table when at least one search record exists for that profile.

**Rationale**: Co-locating search history with the profile it belongs to preserves the existing user/browser/profile hierarchy. Hiding the table when empty avoids visual noise.

## Risks / Trade-offs

- **Google country-code domains** (google.co.uk, google.fr, etc.) — Mitigation: match on suffix `.google.` anywhere in the registered domain, or maintain a list of known cc-TLDs. A simple `strings.Contains(host, "google.")` approach is sufficient given the allowlist strategy.
- **URL parse errors** — Mitigation: `net/url.Parse` errors are silently ignored; `SearchQuery` remains empty string.
- **Performance** — URL parsing adds negligible overhead; `net/url.Parse` is O(n) on URL length and called once per history record.
- **Missed engines** — Non-goal; the allowlist can be extended trivially in future.
