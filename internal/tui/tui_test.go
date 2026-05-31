package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func testModel(t *testing.T) Model {
	t.Helper()
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc/adr")
	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)

	repo := &adrlog.Repository{CWD: dir, ADRDir: "doc/adr"}
	records, err := loadRecords(repo)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(repo, records)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return updated.(Model)
}

func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func TestNewModelStartsInListView(t *testing.T) {
	m := testModel(t)

	if m.state != listView {
		t.Errorf("expected state listView (%d), got %d", listView, m.state)
	}
	if !m.ready {
		t.Error("expected ready to be true after WindowSizeMsg")
	}
}

func TestEnterSwitchesToDetailView(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(Model)

	if result.state != detailView {
		t.Errorf("expected state detailView (%d), got %d", detailView, result.state)
	}
}

func TestEscFromDetailReturnsToList(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	result := updated.(Model)

	if result.state != listView {
		t.Errorf("expected state listView (%d), got %d", listView, result.state)
	}
}

func TestQuestionMarkShowsHelp(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("?"))
	result := updated.(Model)

	if result.state != helpView {
		t.Errorf("expected state helpView (%d), got %d", helpView, result.state)
	}
}

func TestAnyKeyDismissesHelp(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("?"))
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("x"))
	result := updated.(Model)

	if result.state != listView {
		t.Errorf("expected state listView (%d), got %d", listView, result.state)
	}
}

func TestNStartsCreateWizard(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("n"))
	result := updated.(Model)

	if result.state != wizardView {
		t.Errorf("expected state wizardView (%d), got %d", wizardView, result.state)
	}
	if result.wizard.kind != wizardCreate {
		t.Errorf("expected wizard kind wizardCreate (%d), got %d", wizardCreate, result.wizard.kind)
	}
}

func TestGOpensGenerateView(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("g"))
	result := updated.(Model)

	if result.state != generateView {
		t.Errorf("expected state generateView (%d), got %d", generateView, result.state)
	}
}

func TestGenerateViewEscReturnsToList(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("g"))
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	result := updated.(Model)

	if result.state != listView {
		t.Errorf("expected state listView (%d), got %d", listView, result.state)
	}
}

func TestGenerateViewTGeneratesTOC(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("g"))
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("t"))
	result := updated.(Model)

	if result.state != detailView {
		t.Errorf("expected state detailView (%d), got %d", detailView, result.state)
	}
	if result.statusMsg != "Generated TOC" {
		t.Errorf("expected statusMsg %q, got %q", "Generated TOC", result.statusMsg)
	}
	if result.detailTitle != "Generated TOC" {
		t.Fatalf("expected detailTitle %q, got %q", "Generated TOC", result.detailTitle)
	}
	if !strings.Contains(result.detailViewport.View(), "Architecture Decision Records") {
		t.Fatalf("expected generated TOC content in detail viewport, got %q", result.detailViewport.View())
	}
}

func TestGenerateViewDGeneratesGraph(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("g"))
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("d"))
	result := updated.(Model)

	if result.state != detailView {
		t.Errorf("expected state detailView (%d), got %d", detailView, result.state)
	}
	if result.statusMsg != "Generated DOT graph" {
		t.Errorf("expected statusMsg %q, got %q", "Generated DOT graph", result.statusMsg)
	}
	if result.detailTitle != "Generated DOT graph" {
		t.Fatalf("expected detailTitle %q, got %q", "Generated DOT graph", result.detailTitle)
	}
	if !strings.Contains(result.detailViewport.View(), "digraph") {
		t.Fatalf("expected generated graph content in detail viewport, got %q", result.detailViewport.View())
	}
}

func TestGenerateViewUppercaseTOpensTOCExportWizard(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("g"))
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("T"))
	result := updated.(Model)

	if result.state != wizardView {
		t.Fatalf("expected wizardView, got %d", result.state)
	}
	if result.wizard.kind != wizardGenerateTOC {
		t.Fatalf("expected wizardGenerateTOC, got %d", result.wizard.kind)
	}
}

func TestGenerateViewUppercaseDOpensGraphExportWizard(t *testing.T) {
	m := testModel(t)

	updated, _ := m.Update(keyMsg("g"))
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("D"))
	result := updated.(Model)

	if result.state != wizardView {
		t.Fatalf("expected wizardView, got %d", result.state)
	}
	if result.wizard.kind != wizardGenerateGraph {
		t.Fatalf("expected wizardGenerateGraph, got %d", result.wizard.kind)
	}
}

func TestVOpensValidationView(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc/adr")
	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "decision.md", `# 2. Broken filename

Date: 2024-01-15

## Status

Accepted
`)

	repo := &adrlog.Repository{CWD: dir, ADRDir: "doc/adr"}
	records, err := loadRecords(repo)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(repo, records)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("v"))
	result := updated.(Model)

	if result.state != validationView {
		t.Fatalf("expected validationView, got %d", result.state)
	}
	if len(result.validationIssues) == 0 {
		t.Fatal("expected validation issues to be loaded")
	}
	if !strings.Contains(result.detailViewport.View(), "Validation report") {
		t.Fatalf("expected validation report content, got %q", result.detailViewport.View())
	}
	if !strings.Contains(result.detailViewport.View(), "> [WARNING]") {
		t.Fatalf("expected selected issue marker, got %q", result.detailViewport.View())
	}
}

func TestValidationViewNavigatesAndShowsAffectedFile(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc/adr")
	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0001-use-go-for-implementation.md", testutil.SampleADR2)

	repo := &adrlog.Repository{CWD: dir, ADRDir: "doc/adr"}
	records, err := loadRecords(repo)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(repo, records)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	updated, _ = m.Update(keyMsg("v"))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)

	if m.validationIndex != 1 {
		t.Fatalf("expected validation index 1, got %d", m.validationIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(Model)

	if result.state != validationView {
		t.Fatalf("expected to stay in validation view, got %d", result.state)
	}
	if !strings.Contains(result.detailViewport.View(), "Use Go for implementation") {
		t.Fatalf("expected affected ADR content, got %q", result.detailViewport.View())
	}
}

func TestValidationViewKeepsSelectedIssueVisible(t *testing.T) {
	m := testModel(t)
	m.validationIssues = nil
	for i := 0; i < 20; i++ {
		m.validationIssues = append(m.validationIssues, adrlog.ValidationIssue{
			Path:     filepath.Join("doc", "adr", "0001-record-architecture-decisions.md"),
			Severity: "warning",
			Message:  fmt.Sprintf("issue %02d", i),
		})
	}
	m.openValidationView()
	m.detailViewport.Height = 6

	for i := 0; i < 12; i++ {
		m.moveValidationSelection(1)
	}

	if m.detailViewport.YOffset == 0 {
		t.Fatal("expected validation viewport to scroll as selection moves")
	}
	if !strings.Contains(m.detailViewport.View(), "> [WARNING] doc/adr/0001-record-architecture-decisions.md: issue 12") {
		t.Fatalf("expected selected issue to remain visible, got %q", m.detailViewport.View())
	}
}

func TestValidationViewRefreshesAfterEditorCloses(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc/adr")
	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	badPath := filepath.Join(adrDir, "decision.md")
	if err := os.WriteFile(badPath, []byte(`# 2. Broken filename

Date: 2024-01-15

## Status

Accepted
`), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := &adrlog.Repository{CWD: dir, ADRDir: "doc/adr"}
	records, err := loadRecords(repo)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(repo, records)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	m = updated.(Model)
	m.openValidationView()
	if !strings.Contains(m.detailViewport.View(), "malformed ADR filename") {
		t.Fatalf("expected initial validation issue, got %q", m.detailViewport.View())
	}
	if err := os.Remove(badPath); err != nil {
		t.Fatal(err)
	}

	updated, _ = m.Update(editorFinishedMsg{})
	m = updated.(Model)

	if m.state != validationView {
		t.Fatalf("expected to remain in validation view, got %d", m.state)
	}
	if !strings.Contains(m.detailViewport.View(), "No validation issues found.") {
		t.Fatalf("expected refreshed validation report, got %q", m.detailViewport.View())
	}
}

func TestValidationViewEditAffectedFileRequiresEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	m := testModel(t)
	badPath := filepath.Join(m.repo.CWD, "doc/adr/0003-broken.md")
	if err := os.WriteFile(badPath, []byte(`# 3. Broken

## Status

Accepted
`), 0o644); err != nil {
		t.Fatal(err)
	}
	m.refreshValidationIssues()
	m.openValidationView()

	updated, cmd := m.Update(keyMsg("e"))
	result := updated.(Model)

	if cmd != nil {
		t.Fatal("expected no editor command when editor is unset")
	}
	if result.statusMsg != "No $EDITOR or $VISUAL set" {
		t.Fatalf("expected editor status message, got %q", result.statusMsg)
	}
}

func TestLoadRecordsReturnsMalformedADRError(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc/adr")
	testutil.SeedADR(t, adrDir, "0001-broken.md", `# 1. Broken ADR

Date: 2024-01-15

## Context
`)

	repo := &adrlog.Repository{CWD: dir, ADRDir: "doc/adr"}
	_, err := loadRecords(repo)
	if err == nil {
		t.Fatal("expected malformed ADR to be surfaced")
	}
	if !strings.Contains(err.Error(), "missing ## Status heading") {
		t.Fatalf("expected missing status error, got %v", err)
	}
}

func TestQuitFromListView(t *testing.T) {
	m := testModel(t)

	_, cmd := m.Update(keyMsg("q"))

	if cmd == nil {
		t.Fatal("expected a non-nil cmd for quit")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}
