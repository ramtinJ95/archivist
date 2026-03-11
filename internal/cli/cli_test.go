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

	_, err := executeCommand("link", "1", "2", "Relates to", "Relates to")
	if err != nil {
		t.Fatalf("link failed: %v", err)
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
