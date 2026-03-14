package api

import (
	"os"
	"path/filepath"
)

func setupTestEnv(tmpDir string) (string, error) {
	originalWd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Find project root
	root := originalWd
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}

	if err := os.Chdir(tmpDir); err != nil {
		return "", err
	}

	// Create symlink to api templates so parseTemplate works
	if err := os.MkdirAll("api", 0755); err != nil {
		return "", err
	}
	if err := os.Symlink(filepath.Join(root, "api/templates"), "api/templates"); err != nil {
		// If it already exists, it's fine
		if !os.IsExist(err) {
			return "", err
		}
	}

	// Create symlink to help_docs
	if err := os.Symlink(filepath.Join(root, "help_docs"), "help_docs_real"); err != nil {
		if !os.IsExist(err) {
			return "", err
		}
	}

	return originalWd, nil
}
