package tui

import (
	"fmt"
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
)

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

	return wizardModel{
		repo:   repo,
		kind:   wizardCreate,
		inputs: []textinput.Model{titleInput},
		labels: []string{"Title"},
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

func (w *wizardModel) update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		w.cancelled = true
		w.done = true
		return nil
	case "esc":
		if w.confirming {
			w.confirming = false
			return nil
		}
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
		if w.confirming {
			w.done = true
			return nil
		}
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

	return strings.Join([]string{
		fmt.Sprintf("Path: %s", relPath),
		fmt.Sprintf("Number: %d", number),
		fmt.Sprintf("Title: %s", title),
		"",
		"Changes:",
		fmt.Sprintf("  + %s", relPath),
		fmt.Sprintf("    + # %d. %s", number, title),
		"    + Accepted",
	}, "\n")
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
	return adrlog.ParseRecord(absPath)
}

func (w *wizardModel) execute(repo *adrlog.Repository) (string, error) {
	switch w.kind {
	case wizardCreate:
		return w.executeCreate(repo)
	case wizardSupersede:
		return w.executeSupersede(repo)
	case wizardLink:
		return w.executeLink(repo)
	}
	return "", nil
}

func (w *wizardModel) executeCreate(repo *adrlog.Repository) (string, error) {
	title := strings.TrimSpace(w.inputs[0].Value())
	if title == "" {
		return "", fmt.Errorf("title is required")
	}

	path, err := repo.CreateADR(adrlog.CreateOptions{Title: title})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created %s", path), nil
}

func (w *wizardModel) executeSupersede(repo *adrlog.Repository) (string, error) {
	title := strings.TrimSpace(w.inputs[0].Value())
	if title == "" {
		return "", fmt.Errorf("title is required")
	}

	targetRef := fmt.Sprintf("%d", w.subjectRecord.Number)
	path, err := repo.CreateADR(adrlog.CreateOptions{
		Title:      title,
		Supersedes: []string{targetRef},
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created %s (supersedes ADR %d)", path, w.subjectRecord.Number), nil
}

func (w *wizardModel) executeLink(repo *adrlog.Repository) (string, error) {
	targetRef := strings.TrimSpace(w.inputs[0].Value())
	fwdLabel := strings.TrimSpace(w.inputs[1].Value())
	revLabel := strings.TrimSpace(w.inputs[2].Value())

	if targetRef == "" || fwdLabel == "" || revLabel == "" {
		return "", fmt.Errorf("all fields are required")
	}

	targetRecord, err := w.linkTargetRecord()
	if err != nil {
		return "", err
	}
	if targetRecord == nil {
		return "", fmt.Errorf("target ADR is required")
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
		return "", err
	}

	return fmt.Sprintf("Linked ADR %d -> %d", w.subjectRecord.Number, targetRecord.Number), nil
}
