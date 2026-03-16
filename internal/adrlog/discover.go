package adrlog

import (
	"os"
	"path/filepath"
	"strings"
)

const dotADRDirFile = ".adr-dir"
const defaultADRDir = "doc/adr"

func DiscoverADRDir(cwd string) (string, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}

	for dir := abs; ; dir = filepath.Dir(dir) {
		dotPath := filepath.Join(dir, dotADRDirFile)
		if data, err := os.ReadFile(dotPath); err == nil {
			content := strings.TrimSpace(string(data))
			if content != "" {
				resolved := filepath.Join(dir, content)
				rel, err := filepath.Rel(abs, resolved)
				if err != nil {
					return resolved, nil
				}
				return rel, nil
			}
		}

		candidate := filepath.Join(dir, defaultADRDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			rel, err := filepath.Rel(abs, candidate)
			if err != nil {
				return candidate, nil
			}
			return rel, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	return defaultADRDir, nil
}

func OpenRepository(cwd string) (*Repository, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}

	adrDir, err := DiscoverADRDir(abs)
	if err != nil {
		return nil, err
	}

	return &Repository{
		CWD:    abs,
		ADRDir: adrDir,
	}, nil
}

func (r *Repository) AbsADRDir() string {
	if filepath.IsAbs(r.ADRDir) {
		return r.ADRDir
	}
	return filepath.Join(r.CWD, r.ADRDir)
}
