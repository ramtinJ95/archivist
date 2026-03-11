package adrlog

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func (r *Repository) ResolveRef(ref string) (string, error) {
	files, err := r.ListFiles()
	if err != nil {
		return "", err
	}

	if num, err := strconv.Atoi(ref); err == nil {
		padded := fmt.Sprintf("%04d", num)
		for _, f := range files {
			base := filepath.Base(f)
			if strings.HasPrefix(base, padded+"-") || strings.HasPrefix(base, ref+"-") {
				return f, nil
			}
		}

		for _, f := range files {
			base := filepath.Base(f)
			n := ExtractLeadingNumber(base)
			if n == num {
				return f, nil
			}
		}
		return "", fmt.Errorf("ADR %s not found", ref)
	}

	for _, f := range files {
		base := filepath.Base(f)
		if base == ref {
			return f, nil
		}
	}

	for _, f := range files {
		base := filepath.Base(f)
		if strings.Contains(base, ref) {
			return f, nil
		}
	}

	return "", fmt.Errorf("ADR %q not found", ref)
}

func (r *Repository) ResolveRecord(ref string) (*Record, error) {
	path, err := r.ResolveRef(ref)
	if err != nil {
		return nil, err
	}

	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(r.CWD, path)
	}
	return ParseRecord(absPath)
}
