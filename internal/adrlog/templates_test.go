package adrlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestResolveTemplateDefault(t *testing.T) {
	dir := t.TempDir()
	got, err := adrlog.ResolveTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != adrlog.DefaultTemplate {
		t.Errorf("expected default template, got %q", got)
	}
}

func TestResolveTemplateFromRepo(t *testing.T) {
	dir := t.TempDir()
	tplDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}

	custom := "# NUMBER. TITLE\n\nCustom template\n"
	testutil.WriteFile(t, filepath.Join(tplDir, "template.md"), custom)

	got, err := adrlog.ResolveTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != custom {
		t.Errorf("expected custom template, got %q", got)
	}
}

func TestResolveTemplateFromEnv(t *testing.T) {
	dir := t.TempDir()
	envTpl := filepath.Join(dir, "env-template.md")
	custom := "# NUMBER. TITLE\n\nEnv template\n"
	testutil.WriteFile(t, envTpl, custom)

	t.Setenv("ADR_TEMPLATE", envTpl)

	got, err := adrlog.ResolveTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != custom {
		t.Errorf("expected env template, got %q", got)
	}
}

func TestResolveTemplateEnvTakesPrecedence(t *testing.T) {
	dir := t.TempDir()

	tplDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	testutil.WriteFile(t, filepath.Join(tplDir, "template.md"), "repo template")

	envTpl := filepath.Join(dir, "env-template.md")
	testutil.WriteFile(t, envTpl, "env template")

	t.Setenv("ADR_TEMPLATE", envTpl)

	got, err := adrlog.ResolveTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "env template" {
		t.Errorf("expected env template to take precedence, got %q", got)
	}
}

func TestResolveTemplateReturnsErrorForUnreadableEnvTemplate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ADR_TEMPLATE", filepath.Join(dir, "missing-template.md"))

	_, err := adrlog.ResolveTemplate(dir)
	if err == nil {
		t.Fatal("expected unreadable ADR_TEMPLATE to return an error")
	}
}

func TestApplyTemplate(t *testing.T) {
	result := adrlog.ApplyTemplate(adrlog.DefaultTemplate, 5, "Use PostgreSQL", "2024-03-15", "Accepted")

	if !strings.Contains(result, "# 5. Use PostgreSQL") {
		t.Errorf("expected title line, got:\n%s", result)
	}
	if !strings.Contains(result, "Date: 2024-03-15") {
		t.Errorf("expected date line, got:\n%s", result)
	}
	if !strings.Contains(result, "Accepted") {
		t.Errorf("expected status, got:\n%s", result)
	}
}

func TestPadNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{1, "0001"},
		{12, "0012"},
		{100, "0100"},
		{9999, "9999"},
	}

	for _, tt := range tests {
		got := adrlog.PadNumber(tt.n)
		if got != tt.want {
			t.Errorf("PadNumber(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
