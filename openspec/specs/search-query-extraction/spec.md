# search-query-extraction Specification

## Purpose
TBD - created by archiving change extract-search-queries. Update Purpose after archive.
## Requirements
### Requirement: Extract search query from history record URL
The extraction layer SHALL parse the URL of each `HistoryRecord` and populate a `SearchQuery` field with the human-readable search term when the URL matches a known search engine pattern. The field SHALL remain empty string for all other URLs.

#### Scenario: Google search URL yields query
- **WHEN** a history record URL is `https://www.google.com/search?q=browser+forensics`
- **THEN** `SearchQuery` is set to `"browser forensics"`

#### Scenario: YouTube search URL yields query
- **WHEN** a history record URL is `https://www.youtube.com/results?search_query=go+tutorial`
- **THEN** `SearchQuery` is set to `"go tutorial"`

#### Scenario: DuckDuckGo search URL yields query
- **WHEN** a history record URL is `https://duckduckgo.com/?q=privacy+browser`
- **THEN** `SearchQuery` is set to `"privacy browser"`

#### Scenario: StartPage search URL yields query
- **WHEN** a history record URL is `https://www.startpage.com/search?query=open+source`
- **THEN** `SearchQuery` is set to `"open source"`

#### Scenario: Non-search URL yields empty query
- **WHEN** a history record URL does not match any known search engine pattern
- **THEN** `SearchQuery` is set to `""`

#### Scenario: Malformed URL yields empty query
- **WHEN** a history record URL cannot be parsed by the standard URL parser
- **THEN** `SearchQuery` is set to `""` and no error is returned

#### Scenario: Google country-code domain yields query
- **WHEN** a history record URL host contains `google.` (e.g. `google.co.uk`, `google.fr`)
- **THEN** the `q=` parameter is extracted as `SearchQuery`

#### Scenario: Query string is URL-decoded
- **WHEN** the query parameter value contains percent-encoded characters (e.g. `%2B`, `+`)
- **THEN** `SearchQuery` contains the decoded, human-readable string

