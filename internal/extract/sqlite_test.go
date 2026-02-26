package extract

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// seedDB creates a SQLite file at dbPath, runs ddl+dml, and closes it.
// Returns the path to the file (same as dbPath).
func seedDB(t *testing.T, dbPath, ddl, dml string) string {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("seed: open: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("seed: ddl: %v", err)
	}
	if dml != "" {
		if _, err := db.Exec(dml); err != nil {
			t.Fatalf("seed: dml: %v", err)
		}
	}
	return dbPath
}

func TestQueryDB(t *testing.T) {
	t.Run("returns rows as string maps", func(t *testing.T) {
		dbPath := seedDB(t,
			filepath.Join(t.TempDir(), "test.db"),
			`CREATE TABLE items (name TEXT, count INTEGER)`,
			`INSERT INTO items VALUES ('apple', 3), ('banana', 7)`,
		)

		rows, err := QueryDB(dbPath, `SELECT name, count FROM items ORDER BY name`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		if rows[0]["name"] != "apple" {
			t.Errorf("row 0 name: got %q, want %q", rows[0]["name"], "apple")
		}
		if rows[0]["count"] != "3" {
			t.Errorf("row 0 count: got %q, want %q", rows[0]["count"], "3")
		}
		if rows[1]["name"] != "banana" {
			t.Errorf("row 1 name: got %q, want %q", rows[1]["name"], "banana")
		}
	})

	t.Run("NULL column value coerced to empty string", func(t *testing.T) {
		dbPath := seedDB(t,
			filepath.Join(t.TempDir(), "nulls.db"),
			`CREATE TABLE t (a TEXT, b TEXT)`,
			`INSERT INTO t VALUES (NULL, 'hello')`,
		)

		rows, err := QueryDB(dbPath, `SELECT a, b FROM t`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		if rows[0]["a"] != "" {
			t.Errorf("NULL column: got %q, want empty string", rows[0]["a"])
		}
		if rows[0]["b"] != "hello" {
			t.Errorf("non-NULL column: got %q, want %q", rows[0]["b"], "hello")
		}
	})

	t.Run("empty result set returns nil slice without error", func(t *testing.T) {
		dbPath := seedDB(t,
			filepath.Join(t.TempDir(), "empty.db"),
			`CREATE TABLE t (x TEXT)`,
			"",
		)

		rows, err := QueryDB(dbPath, `SELECT x FROM t`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rows) != 0 {
			t.Errorf("expected 0 rows, got %d", len(rows))
		}
	})

	t.Run("query with args applies limit", func(t *testing.T) {
		dbPath := seedDB(t,
			filepath.Join(t.TempDir(), "args.db"),
			`CREATE TABLE t (n INTEGER)`,
			`INSERT INTO t VALUES (1),(2),(3),(4),(5)`,
		)

		rows, err := QueryDB(dbPath, `SELECT n FROM t ORDER BY n LIMIT ?`, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rows) != 3 {
			t.Errorf("expected 3 rows with LIMIT 3, got %d", len(rows))
		}
	})

	t.Run("invalid SQL returns error", func(t *testing.T) {
		dbPath := seedDB(t,
			filepath.Join(t.TempDir(), "err.db"),
			`CREATE TABLE t (x TEXT)`,
			"",
		)

		_, err := QueryDB(dbPath, `SELECT * FROM nonexistent_table`)
		if err == nil {
			t.Error("expected error for query against non-existent table, got nil")
		}
	})

	t.Run("corrupted file returns error", func(t *testing.T) {
		// A file with non-SQLite content must cause a query error.
		bad := filepath.Join(t.TempDir(), "corrupt.db")
		if err := os.WriteFile(bad, []byte("this is not sqlite"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := QueryDB(bad, `SELECT * FROM sqlite_master`)
		if err == nil {
			t.Error("expected error for corrupted database file, got nil")
		}
	})
}
