package extract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDatabase copies the SQLite database at src (plus any -wal and -shm
// sidecar files) into a new temporary directory. The caller MUST defer
// os.RemoveAll(tempDir) to ensure cleanup even on error.
//
//	tempDir, dbCopy, err := CopyDatabase(src)
//	if err != nil { ... }
//	defer os.RemoveAll(tempDir)
func CopyDatabase(src string) (tempDir string, destPath string, err error) {
	tempDir, err = os.MkdirTemp("", "bhr_*")
	if err != nil {
		return "", "", fmt.Errorf("create temp dir: %w", err)
	}

	destPath = filepath.Join(tempDir, filepath.Base(src))
	if err := copyFile(src, destPath); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("copy database: %w", err)
	}

	// Copy WAL and SHM sidecars if present (needed for consistent reads).
	for _, suffix := range []string{"-wal", "-shm"} {
		sidecar := src + suffix
		if _, err := os.Stat(sidecar); err == nil {
			dest := destPath + suffix
			// Best-effort: ignore errors on sidecars.
			_ = copyFile(sidecar, dest)
		}
	}

	return tempDir, destPath, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
