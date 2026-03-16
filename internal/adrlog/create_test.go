package adrlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Use PostgreSQL", "use-postgresql"},
		{"Record architecture decisions", "record-architecture-decisions"},
		{"  Leading and trailing  ", "leading-and-trailing"},
		{"Special!@#chars", "special-chars"},
		{"multiple   spaces", "multiple-spaces"},
		{"Already-slugged", "already-slugged"},
		{"MixedCase Title", "mixedcase-title"},
	}

	for _, tt := range tests {
		got := adrlog.Slugify(tt.title)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestNextNumber(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-second.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	next, err := repo.NextNumber()
	if err != nil {
		t.Fatal(err)
	}

	if next != 3 {
		t.Errorf("NextNumber() = %d, want 3", next)
	}
}

func TestNextNumberEmptyRepo(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	next, err := repo.NextNumber()
	if err != nil {
		t.Fatal(err)
	}

	if next != 1 {
		t.Errorf("NextNumber() = %d, want 1", next)
	}
}

func TestGenerateFilename(t *testing.T) {
	repo := &adrlog.Repository{CWD: "/tmp", ADRDir: "doc/adr"}

	got := repo.GenerateFilename(5, "Use PostgreSQL")
	want := "0005-use-postgresql.md"
	if got != want {
		t.Errorf("GenerateFilename() = %q, want %q", got, want)
	}
}

func TestCreateADR(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")
	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	relPath, err := repo.CreateADR(adrlog.CreateOptions{
		Title: "Use PostgreSQL",
		Date:  "2024-03-15",
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedRel := filepath.Join("doc", "adr", "0002-use-postgresql.md")
	if relPath != expectedRel {
		t.Errorf("relPath = %q, want %q", relPath, expectedRel)
	}

	fullPath := filepath.Join(dir, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "# 2. Use PostgreSQL") {
		t.Errorf("expected title in content:\n%s", content)
	}
	if !strings.Contains(content, "Date: 2024-03-15") {
		t.Errorf("expected date in content:\n%s", content)
	}
	if !strings.Contains(content, "Accepted") {
		t.Errorf("expected status in content:\n%s", content)
	}
}

func TestCreateADRWithADRDateEnv(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	t.Setenv("ADR_DATE", "2020-01-01")

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	relPath, err := repo.CreateADR(adrlog.CreateOptions{
		Title: "Test env date",
	})
	if err != nil {
		t.Fatal(err)
	}

	fullPath := filepath.Join(dir, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "Date: 2020-01-01") {
		t.Errorf("expected ADR_DATE override, got:\n%s", string(data))
	}
}

func TestCreateADRFailsWhenADRTemplateCannotBeRead(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	t.Setenv("ADR_TEMPLATE", filepath.Join(dir, "missing-template.md"))

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.CreateADR(adrlog.CreateOptions{
		Title: "Use PostgreSQL",
	})
	if err == nil {
		t.Fatal("expected CreateADR to fail when ADR_TEMPLATE cannot be read")
	}

	if _, statErr := os.Stat(filepath.Join(dir, "doc/adr/0001-use-postgresql.md")); !os.IsNotExist(statErr) {
		t.Fatalf("expected ADR file to not be created, got stat err %v", statErr)
	}
}

func TestParseLinkSpec(t *testing.T) {
	spec, err := adrlog.ParseLinkSpec("5:Amends:Amended by")
	if err != nil {
		t.Fatal(err)
	}

	if spec.Target != "5" {
		t.Errorf("Target = %q, want %q", spec.Target, "5")
	}
	if spec.ForwardLabel != "Amends" {
		t.Errorf("ForwardLabel = %q, want %q", spec.ForwardLabel, "Amends")
	}
	if spec.ReverseLabel != "Amended by" {
		t.Errorf("ReverseLabel = %q, want %q", spec.ReverseLabel, "Amended by")
	}
}

func TestParseLinkSpecInvalid(t *testing.T) {
	_, err := adrlog.ParseLinkSpec("invalid")
	if err == nil {
		t.Fatal("expected error for invalid link spec")
	}
}
