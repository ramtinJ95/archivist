package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SearchMatch struct {
	Path    string
	Line    int
	Content string
}

type SearchResult struct {
	Path    string
	Matches []SearchMatch
}

func (r *Repository) Search(pattern string) ([]SearchResult, error) {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	files, err := r.ListFiles()
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, f := range files {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(r.CWD, f)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		var matches []SearchMatch
		for i, line := range lines {
			if re.MatchString(line) {
				matches = append(matches, SearchMatch{
					Path:    f,
					Line:    i + 1,
					Content: line,
				})
			}
		}

		if len(matches) > 0 {
			results = append(results, SearchResult{
				Path:    f,
				Matches: matches,
			})
		}
	}

	return results, nil
}
