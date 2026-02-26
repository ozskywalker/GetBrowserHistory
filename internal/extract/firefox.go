package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// PRTimeToTime converts a Firefox PRTime value (microseconds since Unix epoch)
// to a Go time.Time in UTC.
func PRTimeToTime(microseconds int64) time.Time {
	if microseconds == 0 {
		return time.Time{}
	}
	return time.Unix(microseconds/1e6, (microseconds%1e6)*1e3).UTC()
}

// ExtractFirefoxHistory extracts visited URLs from a Firefox profile directory.
// It copies places.sqlite before querying to handle locked files.
func ExtractFirefoxHistory(profilePath string, maxRows int) ([]HistoryRecord, error) {
	dbPath := filepath.Join(profilePath, "places.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("places.sqlite not found: %w", err)
	}

	tempDir, dbCopy, err := CopyDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("copy places.sqlite: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	query := `
		SELECT p.url, COALESCE(p.title,'') AS title,
		       p.visit_count, h.visit_date
		FROM moz_places p
		JOIN moz_historyvisits h ON p.id = h.place_id
		ORDER BY h.visit_date DESC
		LIMIT ?`

	rows, err := QueryDB(dbCopy, query, maxRows)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}

	records := make([]HistoryRecord, 0, len(rows))
	for _, row := range rows {
		vc, _ := strconv.Atoi(row["visit_count"])
		ts, _ := strconv.ParseInt(row["visit_date"], 10, 64)
		q, eng := ExtractSearchQuery(row["url"])
		records = append(records, HistoryRecord{
			URL:          row["url"],
			Title:        row["title"],
			VisitCount:   vc,
			LastVisitUTC: PRTimeToTime(ts),
			SearchQuery:  q,
			SearchEngine: eng,
		})
	}
	return records, nil
}

// ExtractFirefoxDownloads extracts download records from a Firefox profile
// directory by querying the moz_annos table for download destination URIs.
func ExtractFirefoxDownloads(profilePath string) ([]DownloadRecord, error) {
	dbPath := filepath.Join(profilePath, "places.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("places.sqlite not found: %w", err)
	}

	tempDir, dbCopy, err := CopyDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("copy places.sqlite: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// The moz_annos table stores download metadata as annotations on places.
	// The content column for downloads/destinationFileURI holds the local path.
	query := `
		SELECT p.url AS source_url,
		       a.content AS dest_path,
		       a.dateAdded AS date_added
		FROM moz_places p
		JOIN moz_annos a ON p.id = a.place_id
		JOIN moz_anno_attributes attr ON a.anno_attribute_id = attr.id
		WHERE attr.name = 'downloads/destinationFileURI'
		ORDER BY a.dateAdded DESC`

	rows, err := QueryDB(dbCopy, query)
	if err != nil {
		return nil, fmt.Errorf("query downloads: %w", err)
	}

	records := make([]DownloadRecord, 0, len(rows))
	for _, row := range rows {
		ts, _ := strconv.ParseInt(row["date_added"], 10, 64)
		destPath := row["dest_path"]
		// Firefox stores the destination as a file:// URI; strip the prefix.
		if len(destPath) > 7 && destPath[:7] == "file://" {
			destPath = destPath[7:]
		}
		records = append(records, DownloadRecord{
			TargetPath:   destPath,
			SourceURL:    row["source_url"],
			StartTimeUTC: PRTimeToTime(ts),
		})
	}
	return records, nil
}
