package report

import (
	"encoding/json"
	"fmt"
	"os"
)

// RenderJSON serializes the Report to indented JSON.
// time.Time fields serialize as RFC 3339 / ISO 8601 UTC strings by default.
func RenderJSON(r Report) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// LoadJSON reads a report.json file from disk and deserializes it into a Report.
func LoadJSON(path string) (Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var r Report
	if err := json.Unmarshal(data, &r); err != nil {
		return Report{}, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	return r, nil
}
