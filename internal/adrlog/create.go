package adrlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	absDir := r.AbsADRDir()

	info, err := os.Stat(absDir)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("%s is not a directory", r.ADRDir)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return 0, err
	}

	maxNum := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		n := ExtractLeadingNumber(entry.Name())
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
	relPath := filepath.Join(r.ADRDir, filename)
	newContent := content

	type plannedExistingFile struct {
		original string
		updated  string
	}

	plannedExistingFiles := make(map[string]*plannedExistingFile)

	loadPlannedExistingFile := func(path string) (*plannedExistingFile, error) {
		if plannedFile, ok := plannedExistingFiles[path]; ok {
			return plannedFile, nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		plannedFile := &plannedExistingFile{
			original: string(data),
			updated:  string(data),
		}
		plannedExistingFiles[path] = plannedFile
		return plannedFile, nil
	}

	updatePlannedExistingFile := func(path string, update func(string) (string, error)) error {
		plannedFile, err := loadPlannedExistingFile(path)
		if err != nil {
			return err
		}

		updatedContent, err := update(plannedFile.updated)
		if err != nil {
			return err
		}
		plannedFile.updated = updatedContent
		return nil
	}

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

		supLink := fmt.Sprintf("Superceded by [%d. %s](%s)", number, opts.Title, filename)
		if err := updatePlannedExistingFile(absSupPath, func(current string) (string, error) {
			return addStatusLineContent(current, supLink)
		}); err != nil {
			return relPath, err
		}
		if err := updatePlannedExistingFile(absSupPath, func(current string) (string, error) {
			return removeStatusLineContent(current, "Accepted")
		}); err != nil {
			return relPath, err
		}

		newLink := fmt.Sprintf("Supercedes [%d. %s](%s)", supRec.Number, supRec.Title, filepath.Base(supPath))
		newContent, err = addStatusLineContent(newContent, newLink)
		if err != nil {
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

		fwdLine := fmt.Sprintf("%s [%d. %s](%s)", link.ForwardLabel, targetRec.Number, targetRec.Title, filepath.Base(targetPath))
		newContent, err = addStatusLineContent(newContent, fwdLine)
		if err != nil {
			return relPath, err
		}

		revLine := fmt.Sprintf("%s [%d. %s](%s)", link.ReverseLabel, number, opts.Title, filename)
		if err := updatePlannedExistingFile(absTargetPath, func(current string) (string, error) {
			return addStatusLineContent(current, revLine)
		}); err != nil {
			return relPath, err
		}
	}

	if err := atomicWriteFile(fullPath, []byte(newContent)); err != nil {
		return "", err
	}

	plannedPaths := make([]string, 0, len(plannedExistingFiles))
	for path := range plannedExistingFiles {
		plannedPaths = append(plannedPaths, path)
	}
	sort.Strings(plannedPaths)

	appliedPaths := make([]string, 0, len(plannedPaths))
	for _, path := range plannedPaths {
		plannedFile := plannedExistingFiles[path]
		if err := atomicWriteFile(path, []byte(plannedFile.updated)); err != nil {
			rollbackErr := os.Remove(fullPath)
			for i := len(appliedPaths) - 1; i >= 0; i-- {
				appliedPath := appliedPaths[i]
				appliedFile := plannedExistingFiles[appliedPath]
				if restoreErr := atomicWriteFile(appliedPath, []byte(appliedFile.original)); rollbackErr == nil && restoreErr != nil {
					rollbackErr = restoreErr
				}
			}
			if rollbackErr != nil {
				return relPath, fmt.Errorf("%w (rollback failed: %v)", err, rollbackErr)
			}
			return relPath, err
		}
		appliedPaths = append(appliedPaths, path)
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
