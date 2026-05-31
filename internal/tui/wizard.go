package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ramtinJ95/archivist/internal/adrlog"
)

type wizardKind int

const (
	wizardCreate wizardKind = iota
	wizardSupersede
	wizardLink
	wizardGenerateTOC
	wizardGenerateGraph
)

type wizardResult struct {
	message       string
	selectPath    string
	editPath      string
	reloadRecords bool
}

type wizardModel struct {
	repo              *adrlog.Repository
	kind              wizardKind
	inputs            []textinput.Model
	labels            []string
	focusIndex        int
	done              bool
	cancelled         bool
	confirming        bool
	subjectRecord     *adrlog.Record
	records           []*adrlog.Record
	linkMatches       []*adrlog.Record
	selectedLinkMatch int
}

func newCreateWizard(repo *adrlog.Repository) wizardModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "e.g. Use PostgreSQL for persistence"
	titleInput.CharLimit = 120
	titleInput.Width = 50
	titleInput.Focus()

	supersedesInput := textinput.New()
	supersedesInput.Placeholder = "optional, e.g. 2 or 1,3"
	supersedesInput.CharLimit = 120
	supersedesInput.Width = 50

	linksInput := textinput.New()
	linksInput.Placeholder = "optional, e.g. 1:Clarifies:Clarified by"
	linksInput.CharLimit = 240
	linksInput.Width = 70

	return wizardModel{
		repo:   repo,
		kind:   wizardCreate,
		inputs: []textinput.Model{titleInput, supersedesInput, linksInput},
		labels: []string{"Title", "Supersedes", "Links"},
	}
}

func newSupersedeWizard(repo *adrlog.Repository, target *adrlog.Record) wizardModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Title for the new ADR"
	titleInput.CharLimit = 120
	titleInput.Width = 50
	titleInput.Focus()

	return wizardModel{
		repo:          repo,
		kind:          wizardSupersede,
		inputs:        []textinput.Model{titleInput},
		labels:        []string{"New ADR title"},
		subjectRecord: target,
	}
}

func newLinkWizard(repo *adrlog.Repository, source *adrlog.Record, records []*adrlog.Record) wizardModel {
	targetInput := textinput.New()
	targetInput.Placeholder = "Search by ADR number, title, or filename"
	targetInput.CharLimit = 120
	targetInput.Width = 50
	targetInput.Focus()

	fwdInput := textinput.New()
	fwdInput.Placeholder = "e.g. Amends"
	fwdInput.CharLimit = 60
	fwdInput.Width = 50

	revInput := textinput.New()
	revInput.Placeholder = "e.g. Amended by"
	revInput.CharLimit = 60
	revInput.Width = 50

	cloned := append([]*adrlog.Record(nil), records...)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Number < cloned[j].Number
	})

	w := wizardModel{
		repo:          repo,
		kind:          wizardLink,
		inputs:        []textinput.Model{targetInput, fwdInput, revInput},
		labels:        []string{"Target ADR", "Forward label", "Reverse label"},
		subjectRecord: source,
		records:       cloned,
	}
	w.refreshLinkMatches()
	return w
}

func newGenerateTOCWizard(repo *adrlog.Repository) wizardModel {
	outputInput := textinput.New()
	outputInput.Placeholder = defaultTOCOutputPath(repo)
	outputInput.CharLimit = 240
	outputInput.Width = 60
	outputInput.Focus()

	introInput := textinput.New()
	introInput.Placeholder = "optional intro Markdown file"
	introInput.CharLimit = 240
	introInput.Width = 60

	outroInput := textinput.New()
	outroInput.Placeholder = "optional outro Markdown file"
	outroInput.CharLimit = 240
	outroInput.Width = 60

	prefixInput := textinput.New()
	prefixInput.Placeholder = "optional link prefix"
	prefixInput.CharLimit = 240
	prefixInput.Width = 60

	return wizardModel{
		repo:   repo,
		kind:   wizardGenerateTOC,
		inputs: []textinput.Model{outputInput, introInput, outroInput, prefixInput},
		labels: []string{"Output path", "Intro file", "Outro file", "Link prefix"},
	}
}

func newGenerateGraphWizard(repo *adrlog.Repository) wizardModel {
	outputInput := textinput.New()
	outputInput.Placeholder = defaultGraphOutputPath(repo)
	outputInput.CharLimit = 240
	outputInput.Width = 60
	outputInput.Focus()

	prefixInput := textinput.New()
	prefixInput.Placeholder = "optional URL/link prefix"
	prefixInput.CharLimit = 240
	prefixInput.Width = 60

	extInput := textinput.New()
	extInput.Placeholder = "optional link extension, default .html"
	extInput.CharLimit = 60
	extInput.Width = 30

	return wizardModel{
		repo:   repo,
		kind:   wizardGenerateGraph,
		inputs: []textinput.Model{outputInput, prefixInput, extInput},
		labels: []string{"Output path", "Link prefix", "Link extension"},
	}
}

func (w *wizardModel) update(msg tea.KeyMsg) tea.Cmd {
	if w.confirming {
		switch msg.String() {
		case "ctrl+c":
			w.cancelled = true
			w.done = true
		case "esc":
			w.confirming = false
		case "enter":
			w.done = true
		}
		return nil
	}

	switch msg.String() {
	case "ctrl+c":
		w.cancelled = true
		w.done = true
		return nil
	case "esc":
		w.cancelled = true
		w.done = true
		return nil
	case "tab":
		w.commitFocusedSelection()
		w.focusIndex = (w.focusIndex + 1) % len(w.inputs)
		return w.updateFocus()
	case "shift+tab":
		w.commitFocusedSelection()
		w.focusIndex = (w.focusIndex - 1 + len(w.inputs)) % len(w.inputs)
		return w.updateFocus()
	case "down":
		if w.canNavigateLinkMatches() {
			w.moveLinkMatch(1)
			return nil
		}
		w.focusIndex = (w.focusIndex + 1) % len(w.inputs)
		return w.updateFocus()
	case "up":
		if w.canNavigateLinkMatches() {
			w.moveLinkMatch(-1)
			return nil
		}
		w.focusIndex = (w.focusIndex - 1 + len(w.inputs)) % len(w.inputs)
		return w.updateFocus()
	case "enter":
		w.commitFocusedSelection()
		if w.focusIndex == len(w.inputs)-1 {
			w.confirming = true
			return nil
		}
		w.focusIndex++
		return w.updateFocus()
	}

	var cmd tea.Cmd
	w.inputs[w.focusIndex], cmd = w.inputs[w.focusIndex].Update(msg)
	if w.kind == wizardLink && w.focusIndex == 0 {
		w.refreshLinkMatches()
	}
	return cmd
}

func (w *wizardModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(w.inputs))
	for i := range w.inputs {
		if i == w.focusIndex {
			cmds[i] = w.inputs[i].Focus()
		} else {
			w.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (w *wizardModel) canNavigateLinkMatches() bool {
	return w.kind == wizardLink && !w.confirming && w.focusIndex == 0 && len(w.linkMatches) > 0
}

func (w *wizardModel) moveLinkMatch(delta int) {
	if len(w.linkMatches) == 0 {
		return
	}
	w.selectedLinkMatch = (w.selectedLinkMatch + delta + len(w.linkMatches)) % len(w.linkMatches)
}

func (w *wizardModel) commitFocusedSelection() {
	if w.kind != wizardLink || w.focusIndex != 0 {
		return
	}
	selected := w.selectedLinkRecord()
	if selected == nil {
		return
	}
	w.inputs[0].SetValue(fmt.Sprintf("%d", selected.Number))
	w.refreshLinkMatches()
}

func (w *wizardModel) refreshLinkMatches() {
	if w.kind != wizardLink {
		return
	}

	query := strings.TrimSpace(strings.ToLower(w.inputs[0].Value()))
	matches := make([]*adrlog.Record, 0, len(w.records))
	for _, rec := range w.records {
		if w.subjectRecord != nil && rec.Number == w.subjectRecord.Number {
			continue
		}
		if query == "" || linkRecordMatchesQuery(rec, query) {
			matches = append(matches, rec)
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		iScore := linkRecordMatchScore(matches[i], query)
		jScore := linkRecordMatchScore(matches[j], query)
		if iScore != jScore {
			return iScore < jScore
		}
		return matches[i].Number < matches[j].Number
	})

	w.linkMatches = matches
	if len(matches) == 0 {
		w.selectedLinkMatch = 0
		return
	}
	if w.selectedLinkMatch >= len(matches) {
		w.selectedLinkMatch = 0
	}
}

func linkRecordMatchesQuery(rec *adrlog.Record, query string) bool {
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(fmt.Sprintf("%d", rec.Number)), query) ||
		strings.Contains(strings.ToLower(rec.Title), query) ||
		strings.Contains(strings.ToLower(filepath.Base(rec.Path)), query)
}

func linkRecordMatchScore(rec *adrlog.Record, query string) int {
	if query == "" {
		return 0
	}
	number := fmt.Sprintf("%d", rec.Number)
	base := strings.ToLower(filepath.Base(rec.Path))
	title := strings.ToLower(rec.Title)

	switch {
	case number == query:
		return 0
	case strings.HasPrefix(base, query):
		return 1
	case strings.HasPrefix(title, query):
		return 2
	case strings.Contains(base, query):
		return 3
	case strings.Contains(title, query):
		return 4
	default:
		return 5
	}
}

func (w *wizardModel) selectedLinkRecord() *adrlog.Record {
	if len(w.linkMatches) == 0 || w.selectedLinkMatch >= len(w.linkMatches) {
		return nil
	}
	return w.linkMatches[w.selectedLinkMatch]
}

func (w *wizardModel) confirmationSummary() string {
	switch w.kind {
	case wizardCreate:
		title := strings.TrimSpace(w.inputs[0].Value())
		return fmt.Sprintf("Create new ADR: %q", title)
	case wizardSupersede:
		title := strings.TrimSpace(w.inputs[0].Value())
		return fmt.Sprintf("Create %q superseding ADR %d: %s",
			title, w.subjectRecord.Number, w.subjectRecord.Title)
	case wizardLink:
		target := strings.TrimSpace(w.inputs[0].Value())
		fwd := strings.TrimSpace(w.inputs[1].Value())
		rev := strings.TrimSpace(w.inputs[2].Value())
		return fmt.Sprintf("Link ADR %d -%s-> %s (reverse: %s)",
			w.subjectRecord.Number, fwd, target, rev)
	case wizardGenerateTOC:
		return fmt.Sprintf("Export TOC to %q", strings.TrimSpace(w.inputs[0].Value()))
	case wizardGenerateGraph:
		return fmt.Sprintf("Export DOT graph to %q", strings.TrimSpace(w.inputs[0].Value()))
	}
	return ""
}

func (w *wizardModel) view(width, height int) string {
	if w.confirming {
		return w.renderConfirmation(width, height)
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(w.title()))
	sb.WriteString("\n\n")

	for i, input := range w.inputs {
		label := w.labels[i]
		if i == w.focusIndex {
			sb.WriteString(helpKeyStyle.Render(label))
		} else {
			sb.WriteString(helpDescStyle.Render(label))
		}
		sb.WriteString("\n")
		sb.WriteString(input.View())
		sb.WriteString("\n\n")
	}

	if w.kind == wizardLink {
		sb.WriteString(titleStyle.Render("Target matches"))
		sb.WriteString("\n")
		sb.WriteString(w.renderLinkMatches())
		sb.WriteString("\n")
	}

	sb.WriteString(titleStyle.Render("Preview"))
	sb.WriteString("\n")
	sb.WriteString(helpDescStyle.Render(w.previewText()))
	sb.WriteString("\n\n")
	sb.WriteString(helpDescStyle.Render(w.instructions()))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, sb.String())
}

func (w *wizardModel) renderConfirmation(width, height int) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Confirm"))
	sb.WriteString("\n\n")
	sb.WriteString(w.confirmationSummary())
	sb.WriteString("\n\n")
	sb.WriteString(titleStyle.Render("Preview"))
	sb.WriteString("\n")
	sb.WriteString(helpDescStyle.Render(w.previewText()))
	sb.WriteString("\n\n")
	sb.WriteString(helpDescStyle.Render("enter: confirm  esc: back to editing"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, sb.String())
}

func (w *wizardModel) title() string {
	switch w.kind {
	case wizardCreate:
		return "Create New ADR"
	case wizardSupersede:
		return fmt.Sprintf("Supersede ADR %d: %s", w.subjectRecord.Number, w.subjectRecord.Title)
	case wizardLink:
		return fmt.Sprintf("Link from ADR %d: %s", w.subjectRecord.Number, w.subjectRecord.Title)
	case wizardGenerateTOC:
		return "Export Table of Contents"
	case wizardGenerateGraph:
		return "Export DOT Graph"
	default:
		return "Wizard"
	}
}

func (w *wizardModel) instructions() string {
	if w.kind == wizardLink {
		return "tab/shift+tab: navigate fields  up/down: move between fields or target matches  enter: submit  esc: cancel"
	}
	return "tab/shift+tab: navigate fields  enter: submit  esc: cancel"
}

func (w *wizardModel) renderLinkMatches() string {
	if len(w.linkMatches) == 0 {
		return "  No ADR matches the current query."
	}

	var sb strings.Builder
	limit := len(w.linkMatches)
	if limit > 6 {
		limit = 6
	}

	for i := 0; i < limit; i++ {
		rec := w.linkMatches[i]
		marker := " "
		style := helpDescStyle
		if i == w.selectedLinkMatch {
			marker = ">"
			style = helpKeyStyle
		}
		line := fmt.Sprintf("%s %d. %s (%s)", marker, rec.Number, rec.Title, filepath.Base(rec.Path))
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}

	if len(w.linkMatches) > limit {
		sb.WriteString(helpDescStyle.Render(fmt.Sprintf("… %d more match(es)", len(w.linkMatches)-limit)))
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

func (w *wizardModel) previewText() string {
	switch w.kind {
	case wizardCreate:
		return w.previewCreateText()
	case wizardSupersede:
		return w.previewSupersedeText()
	case wizardLink:
		return w.previewLinkText()
	case wizardGenerateTOC:
		return w.previewGenerateTOCText()
	case wizardGenerateGraph:
		return w.previewGenerateGraphText()
	default:
		return ""
	}
}

func (w *wizardModel) previewCreateText() string {
	title := strings.TrimSpace(w.inputs[0].Value())
	if title == "" {
		return "Enter a title to preview the new ADR path and initial status."
	}

	number, relPath, err := w.previewNewRecordPath(title)
	if err != nil {
		return "Unable to preview new ADR: " + err.Error()
	}

	lines := []string{
		fmt.Sprintf("Path: %s", relPath),
		fmt.Sprintf("Number: %d", number),
		fmt.Sprintf("Title: %s", title),
		"",
		"Changes:",
		fmt.Sprintf("  + %s", relPath),
		fmt.Sprintf("    + # %d. %s", number, title),
		"    + Accepted",
	}

	supersedes := parseCommaList(w.inputs[1].Value())
	for _, ref := range supersedes {
		if rec, err := w.repo.ResolveRecord(ref); err == nil {
			lines = append(lines,
				fmt.Sprintf("    + Supercedes [%d. %s](%s)", rec.Number, rec.Title, filepath.Base(rec.Path)),
				fmt.Sprintf("  ~ %s", w.displayPath(rec.Path)),
				"    - Accepted",
				fmt.Sprintf("    + Superceded by [%d. %s](%s)", number, title, filepath.Base(relPath)),
			)
		} else {
			lines = append(lines, fmt.Sprintf("    ! supersede %q: %v", ref, err))
		}
	}

	links, linkErrors := parseCreateWizardLinks(w.inputs[2].Value())
	for _, err := range linkErrors {
		lines = append(lines, fmt.Sprintf("    ! link spec: %v", err))
	}
	for _, link := range links {
		if rec, err := w.repo.ResolveRecord(link.Target); err == nil {
			lines = append(lines,
				fmt.Sprintf("    + %s [%d. %s](%s)", link.ForwardLabel, rec.Number, rec.Title, filepath.Base(rec.Path)),
				fmt.Sprintf("  ~ %s", w.displayPath(rec.Path)),
				fmt.Sprintf("    + %s [%d. %s](%s)", link.ReverseLabel, number, title, filepath.Base(relPath)),
			)
		} else {
			lines = append(lines, fmt.Sprintf("    ! link target %q: %v", link.Target, err))
		}
	}

	return strings.Join(lines, "\n")
}

func (w *wizardModel) previewSupersedeText() string {
	title := strings.TrimSpace(w.inputs[0].Value())
	if title == "" {
		return "Enter a title to preview the new ADR path and supersede mutations."
	}

	number, relPath, err := w.previewNewRecordPath(title)
	if err != nil {
		return "Unable to preview supersede flow: " + err.Error()
	}

	targetPath := w.displayPath(w.subjectRecord.Path)
	targetBase := filepath.Base(w.subjectRecord.Path)
	newBase := filepath.Base(relPath)

	return strings.Join([]string{
		fmt.Sprintf("Path: %s", relPath),
		fmt.Sprintf("Number: %d", number),
		fmt.Sprintf("Supersedes: %d. %s", w.subjectRecord.Number, w.subjectRecord.Title),
		"",
		"Changes:",
		fmt.Sprintf("  + %s", relPath),
		fmt.Sprintf("    + # %d. %s", number, title),
		"    + Accepted",
		fmt.Sprintf("    + Supercedes [%d. %s](%s)", w.subjectRecord.Number, w.subjectRecord.Title, targetBase),
		fmt.Sprintf("  ~ %s", targetPath),
		"    - Accepted",
		fmt.Sprintf("    + Superceded by [%d. %s](%s)", number, title, newBase),
	}, "\n")
}

func (w *wizardModel) previewLinkText() string {
	if w.subjectRecord == nil {
		return "Select a source ADR to preview link changes."
	}

	target, err := w.linkTargetRecord()
	if err != nil {
		return "Resolve target ADR to preview link changes: " + err.Error()
	}
	if target == nil {
		return "Search for a target ADR to preview reciprocal link mutations."
	}

	fwdLabel := strings.TrimSpace(w.inputs[1].Value())
	if fwdLabel == "" {
		fwdLabel = "<forward label>"
	}
	revLabel := strings.TrimSpace(w.inputs[2].Value())
	if revLabel == "" {
		revLabel = "<reverse label>"
	}

	sourcePath := w.displayPath(w.subjectRecord.Path)
	targetPath := w.displayPath(target.Path)

	return strings.Join([]string{
		fmt.Sprintf("Source: %d. %s", w.subjectRecord.Number, w.subjectRecord.Title),
		fmt.Sprintf("Target: %d. %s", target.Number, target.Title),
		"",
		"Changes:",
		fmt.Sprintf("  ~ %s", sourcePath),
		fmt.Sprintf("    + %s [%d. %s](%s)", fwdLabel, target.Number, target.Title, filepath.Base(target.Path)),
		fmt.Sprintf("  ~ %s", targetPath),
		fmt.Sprintf("    + %s [%d. %s](%s)", revLabel, w.subjectRecord.Number, w.subjectRecord.Title, filepath.Base(w.subjectRecord.Path)),
	}, "\n")
}

func (w *wizardModel) previewGenerateTOCText() string {
	output := strings.TrimSpace(w.inputs[0].Value())
	if output == "" {
		output = defaultTOCOutputPath(w.repo)
	}

	toc, err := w.generateTOCContent()
	if err != nil {
		return "Unable to generate TOC preview: " + err.Error()
	}
	return previewGeneratedContent(w.repo, "TOC", output, toc)
}

func (w *wizardModel) previewGenerateGraphText() string {
	output := strings.TrimSpace(w.inputs[0].Value())
	if output == "" {
		output = defaultGraphOutputPath(w.repo)
	}

	graph, err := w.repo.GenerateGraph(adrlog.GraphOptions{
		LinkPrefix:    strings.TrimSpace(w.inputs[1].Value()),
		LinkExtension: strings.TrimSpace(w.inputs[2].Value()),
	})
	if err != nil {
		return "Unable to generate graph preview: " + err.Error()
	}
	return previewGeneratedContent(w.repo, "DOT graph", output, graph)
}

func previewGeneratedContent(repo *adrlog.Repository, kind, output, content string) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) > 12 {
		lines = append(lines[:12], fmt.Sprintf("… %d more line(s)", len(lines)-12))
	}

	action := "Action: create new file"
	if absPath, err := repoRelativePath(repo, output); err != nil {
		action = "Invalid output path: " + err.Error()
	} else if _, err := os.Stat(absPath); err == nil {
		action = "Warning: will overwrite existing file"
	} else if !os.IsNotExist(err) {
		action = "Warning: cannot inspect output path: " + err.Error()
	}

	return strings.Join(append([]string{
		fmt.Sprintf("Output: %s", output),
		fmt.Sprintf("Kind: %s", kind),
		action,
		"",
		"Preview:",
	}, lines...), "\n")
}

func (w *wizardModel) previewNewRecordPath(title string) (int, string, error) {
	if w.repo == nil {
		return 0, "", fmt.Errorf("repository not loaded")
	}
	number, err := w.repo.NextNumber()
	if err != nil {
		return 0, "", err
	}
	filename := w.repo.GenerateFilename(number, title)
	return number, filepath.Join(w.repo.ADRDir, filename), nil
}

func (w *wizardModel) displayPath(path string) string {
	if w.repo == nil || !filepath.IsAbs(path) {
		return path
	}
	rel, err := filepath.Rel(w.repo.CWD, path)
	if err != nil {
		return path
	}
	return rel
}

func (w *wizardModel) linkTargetRecord() (*adrlog.Record, error) {
	if selected := w.selectedLinkRecord(); selected != nil {
		return selected, nil
	}

	query := strings.TrimSpace(w.inputs[0].Value())
	if query == "" {
		return nil, nil
	}

	path, err := w.repo.ResolveRef(query)
	if err != nil {
		return nil, err
	}

	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(w.repo.CWD, path)
	}
	return adrlog.ParseRecordStrict(absPath)
}

func (w *wizardModel) execute(repo *adrlog.Repository) (wizardResult, error) {
	switch w.kind {
	case wizardCreate:
		return w.executeCreate(repo)
	case wizardSupersede:
		return w.executeSupersede(repo)
	case wizardLink:
		return w.executeLink(repo)
	case wizardGenerateTOC:
		return w.executeGenerateTOC(repo)
	case wizardGenerateGraph:
		return w.executeGenerateGraph(repo)
	}
	return wizardResult{}, nil
}

func (w *wizardModel) executeCreate(repo *adrlog.Repository) (wizardResult, error) {
	title := strings.TrimSpace(w.inputs[0].Value())
	if title == "" {
		return wizardResult{}, fmt.Errorf("title is required")
	}

	links, linkErrors := parseCreateWizardLinks(w.inputs[2].Value())
	if len(linkErrors) > 0 {
		return wizardResult{}, linkErrors[0]
	}

	path, err := repo.CreateADR(adrlog.CreateOptions{
		Title:      title,
		Supersedes: parseCommaList(w.inputs[1].Value()),
		Links:      links,
	})
	if err != nil {
		return wizardResult{}, err
	}

	return wizardResult{
		message:       fmt.Sprintf("Created %s", path),
		selectPath:    path,
		editPath:      path,
		reloadRecords: true,
	}, nil
}

func (w *wizardModel) executeSupersede(repo *adrlog.Repository) (wizardResult, error) {
	title := strings.TrimSpace(w.inputs[0].Value())
	if title == "" {
		return wizardResult{}, fmt.Errorf("title is required")
	}

	targetRef := fmt.Sprintf("%d", w.subjectRecord.Number)
	path, err := repo.CreateADR(adrlog.CreateOptions{
		Title:      title,
		Supersedes: []string{targetRef},
	})
	if err != nil {
		return wizardResult{}, err
	}

	return wizardResult{
		message:       fmt.Sprintf("Created %s (supersedes ADR %d)", path, w.subjectRecord.Number),
		selectPath:    path,
		editPath:      path,
		reloadRecords: true,
	}, nil
}

func (w *wizardModel) executeLink(repo *adrlog.Repository) (wizardResult, error) {
	targetRef := strings.TrimSpace(w.inputs[0].Value())
	fwdLabel := strings.TrimSpace(w.inputs[1].Value())
	revLabel := strings.TrimSpace(w.inputs[2].Value())

	if targetRef == "" || fwdLabel == "" || revLabel == "" {
		return wizardResult{}, fmt.Errorf("all fields are required")
	}

	targetRecord, err := w.linkTargetRecord()
	if err != nil {
		return wizardResult{}, err
	}
	if targetRecord == nil {
		return wizardResult{}, fmt.Errorf("target ADR is required")
	}

	absTargetPath := targetRecord.Path
	if !filepath.IsAbs(absTargetPath) {
		absTargetPath = filepath.Join(repo.CWD, absTargetPath)
	}

	absSourcePath := w.subjectRecord.Path
	if !filepath.IsAbs(absSourcePath) {
		absSourcePath = filepath.Join(repo.CWD, absSourcePath)
	}

	if err := adrlog.AddLink(absSourcePath, absTargetPath, fwdLabel, revLabel); err != nil {
		return wizardResult{}, err
	}

	return wizardResult{message: fmt.Sprintf("Linked ADR %d -> %d", w.subjectRecord.Number, targetRecord.Number), reloadRecords: true}, nil
}

func (w *wizardModel) executeGenerateTOC(repo *adrlog.Repository) (wizardResult, error) {
	output := strings.TrimSpace(w.inputs[0].Value())
	if output == "" {
		output = defaultTOCOutputPath(repo)
	}
	content, err := w.generateTOCContent()
	if err != nil {
		return wizardResult{}, err
	}
	if err := writeRepoRelativeFile(repo, output, []byte(content)); err != nil {
		return wizardResult{}, err
	}
	return wizardResult{message: fmt.Sprintf("Exported TOC to %s", output)}, nil
}

func (w *wizardModel) executeGenerateGraph(repo *adrlog.Repository) (wizardResult, error) {
	output := strings.TrimSpace(w.inputs[0].Value())
	if output == "" {
		output = defaultGraphOutputPath(repo)
	}
	content, err := repo.GenerateGraph(adrlog.GraphOptions{
		LinkPrefix:    strings.TrimSpace(w.inputs[1].Value()),
		LinkExtension: strings.TrimSpace(w.inputs[2].Value()),
	})
	if err != nil {
		return wizardResult{}, err
	}
	if err := writeRepoRelativeFile(repo, output, []byte(content)); err != nil {
		return wizardResult{}, err
	}
	return wizardResult{message: fmt.Sprintf("Exported DOT graph to %s", output)}, nil
}

func (w *wizardModel) generateTOCContent() (string, error) {
	intro, err := readOptionalRepoFile(w.repo, strings.TrimSpace(w.inputs[1].Value()))
	if err != nil {
		return "", err
	}
	outro, err := readOptionalRepoFile(w.repo, strings.TrimSpace(w.inputs[2].Value()))
	if err != nil {
		return "", err
	}
	return w.repo.GenerateTOC(adrlog.TOCOptions{
		Intro:      intro,
		Outro:      outro,
		LinkPrefix: strings.TrimSpace(w.inputs[3].Value()),
	})
}

func parseCommaList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func parseCreateWizardLinks(value string) ([]adrlog.LinkSpec, []error) {
	parts := parseCommaList(value)
	links := make([]adrlog.LinkSpec, 0, len(parts))
	errs := make([]error, 0)
	for _, part := range parts {
		link, err := adrlog.ParseLinkSpec(part)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		links = append(links, link)
	}
	return links, errs
}

func readOptionalRepoFile(repo *adrlog.Repository, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	absPath, err := repoRelativePath(repo, path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeRepoRelativeFile(repo *adrlog.Repository, path string, data []byte) error {
	absPath, err := repoRelativePath(repo, path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	return atomicWriteFile(absPath, data)
}

func defaultTOCOutputPath(repo *adrlog.Repository) string {
	return filepath.Join(repo.ADRDir, "README.md")
}

func defaultGraphOutputPath(repo *adrlog.Repository) string {
	return filepath.Join(repo.ADRDir, "graph.dot")
}

func repoRelativePath(repo *adrlog.Repository, path string) (string, error) {
	if repo == nil {
		return "", fmt.Errorf("repository not loaded")
	}
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", path)
	}

	base := filepath.Clean(repo.CWD)
	candidate := filepath.Join(base, filepath.Clean(path))
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes repository: %s", path)
	}
	return candidate, nil
}

func atomicWriteFile(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".archivist-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
