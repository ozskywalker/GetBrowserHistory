package extract

import "time"

// HistoryRecord represents a single browser history visit.
type HistoryRecord struct {
	URL          string
	Title        string
	VisitCount   int
	LastVisitUTC time.Time
	SearchQuery  string `json:"searchQuery"`
	SearchEngine string `json:"searchEngine,omitempty"`
}

// DownloadRecord represents a single browser download entry.
type DownloadRecord struct {
	TargetPath   string
	SourceURL    string
	MimeType     string
	TotalBytes   int64
	StartTimeUTC time.Time
}
