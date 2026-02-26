package report

import "encoding/json"

// RenderJSON serializes the Report to indented JSON.
// time.Time fields serialize as RFC 3339 / ISO 8601 UTC strings by default.
func RenderJSON(r Report) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
