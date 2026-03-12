package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ramtinJ95/archivist/internal/adrlog"
)

func TestCreateWizardHasSingleInput(t *testing.T) {
	w := newCreateWizard()

	if len(w.inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(w.inputs))
	}
	if w.labels[0] != "Title" {
		t.Errorf("expected label %q, got %q", "Title", w.labels[0])
	}
}

func TestLinkWizardPlaceholders(t *testing.T) {
	rec := &adrlog.Record{Number: 1, Title: "Test ADR"}
	w := newLinkWizard(rec)

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
	w := newCreateWizard()
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
	w := newCreateWizard()
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
	w := newCreateWizard()

	w.update(tea.KeyMsg{Type: tea.KeyEscape})

	if !w.cancelled {
		t.Error("expected cancelled to be true")
	}
	if !w.done {
		t.Error("expected done to be true")
	}
}

func TestWizardCtrlCAlwaysCancels(t *testing.T) {
	w := newCreateWizard()
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
	w := newCreateWizard()
	typeIntoWizard(&w, "Use PostgreSQL")

	summary := w.confirmationSummary()

	if !strings.Contains(summary, "Use PostgreSQL") {
		t.Errorf("expected summary to contain title, got %q", summary)
	}
}

func TestConfirmationSummarySupersede(t *testing.T) {
	rec := &adrlog.Record{Number: 5, Title: "Old Decision"}
	w := newSupersedeWizard(rec)
	typeIntoWizard(&w, "New Decision")

	summary := w.confirmationSummary()

	if !strings.Contains(summary, "5") {
		t.Errorf("expected summary to contain target number, got %q", summary)
	}
	if !strings.Contains(summary, "Old Decision") {
		t.Errorf("expected summary to contain target title, got %q", summary)
	}
}

func TestConfirmationSummaryLink(t *testing.T) {
	rec := &adrlog.Record{Number: 3, Title: "Source ADR"}
	w := newLinkWizard(rec)

	typeIntoWizard(&w, "5")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "Amends")
	w.update(tea.KeyMsg{Type: tea.KeyEnter})
	typeIntoWizard(&w, "Amended by")

	summary := w.confirmationSummary()

	if !strings.Contains(summary, "5") {
		t.Errorf("expected summary to contain target ref, got %q", summary)
	}
	if !strings.Contains(summary, "Amends") {
		t.Errorf("expected summary to contain forward label, got %q", summary)
	}
	if !strings.Contains(summary, "Amended by") {
		t.Errorf("expected summary to contain reverse label, got %q", summary)
	}
}
