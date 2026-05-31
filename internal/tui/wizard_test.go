package tui

import (
	"os"
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

func TestCreateWizardHasRichInputs(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)

	if len(w.inputs) != 3 {
		t.Fatalf("expected 3 inputs, got %d", len(w.inputs))
	}
	if w.labels[0] != "Title" {
		t.Errorf("expected label %q, got %q", "Title", w.labels[0])
	}
	if w.labels[1] != "Supersedes" {
		t.Errorf("expected label %q, got %q", "Supersedes", w.labels[1])
	}
	if w.labels[2] != "Links" {
		t.Errorf("expected label %q, got %q", "Links", w.labels[2])
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

func advanceWizardToConfirmation(w *wizardModel) {
	for !w.confirming && !w.done {
		w.update(tea.KeyMsg{Type: tea.KeyEnter})
	}
}

func TestWizardConfirmationStep(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Test")

	advanceWizardToConfirmation(&w)

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

	advanceWizardToConfirmation(&w)
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

func TestWizardConfirmationIgnoresOtherKeys(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Test")

	advanceWizardToConfirmation(&w)
	if !w.confirming {
		t.Fatal("expected confirming to be true")
	}

	w.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	w.update(tea.KeyMsg{Type: tea.KeyTab})

	if got := w.inputs[0].Value(); got != "Test" {
		t.Fatalf("confirmation screen mutated input to %q", got)
	}
	if !w.confirming {
		t.Fatal("expected to remain on confirmation screen")
	}
	if w.done {
		t.Fatal("expected non-confirmation keys to leave wizard unfinished")
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
	advanceWizardToConfirmation(&w)
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

func TestCreateWizardPreviewShowsSupersedeAndLinkDiffs(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Use Rust instead")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "2")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "1:Clarifies:Clarified by")

	preview := w.previewText()

	if !strings.Contains(preview, "Superceded by [4. Use Rust instead](0004-use-rust-instead.md)") {
		t.Fatalf("expected supersede mutation in preview, got %q", preview)
	}
	if !strings.Contains(preview, "Clarifies [1. Record architecture decisions](0001-record-architecture-decisions.md)") {
		t.Fatalf("expected forward link in preview, got %q", preview)
	}
	if !strings.Contains(preview, "Clarified by [4. Use Rust instead](0004-use-rust-instead.md)") {
		t.Fatalf("expected reverse link in preview, got %q", preview)
	}
}

func TestCreateWizardExecuteSupportsSupersedesAndLinks(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newCreateWizard(repo)
	typeIntoWizard(&w, "Use Rust instead")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "2")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "1:Clarifies:Clarified by")

	result, err := w.execute(repo)
	if err != nil {
		t.Fatal(err)
	}
	if result.editPath == "" || !result.reloadRecords {
		t.Fatalf("expected created ADR to request reload and edit, got %+v", result)
	}

	newData, err := os.ReadFile(filepath.Join(repo.CWD, "doc/adr/0004-use-rust-instead.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(newData), "Supercedes [2. Use Go for implementation](0002-use-go-for-implementation.md)") {
		t.Fatalf("new ADR missing supersede link:\n%s", string(newData))
	}
	oldData, err := os.ReadFile(filepath.Join(repo.CWD, "doc/adr/0001-record-architecture-decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(oldData), "Clarified by [4. Use Rust instead](0004-use-rust-instead.md)") {
		t.Fatalf("linked ADR missing reverse link:\n%s", string(oldData))
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

func TestGenerateTOCWizardExportsFileWithOptions(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	introPath := filepath.Join(repo.CWD, "intro.md")
	if err := os.WriteFile(introPath, []byte("Intro text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	w := newGenerateTOCWizard(repo)
	typeIntoWizard(&w, "out/adr-readme.md")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "intro.md")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "docs/adr/")

	preview := w.previewText()
	if !strings.Contains(preview, "Intro text.") {
		t.Fatalf("expected preview to include intro contents, got %q", preview)
	}

	result, err := w.execute(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.message, "Exported TOC") {
		t.Fatalf("expected export message, got %+v", result)
	}
	data, err := os.ReadFile(filepath.Join(repo.CWD, "out/adr-readme.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "(docs/adr/0001-record-architecture-decisions.md)") {
		t.Fatalf("expected exported TOC to use link prefix:\n%s", string(data))
	}
}

func TestGenerateGraphWizardExportsFileWithOptions(t *testing.T) {
	repo, _ := wizardTestRepo(t)
	w := newGenerateGraphWizard(repo)
	typeIntoWizard(&w, "out/graph.dot")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "https://example.test/")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, ".svg")

	result, err := w.execute(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.message, "Exported DOT graph") {
		t.Fatalf("expected export message, got %+v", result)
	}
	data, err := os.ReadFile(filepath.Join(repo.CWD, "out/graph.dot"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "https://example.test/0001-record-architecture-decisions.svg") {
		t.Fatalf("expected exported graph to use prefix and extension:\n%s", string(data))
	}
}
