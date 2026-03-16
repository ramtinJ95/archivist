package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
)

func (r *Repository) Validate() ([]ValidationIssue, error) {
	files, err := r.ListFiles()
	if err != nil {
		return nil, err
	}

	var issues []ValidationIssue

	numberToFiles := make(map[int][]string)

	for _, f := range files {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(r.CWD, f)
		}

		rec, err := ParseRecord(absPath)
		if err != nil {
			issues = append(issues, ValidationIssue{
				Path:     f,
				Severity: "error",
				Message:  "failed to parse: " + err.Error(),
			})
			continue
		}

		if rec.Title == "" {
			issues = append(issues, ValidationIssue{
				Path:     f,
				Severity: "error",
				Message:  "missing title",
			})
		}

		if rec.Date == "" {
			issues = append(issues, ValidationIssue{
				Path:     f,
				Severity: "error",
				Message:  "missing date",
			})
		}

		if len(rec.Status) == 0 {
			if !statusHeadingPattern.MatchString(rec.Content) {
				issues = append(issues, ValidationIssue{
					Path:     f,
					Severity: "error",
					Message:  "missing ## Status heading",
				})
			} else {
				issues = append(issues, ValidationIssue{
					Path:     f,
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
					Path:     f,
					Severity: "error",
					Message:  fmt.Sprintf("broken relation target: %s", targetFile),
				})
			}
		}

		base := filepath.Base(f)
		fileNum := ExtractLeadingNumber(base)
		if fileNum > 0 {
			numberToFiles[fileNum] = append(numberToFiles[fileNum], f)
		}
		if fileNum > 0 && rec.Number > 0 && fileNum != rec.Number {
			issues = append(issues, ValidationIssue{
				Path:     f,
				Severity: "warning",
				Message:  fmt.Sprintf("filename number %d does not match title number %d", fileNum, rec.Number),
			})
		}
	}

	for num, paths := range numberToFiles {
		if len(paths) < 2 {
			continue
		}
		for _, p := range paths {
			issues = append(issues, ValidationIssue{
				Path:     p,
				Severity: "warning",
				Message:  fmt.Sprintf("duplicate ADR number %d", num),
			})
		}
	}

	return issues, nil
}
