package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// webKitEpoch is the number of microseconds between 1601-01-01 and 1970-01-01.
// WebKit timestamps are microseconds since 1601-01-01 00:00:00 UTC.
const webKitEpoch = int64(11644473600 * 1e6)

// WebKitEpochToTime converts a WebKit timestamp (microseconds since 1601-01-01)
// to a Go time.Time in UTC.
func WebKitEpochToTime(microseconds int64) time.Time {
	if microseconds == 0 {
		return time.Time{}
	}
	unixMicro := microseconds - webKitEpoch
	return time.Unix(unixMicro/1e6, (unixMicro%1e6)*1e3).UTC()
}

// ExtractChromiumHistory extracts visited URLs from a Chromium-based browser
// profile directory. It copies the History database before querying to handle
// locked files when the browser is running.
func ExtractChromiumHistory(profilePath string, maxRows int) ([]HistoryRecord, error) {
	dbPath := filepath.Join(profilePath, "History")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("history db not found: %w", err)
	}

	tempDir, dbCopy, err := CopyDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("copy History db: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	query := `
		SELECT u.url, COALESCE(u.title,'') AS title, u.visit_count, v.visit_time
		FROM urls u
		JOIN visits v ON u.id = v.url
		ORDER BY v.visit_time DESC
		LIMIT ?`

	rows, err := QueryDB(dbCopy, query, maxRows)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}

	records := make([]HistoryRecord, 0, len(rows))
	for _, row := range rows {
		vc, _ := strconv.Atoi(row["visit_count"])
		ts, _ := strconv.ParseInt(row["visit_time"], 10, 64)
		q, eng := ExtractSearchQuery(row["url"])
		records = append(records, HistoryRecord{
			URL:          row["url"],
			Title:        row["title"],
			VisitCount:   vc,
			LastVisitUTC: WebKitEpochToTime(ts),
			SearchQuery:  q,
			SearchEngine: eng,
		})
	}
	return records, nil
}

// ExtractChromiumDownloads extracts download records from a Chromium-based
// browser profile directory.
func ExtractChromiumDownloads(profilePath string) ([]DownloadRecord, error) {
	dbPath := filepath.Join(profilePath, "History")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("history db not found: %w", err)
	}

	tempDir, dbCopy, err := CopyDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("copy History db: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	query := `
		SELECT d.target_path, COALESCE(c.url,'') AS source_url, d.start_time,
		       d.total_bytes, COALESCE(d.mime_type,'') AS mime_type
		FROM downloads d
		LEFT JOIN downloads_url_chains c ON d.id = c.id
		ORDER BY d.start_time DESC`

	rows, err := QueryDB(dbCopy, query)
	if err != nil {
		return nil, fmt.Errorf("query downloads: %w", err)
	}

	records := make([]DownloadRecord, 0, len(rows))
	for _, row := range rows {
		ts, _ := strconv.ParseInt(row["start_time"], 10, 64)
		tb, _ := strconv.ParseInt(row["total_bytes"], 10, 64)
		records = append(records, DownloadRecord{
			TargetPath:   row["target_path"],
			SourceURL:    row["source_url"],
			MimeType:     row["mime_type"],
			TotalBytes:   tb,
			StartTimeUTC: WebKitEpochToTime(ts),
		})
	}
	return records, nil
}
