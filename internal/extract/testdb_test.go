package extract

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// buildTestDB opens (or creates) a SQLite database at dbPath and executes each
// provided statement in sequence. This supplements seedDB for fixtures that
// require multiple CREATE TABLE and INSERT statements across different tables.
func buildTestDB(t *testing.T, dbPath string, stmts ...string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("buildTestDB: open: %v", err)
	}
	defer func() { _ = db.Close() }()
	for _, s := range stmts {
		if s == "" {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("buildTestDB: exec %q: %v", s, err)
		}
	}
}
