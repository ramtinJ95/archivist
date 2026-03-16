package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(title string) string {
	s := strings.ToLower(title)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func (r *Repository) NextNumber() (int, error) {
	files, err := r.ListFiles()
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 1, nil
	}

	maxNum := 0
	for _, f := range files {
		base := filepath.Base(f)
		n := ExtractLeadingNumber(base)
		if n > maxNum {
			maxNum = n
		}
	}
	return maxNum + 1, nil
}

func (r *Repository) GenerateFilename(number int, title string) string {
	return fmt.Sprintf("%s-%s.md", PadNumber(number), Slugify(title))
}

type CreateOptions struct {
	Title      string
	Supersedes []string
	Links      []LinkSpec
	Date       string
}

type LinkSpec struct {
	Target       string
	ForwardLabel string
	ReverseLabel string
}

func ParseLinkSpec(spec string) (LinkSpec, error) {
	parts := strings.SplitN(spec, ":", 3)
	if len(parts) != 3 {
		return LinkSpec{}, fmt.Errorf("invalid link spec %q: expected TARGET:LINK:REVERSE-LINK", spec)
	}
	return LinkSpec{
		Target:       parts[0],
		ForwardLabel: parts[1],
		ReverseLabel: parts[2],
	}, nil
}

func (r *Repository) CreateADR(opts CreateOptions) (string, error) {
	absDir := r.AbsADRDir()
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", err
	}

	number, err := r.NextNumber()
	if err != nil {
		return "", err
	}

	date := opts.Date
	if date == "" {
		if envDate := os.Getenv("ADR_DATE"); envDate != "" {
			date = envDate
		} else {
			date = time.Now().Format("2006-01-02")
		}
	}

	template, err := ResolveTemplate(absDir)
	if err != nil {
		return "", err
	}
	content := ApplyTemplate(template, number, opts.Title, date, "Accepted")

	filename := r.GenerateFilename(number, opts.Title)
	fullPath := filepath.Join(absDir, filename)

	if err := atomicWriteFile(fullPath, []byte(content)); err != nil {
		return "", err
	}

	relPath := filepath.Join(r.ADRDir, filename)

	for _, sup := range opts.Supersedes {
		supPath, err := r.ResolveRef(sup)
		if err != nil {
			return relPath, fmt.Errorf("resolving supersede target: %w", err)
		}

		absSupPath := supPath
		if !filepath.IsAbs(supPath) {
			absSupPath = filepath.Join(r.CWD, supPath)
		}

		supRec, err := ParseRecord(absSupPath)
		if err != nil {
			return relPath, err
		}

		newRec, err := ParseRecord(fullPath)
		if err != nil {
			return relPath, err
		}

		supLink := fmt.Sprintf("Superceded by [%d. %s](%s)", newRec.Number, newRec.Title, filename)
		if err := addStatusLine(absSupPath, supLink); err != nil {
			return relPath, err
		}
		if err := removeStatusLine(absSupPath, "Accepted"); err != nil {
			return relPath, err
		}

		newLink := fmt.Sprintf("Supercedes [%d. %s](%s)", supRec.Number, supRec.Title, filepath.Base(supPath))
		if err := addStatusLine(fullPath, newLink); err != nil {
			return relPath, err
		}
	}

	for _, link := range opts.Links {
		targetPath, err := r.ResolveRef(link.Target)
		if err != nil {
			return relPath, fmt.Errorf("resolving link target: %w", err)
		}

		absTargetPath := targetPath
		if !filepath.IsAbs(targetPath) {
			absTargetPath = filepath.Join(r.CWD, targetPath)
		}

		targetRec, err := ParseRecord(absTargetPath)
		if err != nil {
			return relPath, err
		}

		newRec, err := ParseRecord(fullPath)
		if err != nil {
			return relPath, err
		}

		fwdLine := fmt.Sprintf("%s [%d. %s](%s)", link.ForwardLabel, targetRec.Number, targetRec.Title, filepath.Base(targetPath))
		if err := addStatusLine(fullPath, fwdLine); err != nil {
			return relPath, err
		}

		revLine := fmt.Sprintf("%s [%d. %s](%s)", link.ReverseLabel, newRec.Number, newRec.Title, filename)
		if err := addStatusLine(absTargetPath, revLine); err != nil {
			return relPath, err
		}
	}

	return relPath, nil
}

func atomicWriteFile(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
