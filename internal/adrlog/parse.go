package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var titlePattern = regexp.MustCompile(`^#\s+(\d+)\.\s+(.+)$`)
var datePattern = regexp.MustCompile(`(?m)^Date:\s*(.+)$`)
var exactStatusHeadingPattern = regexp.MustCompile(`(?m)^## Status$`)
var statusHeadingPattern = regexp.MustCompile(`(?m)^##\s+Status\s*$`)
var nextHeadingPattern = regexp.MustCompile(`(?m)^##\s+`)

func ParseRecord(path string) (*Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	return ParseRecordFromContent(path, content)
}

func ParseRecordStrict(path string) (*Record, error) {
	rec, err := ParseRecord(path)
	if err != nil {
		return nil, err
	}
	if err := requireRecordMetadata(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func ParseRecordFromContent(path, content string) (*Record, error) {
	rec := &Record{
		Path:     path,
		Filename: extractFilename(path),
		Content:  content,
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if m := titlePattern.FindStringSubmatch(line); m != nil {
			rec.Number, _ = strconv.Atoi(m[1])
			rec.Title = strings.TrimSpace(m[2])
			break
		}
	}

	if m := datePattern.FindStringSubmatch(content); m != nil {
		rec.Date = strings.TrimSpace(m[1])
	}

	rec.Status = extractStatusLines(content)

	return rec, nil
}

func requireRecordMetadata(rec *Record) error {
	if rec.Number == 0 || rec.Title == "" {
		return fmt.Errorf("%s: missing numbered title", rec.Path)
	}
	if rec.Date == "" {
		return fmt.Errorf("%s: missing date", rec.Path)
	}
	if !statusHeadingPattern.MatchString(rec.Content) {
		return fmt.Errorf("%s: missing ## Status heading", rec.Path)
	}
	return nil
}

func extractStatusLines(content string) []string {
	loc := statusHeadingPattern.FindStringIndex(content)
	if loc == nil {
		return nil
	}

	afterHeading := content[loc[1]:]

	nextLoc := nextHeadingPattern.FindStringIndex(afterHeading)
	var section string
	if nextLoc != nil {
		section = afterHeading[:nextLoc[0]]
	} else {
		section = afterHeading
	}

	var statuses []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			statuses = append(statuses, trimmed)
		}
	}
	return statuses
}

func extractFilename(path string) string {
	return filepath.Base(path)
}

func ExtractLeadingNumber(filename string) int {
	var digits []byte
	for _, c := range []byte(filename) {
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
		} else {
			break
		}
	}
	if len(digits) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(string(digits))
	return n
}
