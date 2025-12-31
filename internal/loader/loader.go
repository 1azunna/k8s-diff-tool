package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadFile reads a file from the filesystem.
func LoadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

// IsDir checks if a path is a directory.
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ListYAMLFiles returns a list of YAML filenames (non-recursive) in a directory.
func ListYAMLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".yaml" || ext == ".yml" {
				files = append(files, name)
			}
		}
	}
	return files, nil
}
