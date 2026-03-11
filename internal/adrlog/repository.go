package adrlog

import (
	"os"
	"path/filepath"
	"time"
)

func InitRepository(cwd, dir string) (string, error) {
	if dir == "" {
		dir = defaultADRDir
	}

	absBase, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}

	adrPath := filepath.Join(absBase, dir)
	if err := os.MkdirAll(adrPath, 0o755); err != nil {
		return "", err
	}

	if dir != defaultADRDir {
		dotFile := filepath.Join(absBase, dotADRDirFile)
		if err := os.WriteFile(dotFile, []byte(dir), 0o644); err != nil {
			return "", err
		}
	}

	date := os.Getenv("ADR_DATE")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	content := ApplyTemplate(InitTemplate, 1, "Record architecture decisions", date, "Accepted")

	initFile := filepath.Join(adrPath, "0001-record-architecture-decisions.md")
	if err := atomicWriteFile(initFile, []byte(content)); err != nil {
		return "", err
	}

	return filepath.Join(dir, "0001-record-architecture-decisions.md"), nil
}
