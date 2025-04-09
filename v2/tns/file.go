package tns

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ResolveFilepath resolves a path to a TNS names file.
//
// Returns a non-empty string if a TNS names file has been found at the path specified or
// such a file has been found in "${TNS_ADMIN}/" folder.
func ResolveFilepath(explicitPath string) (path string, err error) {
	if explicitPath != "" {
		// We've specified TNS names filepath explicitly => it must be present
		exists, err := isExistingFile(explicitPath)
		if err != nil {
			return "", err
		}
		if exists {
			return explicitPath, nil
		}
		return "", fmt.Errorf("TNS Names file '%s' doesn't exist", explicitPath)
	}

	// We haven't specified TNS names filepath explicitly => try to locate the file in "${TNS_ADMIN}/",
	// like Oracle Instant Client. In this case existence of TNS names file is not required.
	tnsAdminDir := os.Getenv("TNS_ADMIN")
	if tnsAdminDir == "" {
		return "", nil
	}
	path = filepath.Join(tnsAdminDir, "tnsnames.ora")
	exists, err := isExistingFile(path)
	if err != nil {
		return "", err
	}
	if exists {
		// Found a TNS names file in "${TNS_ADMIN}/".
		return path, nil
	}
	// "tnsnames.ora" file hasn't been found in "${TNS_ADMIN}/". That's fine, just return an empty string
	return "", nil
}

// isExistingFile returns true if a file at specified path exists and is an actual file, false otherwise.
func isExistingFile(path string) (ok bool, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("cannot stat file '%s': %w", path, err)
	}
	return !stat.IsDir(), nil
}
