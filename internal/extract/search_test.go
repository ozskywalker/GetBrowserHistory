package extract

import "testing"

func TestExtractSearchQuery(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantQuery  string
		wantEngine string
	}{
		{
			name:       "Google search",
			url:        "https://www.google.com/search?q=browser+forensics",
			wantQuery:  "browser forensics",
			wantEngine: "Google",
		},
		{
			name:       "Google country-code domain",
			url:        "https://www.google.co.uk/search?q=open+source+tools",
			wantQuery:  "open source tools",
			wantEngine: "Google",
		},
		{
			name:       "YouTube search",
			url:        "https://www.youtube.com/results?search_query=go+tutorial",
			wantQuery:  "go tutorial",
			wantEngine: "YouTube",
		},
		{
			name:       "DuckDuckGo search",
			url:        "https://duckduckgo.com/?q=privacy+browser",
			wantQuery:  "privacy browser",
			wantEngine: "DuckDuckGo",
		},
		{
			name:       "StartPage search",
			url:        "https://www.startpage.com/search?query=open+source",
			wantQuery:  "open source",
			wantEngine: "StartPage",
		},
		{
			name:       "Bing search",
			url:        "https://www.bing.com/search?q=golang+windows",
			wantQuery:  "golang windows",
			wantEngine: "Bing",
		},
		{
			name:       "Ecosia search",
			url:        "https://www.ecosia.org/search?q=green+tech",
			wantQuery:  "green tech",
			wantEngine: "Ecosia",
		},
		{
			name:       "Brave search",
			url:        "https://search.brave.com/search?q=privacy",
			wantQuery:  "privacy",
			wantEngine: "Brave Search",
		},
		{
			name:       "Percent-encoded query",
			url:        "https://www.google.com/search?q=C%2B%2B+programming",
			wantQuery:  "C++ programming",
			wantEngine: "Google",
		},
		{
			name:       "Non-search URL",
			url:        "https://github.com/golang/go",
			wantQuery:  "",
			wantEngine: "",
		},
		{
			name:       "Malformed URL",
			url:        "://not a url",
			wantQuery:  "",
			wantEngine: "",
		},
		{
			name:       "Search engine URL without query param",
			url:        "https://www.google.com/",
			wantQuery:  "",
			wantEngine: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotQuery, gotEngine := ExtractSearchQuery(tc.url)
			if gotQuery != tc.wantQuery {
				t.Errorf("query: got %q, want %q", gotQuery, tc.wantQuery)
			}
			if gotEngine != tc.wantEngine {
				t.Errorf("engine: got %q, want %q", gotEngine, tc.wantEngine)
			}
		})
	}
}
