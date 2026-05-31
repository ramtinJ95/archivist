package adrlog

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type TOCOptions struct {
	Intro      string
	Outro      string
	LinkPrefix string
}

func (r *Repository) GenerateTOC(opts TOCOptions) (string, error) {
	files, err := r.ListFiles()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("# Architecture Decision Records\n\n")

	if opts.Intro != "" {
		sb.WriteString(opts.Intro)
		sb.WriteString("\n")
	}

	for _, f := range files {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(r.CWD, f)
		}
		rec, err := ParseRecordStrict(absPath)
		if err != nil {
			return "", err
		}

		base := filepath.Base(f)
		link := opts.LinkPrefix + base
		title := fmt.Sprintf("%d. %s", rec.Number, rec.Title)
		sb.WriteString(fmt.Sprintf("* [%s](%s)\n", title, link))
	}

	if opts.Outro != "" {
		sb.WriteString("\n")
		sb.WriteString(opts.Outro)
	}

	return sb.String(), nil
}

type GraphOptions struct {
	LinkPrefix    string
	LinkExtension string
}

var relationPattern = regexp.MustCompile(`^(.+?)\s+\[(\d+)\.\s+.+?\]\((.+?)\)$`)
var dotQuotedStringReplacer = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func (r *Repository) GenerateGraph(opts GraphOptions) (string, error) {
	files, err := r.ListFiles()
	if err != nil {
		return "", err
	}

	ext := opts.LinkExtension
	if ext == "" {
		ext = ".html"
	}

	var records []*Record
	for _, f := range files {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(r.CWD, f)
		}
		rec, err := ParseRecordStrict(absPath)
		if err != nil {
			return "", err
		}
		records = append(records, rec)
	}

	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString("  node [shape=plaintext];\n")
	sb.WriteString("  subgraph {\n")

	for _, rec := range records {
		nodeID := fmt.Sprintf("_%d", rec.Number)
		base := filepath.Base(rec.Path)
		linkTarget := replaceExtension(base, ext)
		url := escapeDOTQuotedString(opts.LinkPrefix + linkTarget)
		label := escapeDOTQuotedString(fmt.Sprintf("%d. %s", rec.Number, rec.Title))

		sb.WriteString(fmt.Sprintf("    %s [label=\"%s\"; URL=\"%s\"];\n", nodeID, label, url))
		if rec.Number > 1 {
			prev := fmt.Sprintf("_%d", rec.Number-1)
			sb.WriteString(fmt.Sprintf("    %s -> %s [style=\"dotted\", weight=1];\n", prev, nodeID))
		}
	}
	sb.WriteString("  }\n")

	for _, rec := range records {
		for _, statusLine := range rec.Status {
			m := relationPattern.FindStringSubmatch(statusLine)
			if m == nil {
				continue
			}
			label := m[1]
			targetBase := m[3]

			if strings.HasSuffix(label, " by") {
				continue
			}

			targetNum := ExtractLeadingNumber(targetBase)
			if targetNum == 0 {
				continue
			}

			srcNode := fmt.Sprintf("_%d", rec.Number)
			tgtNode := fmt.Sprintf("_%d", targetNum)
			sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"%s\", weight=0]\n", srcNode, tgtNode, escapeDOTQuotedString(label)))
		}
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}

func replaceExtension(filename, newExt string) string {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)] + newExt
}

func escapeDOTQuotedString(value string) string {
	return dotQuotedStringReplacer.Replace(value)
}
