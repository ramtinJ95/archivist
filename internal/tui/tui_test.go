package tui

import (
	"path/filepath"
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
