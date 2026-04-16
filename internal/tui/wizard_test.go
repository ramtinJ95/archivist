package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func wizardTestRepo(t *testing.T) (*adrlog.Repository, []*adrlog.Record) {
	t.Helper()
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc/adr")
	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)
	testutil.SeedADR(t, adrDir, "0003-use-cobra-for-cli.md", testutil.SampleADR3)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	records, err := loadRecords(repo)
	if err != nil {
		t.Fatal(err)
	}
	return repo, records
}

func TestCreateWizardHasSingleInput(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)

	if len(w.inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(w.inputs))
	}
	if w.labels[0] != "Title" {
		t.Errorf("expected label %q, got %q", "Title", w.labels[0])
	}
}

func TestLinkWizardPlaceholders(t *testing.T) {
	repo, records := wizardTestRepo(t)
	w := newLinkWizard(repo, records[0], records)

	if len(w.inputs) != 3 {
		t.Fatalf("expected 3 inputs, got %d", len(w.inputs))
	}
	if w.inputs[1].Placeholder != "e.g. Amends" {
		t.Errorf("expected forward placeholder %q, got %q", "e.g. Amends", w.inputs[1].Placeholder)
	}
	if w.inputs[2].Placeholder != "e.g. Amended by" {
		t.Errorf("expected reverse placeholder %q, got %q", "e.g. Amended by", w.inputs[2].Placeholder)
	}
}

func typeIntoWizard(w *wizardModel, text string) {
	for _, r := range text {
		w.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestWizardConfirmationStep(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Test")

	w.update(tea.KeyMsg{Type: tea.KeyEnter})

	if !w.confirming {
		t.Error("expected confirming to be true after Enter")
	}
	if w.done {
		t.Error("expected done to be false while confirming")
	}

	w.update(tea.KeyMsg{Type: tea.KeyEnter})

	if !w.done {
		t.Error("expected done to be true after confirming Enter")
	}
}

func TestWizardConfirmationEscGoesBack(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Test")

	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !w.confirming {
		t.Fatal("expected confirming to be true")
	}

	w.update(tea.KeyMsg{Type: tea.KeyEscape})

	if w.confirming {
		t.Error("expected confirming to be false after Esc")
	}
	if w.done {
		t.Error("expected done to be false after Esc from confirming")
	}
}

func TestWizardEscCancels(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)

	w.update(tea.KeyMsg{Type: tea.KeyEscape})

	if !w.cancelled {
		t.Error("expected cancelled to be true")
	}
	if !w.done {
		t.Error("expected done to be true")
	}
}

func TestWizardCtrlCAlwaysCancels(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Test")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !w.confirming {
		t.Fatal("expected confirming to be true")
	}

	w.update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if !w.cancelled {
		t.Error("expected cancelled to be true after Ctrl+C")
	}
	if !w.done {
		t.Error("expected done to be true after Ctrl+C")
	}
}

func TestConfirmationSummaryCreate(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Use PostgreSQL")

	summary := w.confirmationSummary()

	if !strings.Contains(summary, "Use PostgreSQL") {
		t.Errorf("expected summary to contain title, got %q", summary)
	}
}

func TestConfirmationSummarySupersede(t *testing.T) {
	repo, records := wizardTestRepo(t)
	w := newSupersedeWizard(repo, records[1])
	typeIntoWizard(&w, "New Decision")

	summary := w.confirmationSummary()

	if !strings.Contains(summary, "2") {
		t.Errorf("expected summary to contain target number, got %q", summary)
	}
	if !strings.Contains(summary, "Use Go for implementation") {
		t.Errorf("expected summary to contain target title, got %q", summary)
	}
}

func TestConfirmationSummaryLink(t *testing.T) {
	repo, records := wizardTestRepo(t)
	w := newLinkWizard(repo, records[0], records)

	typeIntoWizard(&w, "3")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "Amends")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "Amended by")

	summary := w.confirmationSummary()

	if !strings.Contains(summary, "3") {
		t.Errorf("expected summary to contain target ref, got %q", summary)
	}
	if !strings.Contains(summary, "Amends") {
		t.Errorf("expected summary to contain forward label, got %q", summary)
	}
	if !strings.Contains(summary, "Amended by") {
		t.Errorf("expected summary to contain reverse label, got %q", summary)
	}
}

func TestCreateWizardPreviewShowsPath(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Use PostgreSQL")

	preview := w.previewText()

	if !strings.Contains(preview, "doc/adr/0004-use-postgresql.md") {
		t.Fatalf("expected preview to include generated path, got %q", preview)
	}
	if !strings.Contains(preview, "+ Accepted") {
		t.Fatalf("expected preview to include initial status, got %q", preview)
	}
}

func TestSupersedeWizardPreviewShowsStatusDiff(t *testing.T) {
	repo, records := wizardTestRepo(t)
	w := newSupersedeWizard(repo, records[1])
	typeIntoWizard(&w, "Use Rust instead")

	preview := w.previewText()

	if !strings.Contains(preview, "doc/adr/0004-use-rust-instead.md") {
		t.Fatalf("expected preview path, got %q", preview)
	}
	if !strings.Contains(preview, "- Accepted") {
		t.Fatalf("expected preview to remove Accepted, got %q", preview)
	}
	if !strings.Contains(preview, "Superceded by [4. Use Rust instead](0004-use-rust-instead.md)") {
		t.Fatalf("expected preview to include reverse supersede link, got %q", preview)
	}
}

func TestLinkWizardPreviewUsesRepoAwareSelection(t *testing.T) {
	repo, records := wizardTestRepo(t)
	w := newLinkWizard(repo, records[0], records)

	typeIntoWizard(&w, "cobra")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "Amends")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "Amended by")

	preview := w.previewText()

	if !strings.Contains(preview, "Target: 3. Use Cobra for CLI") {
		t.Fatalf("expected preview to resolve target ADR, got %q", preview)
	}
	if !strings.Contains(preview, "Amends [3. Use Cobra for CLI](0003-use-cobra-for-cli.md)") {
		t.Fatalf("expected preview to include forward link mutation, got %q", preview)
	}
}

func TestLinkWizardArrowSelectsMatch(t *testing.T) {
	repo, records := wizardTestRepo(t)
	w := newLinkWizard(repo, records[0], records)

	if got := w.selectedLinkRecord(); got == nil || got.Number != 2 {
		t.Fatalf("expected first selectable match to be ADR 2, got %+v", got)
	}

	w.update(tea.KeyMsg{Type: tea.KeyDown})

	if got := w.selectedLinkRecord(); got == nil || got.Number != 3 {
		t.Fatalf("expected down arrow to select ADR 3, got %+v", got)
	}
}
