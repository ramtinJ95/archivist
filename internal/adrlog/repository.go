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

	repo := &Repository{
		CWD:    absBase,
		ADRDir: dir,
	}

	number, err := repo.NextNumber()
	if err != nil {
		return "", err
	}

	date := os.Getenv("ADR_DATE")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	title := "Record architecture decisions"
	content := ApplyTemplate(InitTemplate, number, title, date, "Accepted")

	filename := repo.GenerateFilename(number, title)
	initFile := filepath.Join(adrPath, filename)
	if err := atomicWriteFile(initFile, []byte(content)); err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}
