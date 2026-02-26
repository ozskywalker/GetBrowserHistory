package report

import (
	"bytes"
	_ "embed"
	"html/template"
	"path/filepath"

	"github.com/ozskywalker/GetBrowserHistory/internal/extract"
)

//go:embed template.html
var rawTemplate string

// base returns the final path component of a filesystem path.
// Registered as a template function so {{base .ProfilePath}} works.
func base(path string) string {
	return filepath.Base(path)
}

// searchRecords filters a history slice to records that have a non-empty
// SearchQuery, used by the template to render the Search History table.
func searchRecords(records []extract.HistoryRecord) []extract.HistoryRecord {
	var out []extract.HistoryRecord
	for _, r := range records {
		if r.SearchQuery != "" {
			out = append(out, r)
		}
	}
	return out
}

var reportTemplate = template.Must(
	template.New("report").
		Funcs(template.FuncMap{
			"base":          base,
			"searchRecords": searchRecords,
		}).
		Parse(rawTemplate),
)

// RenderHTML executes the embedded HTML template against the given Report and
// returns the rendered bytes. html/template auto-escapes all user data.
func RenderHTML(r Report) ([]byte, error) {
	var buf bytes.Buffer
	if err := reportTemplate.Execute(&buf, r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
