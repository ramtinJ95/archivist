package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (r *Repository) Validate() ([]ValidationIssue, error) {
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

	var issues []ValidationIssue
	var records []*Record
	numberToFiles := make(map[int][]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		relPath := filepath.Join(r.ADRDir, entry.Name())
		absPath := filepath.Join(absDir, entry.Name())

		filenameIssue, err := validateFilenameShape(relPath, absPath, entry.Name())
		if err != nil {
			return nil, err
		}
		if filenameIssue != nil {
			issues = append(issues, *filenameIssue)
		}

		if !adrFilenamePattern.MatchString(entry.Name()) {
			continue
		}

		if digitWidth := leadingDigitWidth(entry.Name()); digitWidth > 0 && digitWidth != 4 {
			issues = append(issues, ValidationIssue{
				Path:     relPath,
				Severity: "warning",
				Message:  fmt.Sprintf("filename number should be zero-padded to 4 digits, got %d", digitWidth),
			})
		}

		rec, err := ParseRecord(absPath)
		if err != nil {
			issues = append(issues, ValidationIssue{
				Path:     relPath,
				Severity: "error",
				Message:  "failed to parse: " + err.Error(),
			})
			continue
		}
		rec.Path = relPath
		records = append(records, rec)

		if rec.Title == "" {
			issues = append(issues, ValidationIssue{
				Path:     relPath,
				Severity: "error",
				Message:  "missing title",
			})
		}

		if rec.Date == "" {
			issues = append(issues, ValidationIssue{
				Path:     relPath,
				Severity: "error",
				Message:  "missing date",
			})
		}

		if len(rec.Status) == 0 {
			if !statusHeadingPattern.MatchString(rec.Content) {
				issues = append(issues, ValidationIssue{
					Path:     relPath,
					Severity: "error",
					Message:  "missing ## Status heading",
				})
			} else {
				issues = append(issues, ValidationIssue{
					Path:     relPath,
					Severity: "warning",
					Message:  "empty status section",
				})
			}
		}

		for _, statusLine := range rec.Status {
			m := relationPattern.FindStringSubmatch(statusLine)
			if m == nil {
				continue
			}
			targetFile := m[3]
			targetPath := filepath.Join(r.AbsADRDir(), targetFile)
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				issues = append(issues, ValidationIssue{
					Path:     relPath,
					Severity: "error",
					Message:  fmt.Sprintf("broken relation target: %s", targetFile),
				})
			}
		}

		base := filepath.Base(relPath)
		fileNum := ExtractLeadingNumber(base)
		if fileNum > 0 {
			numberToFiles[fileNum] = append(numberToFiles[fileNum], relPath)
		}
		if fileNum > 0 && rec.Number > 0 && fileNum != rec.Number {
			issues = append(issues, ValidationIssue{
				Path:     relPath,
				Severity: "warning",
				Message:  fmt.Sprintf("filename number %d does not match title number %d", fileNum, rec.Number),
			})
		}
	}

	duplicateNumbers := sortedKeys(numberToFiles)
	for _, num := range duplicateNumbers {
		paths := numberToFiles[num]
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		for _, p := range paths {
			issues = append(issues, ValidationIssue{
				Path:     p,
				Severity: "warning",
				Message:  fmt.Sprintf("duplicate ADR number %d", num),
			})
		}
	}

	issues = append(issues, ambiguousRefIssues(r.ADRDir, records)...)
	sortValidationIssues(issues)

	return issues, nil
}

func validateFilenameShape(relPath, absPath, filename string) (*ValidationIssue, error) {
	if adrFilenamePattern.MatchString(filename) {
		return nil, nil
	}

	if filepath.Ext(filename) != ".md" && ExtractLeadingNumber(filename) == 0 {
		return nil, nil
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return &ValidationIssue{
			Path:     relPath,
			Severity: "error",
			Message:  "failed to inspect ADR candidate: " + err.Error(),
		}, nil
	}

	rec, err := ParseRecordFromContent(relPath, string(data))
	if err != nil {
		return &ValidationIssue{
			Path:     relPath,
			Severity: "error",
			Message:  "failed to inspect ADR candidate: " + err.Error(),
		}, nil
	}
	if ExtractLeadingNumber(filename) == 0 && rec.Number == 0 {
		return nil, nil
	}

	return &ValidationIssue{
		Path:     relPath,
		Severity: "warning",
		Message:  "malformed ADR filename: expected NNNN-title.md",
	}, nil
}

func ambiguousRefIssues(adrDir string, records []*Record) []ValidationIssue {
	fragmentPaths := make(map[string]map[string]struct{})

	for i := 0; i < len(records); i++ {
		leftTokens := slugTokens(records[i].Filename)
		for j := i + 1; j < len(records); j++ {
			fragment := longestSharedSlugFragment(leftTokens, slugTokens(records[j].Filename))
			if fragment == "" {
				continue
			}

			paths := fragmentPaths[fragment]
			if paths == nil {
				paths = make(map[string]struct{})
				fragmentPaths[fragment] = paths
			}
			paths[records[i].Path] = struct{}{}
			paths[records[j].Path] = struct{}{}
		}
	}

	fragments := make([]string, 0, len(fragmentPaths))
	for fragment := range fragmentPaths {
		fragments = append(fragments, fragment)
	}
	sort.Strings(fragments)

	issues := make([]ValidationIssue, 0, len(fragments))
	for _, fragment := range fragments {
		paths := make([]string, 0, len(fragmentPaths[fragment]))
		for path := range fragmentPaths[fragment] {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		issues = append(issues, ValidationIssue{
			Path:     adrDir,
			Severity: "warning",
			Message:  fmt.Sprintf("ambiguous partial ref %q matches %s", fragment, strings.Join(paths, ", ")),
		})
	}

	return issues
}

func slugTokens(filename string) []string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	base = strings.TrimLeft(base, "0123456789")
	base = strings.TrimPrefix(base, "-")
	if base == "" {
		return nil
	}

	parts := strings.Split(base, "-")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return tokens
}

func longestSharedSlugFragment(left, right []string) string {
	bestLen := 0
	bestStart := -1

	for i := 0; i < len(left); i++ {
		for j := 0; j < len(right); j++ {
			k := 0
			for i+k < len(left) && j+k < len(right) && left[i+k] == right[j+k] {
				k++
			}
			if k == 0 {
				continue
			}

			fragment := left[i : i+k]
			if !sharedFragmentQualifies(fragment) {
				continue
			}
			if k > bestLen {
				bestLen = k
				bestStart = i
			}
		}
	}

	if bestLen == 0 || bestStart < 0 {
		return ""
	}

	return strings.Join(left[bestStart:bestStart+bestLen], "-")
}

func sharedFragmentQualifies(tokens []string) bool {
	return len(tokens) >= 2
}

func leadingDigitWidth(filename string) int {
	width := 0
	for _, r := range filename {
		if r < '0' || r > '9' {
			break
		}
		width++
	}
	return width
}

func sortedKeys(values map[int][]string) []int {
	keys := make([]int, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func sortValidationIssues(issues []ValidationIssue) {
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		if issues[i].Severity != issues[j].Severity {
			return issues[i].Severity < issues[j].Severity
		}
		return issues[i].Message < issues[j].Message
	})
}
