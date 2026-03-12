package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/editor"
)

type viewState int

const (
	listView viewState = iota
	detailView
	helpView
	wizardView
	generateView
)

type editorFinishedMsg struct{ err error }

type Model struct {
	repo             *adrlog.Repository
	list             list.Model
	viewport         viewport.Model
	wizard           wizardModel
	state            viewState
	statusMsg        string
	generatedContent string
	width            int
	height           int
	ready            bool
}

func NewModel(repo *adrlog.Repository, records []*adrlog.Record) Model {
	items := make([]list.Item, len(records))
	for i, rec := range records {
		items[i] = ADRItem{record: rec}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderLeftForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("170")).
		BorderLeftForeground(lipgloss.Color("170"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Architecture Decision Records"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	vp := viewport.New(0, 0)

	return Model{
		repo:     repo,
		list:     l,
		viewport: vp,
		state:    listView,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateLayout()
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Editor error: %v", msg.err)
		} else {
			m.statusMsg = "Editor closed"
		}
		m.reloadRecords()
		m.updatePreview()
		return m, nil

	case tea.KeyMsg:
		if m.state == wizardView {
			return m.updateWizardView(msg)
		}

		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			m.updatePreview()
			return m, cmd
		}

		switch m.state {
		case listView:
			return m.updateListView(msg)
		case detailView:
			return m.updateDetailView(msg)
		case helpView:
			return m.updateHelpView(msg)
		case generateView:
			return m.updateGenerateView(msg)
		}
	}

	if m.state == listView {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		m.updatePreview()
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	switch m.state {
	case detailView:
		return m.renderDetailView()
	case helpView:
		return m.renderHelpView()
	case wizardView:
		return m.wizard.view(m.width, m.height)
	case generateView:
		return m.renderGenerateView()
	default:
		return m.renderListView()
	}
}

func (m *Model) updateLayout() {
	statusBarHeight := 1
	availableHeight := m.height - statusBarHeight

	switch m.state {
	case listView:
		listWidth := m.width * 2 / 5
		previewWidth := m.width - listWidth

		borderH, borderV := listPanelBorder.GetFrameSize()
		m.list.SetSize(listWidth-borderH, availableHeight-borderV)

		borderH, borderV = detailPanelBorder.GetFrameSize()
		m.viewport.Width = previewWidth - borderH
		m.viewport.Height = availableHeight - borderV

	case detailView:
		borderH, borderV := detailPanelBorder.GetFrameSize()
		m.viewport.Width = m.width - borderH
		m.viewport.Height = availableHeight - borderV
	}

	m.updatePreview()
}

func (m *Model) updatePreview() {
	if item, ok := m.list.SelectedItem().(ADRItem); ok {
		m.viewport.SetContent(item.record.Content)
		m.viewport.GotoTop()
	}
}

func (m Model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter":
		m.state = detailView
		m.updateLayout()
		return m, nil
	case "/":
		m.list.ResetFilter()
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case "?":
		m.state = helpView
		return m, nil
	case "n":
		m.wizard = newCreateWizard()
		m.state = wizardView
		m.statusMsg = ""
		return m, nil
	case "e":
		return m.openEditorForSelected()
	case "s":
		if item, ok := m.list.SelectedItem().(ADRItem); ok {
			m.wizard = newSupersedeWizard(item.record)
			m.state = wizardView
			m.statusMsg = ""
		}
		return m, nil
	case "l":
		if item, ok := m.list.SelectedItem().(ADRItem); ok {
			m.wizard = newLinkWizard(item.record)
			m.state = wizardView
			m.statusMsg = ""
		}
		return m, nil
	case "g":
		m.state = generateView
		m.statusMsg = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.updatePreview()
	return m, cmd
}

func (m Model) updateWizardView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cmd := m.wizard.update(msg)

	if m.wizard.done {
		if m.wizard.cancelled {
			m.state = listView
			m.statusMsg = "Cancelled"
			m.updateLayout()
			return m, nil
		}

		result, err := m.wizard.execute(m.repo)
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = result
		}

		m.reloadRecords()
		m.state = listView
		m.updateLayout()
		return m, nil
	}

	return m, cmd
}

func (m *Model) openEditorForSelected() (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(ADRItem)
	if !ok {
		return m, nil
	}

	editorCmd := editor.ResolveEditor()
	if editorCmd == "" {
		m.statusMsg = "No $EDITOR or $VISUAL set"
		return m, nil
	}

	absPath := item.record.Path
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(m.repo.CWD, absPath)
	}

	c := editor.EditorCommand(editorCmd, absPath)

	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m *Model) reloadRecords() {
	records, err := loadRecords(m.repo)
	if err != nil {
		return
	}

	items := make([]list.Item, len(records))
	for i, rec := range records {
		items[i] = ADRItem{record: rec}
	}

	m.list.SetItems(items)
}

func (m Model) updateDetailView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.state = listView
		m.updateLayout()
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) updateHelpView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.state = listView
	return m, nil
}

func (m Model) updateGenerateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "t":
		toc, err := m.repo.GenerateTOC(adrlog.TOCOptions{})
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
			m.state = listView
			m.updateLayout()
			return m, nil
		}
		m.generatedContent = toc
		m.viewport.SetContent(toc)
		m.state = detailView
		m.updateLayout()
		m.statusMsg = "Generated TOC"
		return m, nil
	case "d":
		graph, err := m.repo.GenerateGraph(adrlog.GraphOptions{})
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
			m.state = listView
			m.updateLayout()
			return m, nil
		}
		m.generatedContent = graph
		m.viewport.SetContent(graph)
		m.state = detailView
		m.updateLayout()
		m.statusMsg = "Generated DOT graph"
		return m, nil
	case "esc", "q":
		m.state = listView
		m.updateLayout()
		return m, nil
	}
	return m, nil
}

func (m Model) renderListView() string {
	listWidth := m.width * 2 / 5
	previewWidth := m.width - listWidth

	statusBarHeight := 1
	availableHeight := m.height - statusBarHeight

	listPanel := listPanelBorder.
		Width(listWidth - listPanelBorder.GetHorizontalFrameSize()).
		Height(availableHeight - listPanelBorder.GetVerticalFrameSize()).
		Render(m.list.View())

	previewPanel := detailPanelBorder.
		Width(previewWidth - detailPanelBorder.GetHorizontalFrameSize()).
		Height(availableHeight - detailPanelBorder.GetVerticalFrameSize()).
		Render(m.viewport.View())

	panels := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, previewPanel)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)
}

func (m Model) renderDetailView() string {
	statusBarHeight := 1
	availableHeight := m.height - statusBarHeight

	content := detailPanelBorder.
		Width(m.width - detailPanelBorder.GetHorizontalFrameSize()).
		Height(availableHeight - detailPanelBorder.GetVerticalFrameSize()).
		Render(m.viewport.View())

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m Model) renderHelpView() string {
	bindings := []struct{ key, desc string }{
		{"j/k, up/down", "Navigate list"},
		{"enter", "Open detail view"},
		{"/", "Filter ADRs by title"},
		{"n", "Create new ADR"},
		{"e", "Edit selected ADR in $EDITOR"},
		{"s", "Supersede selected ADR"},
		{"l", "Link selected ADR"},
		{"g", "Generate TOC or graph"},
		{"?", "Show this help"},
		{"esc", "Back / Cancel filter"},
		{"q, ctrl+c", "Quit"},
		{"", ""},
		{"Detail view:", ""},
		{"j/k, up/down", "Scroll content"},
		{"esc, backspace", "Back to list"},
		{"q", "Quit"},
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Keybindings"))
	sb.WriteString("\n\n")

	for _, b := range bindings {
		if b.key == "" {
			sb.WriteString("\n")
			continue
		}
		if b.desc == "" {
			sb.WriteString(titleStyle.Render(b.key))
			sb.WriteString("\n")
			continue
		}
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			helpKeyStyle.Render(fmt.Sprintf("%-18s", b.key)),
			helpDescStyle.Render(b.desc),
		))
	}

	sb.WriteString("\n" + helpDescStyle.Render("Press any key to dismiss"))

	content := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		sb.String(),
	)

	return content
}

func (m Model) renderGenerateView() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Generate"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		helpKeyStyle.Render("t"),
		helpDescStyle.Render("Generate Table of Contents"),
	))
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		helpKeyStyle.Render("d"),
		helpDescStyle.Render("Generate DOT dependency graph"),
	))
	sb.WriteString("\n")
	sb.WriteString(helpDescStyle.Render("esc: back"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		sb.String(),
	)
}

func (m Model) renderStatusBar() string {
	var leftText string
	if m.statusMsg != "" {
		leftText = m.statusMsg
	} else if item, ok := m.list.SelectedItem().(ADRItem); ok {
		leftText = filepath.Base(item.record.Path)
	}

	left := statusBarStyle.Render(leftText)

	var hints string
	switch m.state {
	case listView:
		hints = "↑/↓:nav  enter:detail  /:filter  n:new  e:edit  s:supersede  l:link  g:generate  ?:help  q:quit"
	case detailView:
		hints = "↑/↓:scroll  esc:back  q:quit"
	}
	right := statusBarStyle.Render(hints)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	mid := statusBarStyle.Render(strings.Repeat(" ", gap))

	return left + mid + right
}

func Run(repo *adrlog.Repository) error {
	records, err := loadRecords(repo)
	if err != nil {
		return err
	}

	m := NewModel(repo, records)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func loadRecords(repo *adrlog.Repository) ([]*adrlog.Record, error) {
	files, err := repo.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("listing ADR files: %w", err)
	}

	var records []*adrlog.Record
	for _, f := range files {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(repo.CWD, f)
		}
		rec, err := adrlog.ParseRecord(absPath)
		if err != nil {
			continue
		}
		records = append(records, rec)
	}

	return records, nil
}
