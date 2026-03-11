package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var adrFilenamePattern = regexp.MustCompile(`^\d{4}-.*\.md$`)

func (r *Repository) ListFiles() ([]string, error) {
	absDir := r.AbsADRDir()

	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("%s is not a directory", r.ADRDir)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", r.ADRDir)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if adrFilenamePattern.MatchString(e.Name()) {
			files = append(files, filepath.Join(r.ADRDir, e.Name()))
		}
	}

	sort.Strings(files)
	return files, nil
}
