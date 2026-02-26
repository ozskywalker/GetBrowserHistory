package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// systemDirs are well-known Windows directories under C:\Users\ that are not
// real user profiles and should be skipped during enumeration.
var systemDirs = map[string]struct{}{
	"public":       {},
	"default":      {},
	"default user": {},
	"all users":    {},
}

// EnumerateUsers reads C:\Users\ (or the given usersRoot) and returns the
// names of directories that appear to be real user profiles.
func EnumerateUsers(usersRoot string) ([]string, error) {
	entries, err := os.ReadDir(usersRoot)
	if err != nil {
		return nil, fmt.Errorf("read users root %q: %w", usersRoot, err)
	}

	var users []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if _, skip := systemDirs[strings.ToLower(name)]; skip {
			continue
		}
		users = append(users, name)
	}
	return users, nil
}

// ResolveUser validates that C:\Users\<userName> (or <usersRoot>\<userName>)
// exists and returns its full path. Returns an error if not found.
func ResolveUser(usersRoot, userName string) (string, error) {
	full := filepath.Join(usersRoot, userName)
	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("user profile not found: %s", full)
		}
		return "", fmt.Errorf("stat %q: %w", full, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", full)
	}
	return full, nil
}
