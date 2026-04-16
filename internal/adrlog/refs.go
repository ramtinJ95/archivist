package adrlog

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func (r *Repository) ResolveRefCandidates(ref string) ([]string, error) {
	files, err := r.ListFiles()
	if err != nil {
		return nil, err
	}

	if num, err := strconv.Atoi(ref); err == nil {
		padded := fmt.Sprintf("%04d", num)
		var matches []string
		for _, f := range files {
			base := filepath.Base(f)
			if strings.HasPrefix(base, padded+"-") || strings.HasPrefix(base, ref+"-") {
				matches = append(matches, f)
			}
		}
		if len(matches) > 0 {
			return matches, nil
		}

		for _, f := range files {
			base := filepath.Base(f)
			if ExtractLeadingNumber(base) == num {
				matches = append(matches, f)
			}
		}
		return matches, nil
	}

	var matches []string
	for _, f := range files {
		if filepath.Base(f) == ref {
			matches = append(matches, f)
		}
	}
	if len(matches) > 0 {
		return matches, nil
	}

	for _, f := range files {
		if strings.Contains(filepath.Base(f), ref) {
			matches = append(matches, f)
		}
	}

	return matches, nil
}

func (r *Repository) ResolveRef(ref string) (string, error) {
	matches, err := r.ResolveRefCandidates(ref)
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		return matches[0], nil
	}

	if _, err := strconv.Atoi(ref); err == nil {
		return "", fmt.Errorf("ADR %s not found", ref)
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
