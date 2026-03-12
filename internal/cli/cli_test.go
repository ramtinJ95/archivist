// These tests mutate global state (rootCmd and os.Chdir) and must NOT use
// t.Parallel(). The shared Cobra rootCmd is reused across calls, and chdir
// affects the entire process. Running them concurrently would cause races.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-use-go-for-implementation.md", testutil.SampleADR2)
	return dir
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionCommand(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(out, "archivist dev") {
		t.Errorf("expected 'archivist dev', got %q", out)
	}
}

func TestInitCommand(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	out, err := executeCommand("init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if !strings.Contains(out, "0001-record-architecture-decisions.md") {
		t.Errorf("expected init output to contain seed file, got %q", out)
	}
}

func TestInitCommandCustomDir(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	out, err := executeCommand("init", "decisions")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if !strings.Contains(out, "decisions") {
		t.Errorf("expected 'decisions' in output, got %q", out)
	}
}

func TestListCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "0001") {
		t.Errorf("expected list to contain 0001, got %q", out)
	}
	if !strings.Contains(out, "0002") {
		t.Errorf("expected list to contain 0002, got %q", out)
	}
}

func TestShowCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("show", "1")
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(out, "Record architecture decisions") {
		t.Errorf("expected ADR content, got %q", out)
	}
}

func TestSearchCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("search", "Go")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !strings.Contains(out, "0002") {
		t.Errorf("expected search to find ADR 2, got %q", out)
	}
}

func TestSearchCommandCaseInsensitive(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("search", "architecture")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !strings.Contains(out, "0001") {
		t.Errorf("expected search to find ADR 1, got %q", out)
	}
}

func TestValidateCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	_, err := executeCommand("validate")
	if err != nil {
		t.Fatalf("validate failed on valid repo: %v", err)
	}
}

func TestValidateCommandFindsIssues(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-broken.md", `no title here

## Status

## Context
`)
	chdir(t, dir)

	out, err := executeCommand("validate")
	if err == nil {
		t.Fatal("expected validate to fail on broken ADR")
	}
	if !strings.Contains(out, "missing") {
		t.Errorf("expected 'missing' in output, got %q", out)
	}
}

func TestGenerateTOCCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("generate", "toc")
	if err != nil {
		t.Fatalf("generate toc failed: %v", err)
	}
	if !strings.Contains(out, "Architecture Decision Records") {
		t.Errorf("expected TOC header, got %q", out)
	}
}

func TestGenerateTOCCommandReadsIntroAndOutroFiles(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	introPath := filepath.Join(dir, "intro.md")
	outroPath := filepath.Join(dir, "outro.md")
	if err := os.WriteFile(introPath, []byte("Intro text.\n\nMore intro.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outroPath, []byte("Outro text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand("generate", "toc", "-i", introPath, "-o", outroPath)
	if err != nil {
		t.Fatalf("generate toc with intro/outro failed: %v", err)
	}

	if !strings.Contains(out, "Intro text.\n\nMore intro.") {
		t.Fatalf("expected intro file contents in output, got %q", out)
	}
	if !strings.Contains(out, "Outro text.") {
		t.Fatalf("expected outro file contents in output, got %q", out)
	}
	if strings.Contains(out, introPath) || strings.Contains(out, outroPath) {
		t.Fatalf("expected TOC output to contain file contents, not file paths: %q", out)
	}
}

func TestGenerateGraphCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("generate", "graph")
	if err != nil {
		t.Fatalf("generate graph failed: %v", err)
	}
	if !strings.Contains(out, "digraph") {
		t.Errorf("expected digraph output, got %q", out)
	}
}

func TestNewCommand(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	_, err := adrlog.InitRepository(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("EDITOR", "true")

	out, err := executeCommand("new", "Use", "PostgreSQL")
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}
	if !strings.Contains(out, "use-postgresql") {
		t.Errorf("expected slugified filename in output, got %q", out)
	}
}

func TestLinkCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	_, err := executeCommand("link", "2", "Clarifies", "1", "Clarified by")
	if err != nil {
		t.Fatalf("link failed: %v", err)
	}

	sourceContent, err := os.ReadFile(filepath.Join(dir, "doc/adr/0002-use-go-for-implementation.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sourceContent), "Clarifies [1. Record architecture decisions](0001-record-architecture-decisions.md)") {
		t.Fatalf("source ADR missing forward link:\n%s", string(sourceContent))
	}

	targetContent, err := os.ReadFile(filepath.Join(dir, "doc/adr/0001-record-architecture-decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(targetContent), "Clarified by [2. Use Go for implementation](0002-use-go-for-implementation.md)") {
		t.Fatalf("target ADR missing reverse link:\n%s", string(targetContent))
	}
}

func TestUpgradeCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("upgrade-repository")
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if !strings.Contains(out, "Upgraded") {
		t.Errorf("expected 'Upgraded' in output, got %q", out)
	}
}

func TestEditCommand(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	t.Setenv("EDITOR", "true")

	_, err := executeCommand("edit", "1")
	if err != nil {
		t.Fatalf("edit failed: %v", err)
	}
}

func TestNewCommandWithSupersede(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	_, err := adrlog.InitRepository(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("EDITOR", "true")

	_, err = executeCommand("new", "First")
	if err != nil {
		t.Fatalf("new First failed: %v", err)
	}

	out, err := executeCommand("new", "-s", "1", "Replacement")
	if err != nil {
		t.Fatalf("new -s 1 Replacement failed: %v", err)
	}
	if !strings.Contains(out, "replacement") {
		t.Errorf("expected new filename in output, got %q", out)
	}

	oldADR, err := os.ReadFile(filepath.Join(dir, "doc/adr/0001-record-architecture-decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(oldADR), "Superceded by") {
		t.Errorf("expected old ADR to contain 'Superceded by', got:\n%s", string(oldADR))
	}
}

func TestNewCommandWithLink(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	_, err := adrlog.InitRepository(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("EDITOR", "true")

	out, err := executeCommand("new", "-l", "1:Amends:Amended by", "Second", "ADR")
	if err != nil {
		t.Fatalf("new -l failed: %v", err)
	}
	if !strings.Contains(out, "second-adr") {
		t.Errorf("expected new filename in output, got %q", out)
	}

	oldADR, err := os.ReadFile(filepath.Join(dir, "doc/adr/0001-record-architecture-decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(oldADR), "Amended by") {
		t.Errorf("expected old ADR to contain 'Amended by', got:\n%s", string(oldADR))
	}
}

func TestUpgradeAlias(t *testing.T) {
	dir := setupTestRepo(t)
	chdir(t, dir)

	out, err := executeCommand("upgrade")
	if err != nil {
		t.Fatalf("upgrade alias failed: %v", err)
	}
	if !strings.Contains(out, "Upgraded") {
		t.Errorf("expected 'Upgraded' in output, got %q", out)
	}
}
