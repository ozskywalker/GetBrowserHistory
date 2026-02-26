package report

import (
	"time"

	"github.com/ozskywalker/GetBrowserHistory/internal/extract"
)

// ReportMeta holds metadata about the report generation run.
type ReportMeta struct {
	GeneratedAt      time.Time `json:"generatedAt"`
	Hostname         string    `json:"hostname"`
	ExecutingAccount string    `json:"executingAccount"`
	Version          string    `json:"scriptVersion"`
}

// ProfileData holds extracted data for a single browser profile.
type ProfileData struct {
	BrowserName string                 `json:"browserName"`
	ProfilePath string                 `json:"profilePath"`
	History     []extract.HistoryRecord  `json:"history"`
	Downloads   []extract.DownloadRecord `json:"downloads"`
	// Truncated is true when history results were capped by --max-rows.
	Truncated   bool `json:"truncated,omitempty"`
	TruncatedAt int  `json:"truncatedAt,omitempty"`
}

// UserData holds all browser profiles discovered for one Windows user.
type UserData struct {
	Username string        `json:"username"`
	Profiles []ProfileData `json:"profiles"`
}

// Report is the top-level structure written to both the HTML and JSON outputs.
type Report struct {
	Meta     ReportMeta `json:"meta"`
	Users    []UserData `json:"users"`
	Warnings []string   `json:"warnings,omitempty"`
}
