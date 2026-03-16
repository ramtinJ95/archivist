package adrlog

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var titlePattern = regexp.MustCompile(`^#\s+(\d+)\.\s+(.+)$`)
var datePattern = regexp.MustCompile(`(?m)^Date:\s*(.+)$`)
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
