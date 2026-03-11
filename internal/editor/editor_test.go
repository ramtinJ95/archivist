package editor_test

import (
	"testing"

	"github.com/ramtinJ95/archivist/internal/editor"
)

func TestResolveEditorVisualPrecedence(t *testing.T) {
	t.Setenv("VISUAL", "code")
	t.Setenv("EDITOR", "vim")

	got := editor.ResolveEditor()
	if got != "code" {
		t.Errorf("got %q, want %q", got, "code")
	}
}

func TestResolveEditorFallsBackToEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nano")

	got := editor.ResolveEditor()
	if got != "nano" {
		t.Errorf("got %q, want %q", got, "nano")
	}
}

func TestResolveEditorEmptyWhenUnset(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	got := editor.ResolveEditor()
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestLaunchEditorNoOpWhenNoEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	err := editor.LaunchEditor("/tmp/test.md")
	if err != nil {
		t.Errorf("expected no error when no editor set, got: %v", err)
	}
}

func TestLaunchEditorNoOpWhenTrue(t *testing.T) {
	t.Setenv("VISUAL", "true")

	err := editor.LaunchEditor("/tmp/test.md")
	if err != nil {
		t.Errorf("expected no error with VISUAL=true, got: %v", err)
	}
}
