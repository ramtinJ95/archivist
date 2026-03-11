package adrlog

import (
	"path/filepath"
)

func (r *Repository) Validate() ([]ValidationIssue, error) {
	files, err := r.ListFiles()
	if err != nil {
		return nil, err
	}

	var issues []ValidationIssue

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
			issues = append(issues, ValidationIssue{
				Path:     f,
				Severity: "warning",
				Message:  "missing status",
			})
		}
	}

	return issues, nil
}
