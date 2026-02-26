package extract

import (
	"net/url"
	"strings"
)

// engineDef maps a URL host pattern to the search parameter name and display name.
type engineDef struct {
	param   string // query parameter name (e.g. "q", "search_query")
	display string // human-readable engine name shown in the report
}

// hostEngines maps registered-domain substrings to their engine definitions.
// Checked in order; first match wins.
var hostEngines = []struct {
	hostContains string
	def          engineDef
}{
	{"google.", engineDef{"q", "Google"}},
	{"bing.com", engineDef{"q", "Bing"}},
	{"duckduckgo.com", engineDef{"q", "DuckDuckGo"}},
	{"yahoo.com", engineDef{"p", "Yahoo"}},
	{"youtube.com", engineDef{"search_query", "YouTube"}},
	{"ecosia.org", engineDef{"q", "Ecosia"}},
	{"search.brave.com", engineDef{"q", "Brave Search"}},
	{"startpage.com", engineDef{"query", "StartPage"}},
}

// ExtractSearchQuery inspects rawURL and returns the decoded search query and
// engine display name if the URL belongs to a known search engine. Returns
// empty strings for unrecognised URLs or when the URL cannot be parsed.
func ExtractSearchQuery(rawURL string) (query, engine string) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "", ""
	}

	host := strings.ToLower(u.Host)

	for _, e := range hostEngines {
		if strings.Contains(host, e.hostContains) {
			q := u.Query().Get(e.def.param)
			if q == "" {
				return "", ""
			}
			return q, e.def.display
		}
	}
	return "", ""
}
