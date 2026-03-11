package editor_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLaunchEditorSupportsCommandWithFlags(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "editor.log")
	scriptPath := filepath.Join(dir, "fake-editor.sh")

	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"" + logPath + "\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("VISUAL", scriptPath+" --wait --reuse-window")
	targetPath := filepath.Join(dir, "decision.md")

	if err := editor.LaunchEditor(targetPath); err != nil {
		t.Fatalf("LaunchEditor returned error: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	got := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"--wait", "--reuse-window", targetPath}
	if len(got) != len(want) {
		t.Fatalf("got args %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, got[i], want[i])
		}
	}
}
