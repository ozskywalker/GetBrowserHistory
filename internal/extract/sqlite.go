package extract

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// QueryDB opens the SQLite database at dbPath, executes the given SQL query
// with optional args, and returns all rows as a slice of string maps.
// Returns (nil, err) on any error, including schema mismatches.
func QueryDB(dbPath, query string, args ...any) ([]map[string]string, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db %q: %w", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}

	var results []map[string]string
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		row := make(map[string]string, len(cols))
		for i, col := range cols {
			switch v := vals[i].(type) {
			case nil:
				row[col] = ""
			case []byte:
				row[col] = string(v)
			default:
				row[col] = fmt.Sprintf("%v", v)
			}
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return results, nil
}
