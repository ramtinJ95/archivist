package tui

import (
	"fmt"
	"path/filepath"
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
	kind         wizardKind
	inputs       []textinput.Model
	labels       []string
	focusIndex   int
	done         bool
	cancelled    bool
	resultMsg    string
	targetRecord *adrlog.Record
}

func newCreateWizard() wizardModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "e.g. Use PostgreSQL for persistence"
	titleInput.CharLimit = 120
	titleInput.Width = 50
	titleInput.Focus()

	return wizardModel{
		kind:   wizardCreate,
		inputs: []textinput.Model{titleInput},
		labels: []string{"Title"},
	}
}

func newSupersedeWizard(target *adrlog.Record) wizardModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Title for the new ADR"
	titleInput.CharLimit = 120
	titleInput.Width = 50
	titleInput.Focus()

	return wizardModel{
		kind:         wizardSupersede,
		inputs:       []textinput.Model{titleInput},
		labels:       []string{"New ADR title"},
		targetRecord: target,
	}
}

func newLinkWizard(source *adrlog.Record) wizardModel {
	targetInput := textinput.New()
	targetInput.Placeholder = "ADR number or filename"
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

	return wizardModel{
		kind:         wizardLink,
		inputs:       []textinput.Model{targetInput, fwdInput, revInput},
		labels:       []string{"Target ADR ref", "Forward label", "Reverse label"},
		targetRecord: source,
	}
}

func (w *wizardModel) update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		w.cancelled = true
		w.done = true
		return nil
	case "esc":
		w.cancelled = true
		w.done = true
		return nil
	case "tab", "down":
		w.focusIndex = (w.focusIndex + 1) % len(w.inputs)
		return w.updateFocus()
	case "shift+tab", "up":
		w.focusIndex = (w.focusIndex - 1 + len(w.inputs)) % len(w.inputs)
		return w.updateFocus()
	case "enter":
		if w.focusIndex == len(w.inputs)-1 {
			w.done = true
			return nil
		}
		w.focusIndex++
		return w.updateFocus()
	}

	var cmd tea.Cmd
	w.inputs[w.focusIndex], cmd = w.inputs[w.focusIndex].Update(msg)
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

func (w *wizardModel) view(width, height int) string {
	var sb strings.Builder

	switch w.kind {
	case wizardCreate:
		sb.WriteString(titleStyle.Render("Create New ADR"))
	case wizardSupersede:
		sb.WriteString(titleStyle.Render(fmt.Sprintf("Supersede ADR %d: %s",
			w.targetRecord.Number, w.targetRecord.Title)))
	case wizardLink:
		sb.WriteString(titleStyle.Render(fmt.Sprintf("Link from ADR %d: %s",
			w.targetRecord.Number, w.targetRecord.Title)))
	}

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

	sb.WriteString(helpDescStyle.Render("tab/shift+tab: navigate fields  enter: submit  esc: cancel"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, sb.String())
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

	targetRef := fmt.Sprintf("%d", w.targetRecord.Number)
	path, err := repo.CreateADR(adrlog.CreateOptions{
		Title:      title,
		Supersedes: []string{targetRef},
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created %s (supersedes ADR %d)", path, w.targetRecord.Number), nil
}

func (w *wizardModel) executeLink(repo *adrlog.Repository) (string, error) {
	targetRef := strings.TrimSpace(w.inputs[0].Value())
	fwdLabel := strings.TrimSpace(w.inputs[1].Value())
	revLabel := strings.TrimSpace(w.inputs[2].Value())

	if targetRef == "" || fwdLabel == "" || revLabel == "" {
		return "", fmt.Errorf("all fields are required")
	}

	targetPath, err := repo.ResolveRef(targetRef)
	if err != nil {
		return "", err
	}
	absTargetPath := targetPath
	if !filepath.IsAbs(targetPath) {
		absTargetPath = filepath.Join(repo.CWD, targetPath)
	}

	absSourcePath := w.targetRecord.Path
	if !filepath.IsAbs(absSourcePath) {
		absSourcePath = filepath.Join(repo.CWD, absSourcePath)
	}

	if err := adrlog.AddLink(absSourcePath, absTargetPath, fwdLabel, revLabel); err != nil {
		return "", err
	}

	return fmt.Sprintf("Linked ADR %d -> %s", w.targetRecord.Number, targetRef), nil
}
