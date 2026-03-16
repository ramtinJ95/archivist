package adrlog

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ddmmyyyyPattern = regexp.MustCompile(`^(Date:\s*)(\d{2})/(\d{2})/(\d{4})\s*$`)

func (r *Repository) UpgradeRepository() (int, error) {
	files, err := r.ListFiles()
	if err != nil {
		return 0, err
	}

	upgraded := 0
	for _, f := range files {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(r.CWD, f)
		}

		changed, err := upgradeDateFormat(absPath)
		if err != nil {
			return upgraded, err
		}
		if changed {
			upgraded++
		}
	}
	return upgraded, nil
}

func upgradeDateFormat(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	changed := false

	for i, line := range lines {
		m := ddmmyyyyPattern.FindStringSubmatch(line)
		if m != nil {
			prefix := m[1]
			dd := m[2]
			mm := m[3]
			yyyy := m[4]
			lines[i] = prefix + yyyy + "-" + mm + "-" + dd
			changed = true
		}
	}

	if changed {
		newContent := strings.Join(lines, "\n")
		return true, atomicWriteFile(path, []byte(newContent))
	}
	return false, nil
}
