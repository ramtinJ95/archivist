package adrlog_test

import (
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestSearch_MatchesContent(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-first.md", testutil.SampleADR1)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-second.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("Go")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "doc/adr/0002-second.md" {
		t.Errorf("expected match in 0002-second.md, got %s", results[0].Path)
	}
	if len(results[0].Matches) == 0 {
		t.Fatal("expected at least one match")
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("architecture")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result for case-insensitive search")
	}
}

func TestSearch_NoMatch(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("zzzznonexistent")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_InvalidPattern(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.Search("[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestSearch_MatchLineNumbers(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("Accepted")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	for _, m := range results[0].Matches {
		if m.Line < 1 {
			t.Errorf("expected positive line number, got %d", m.Line)
		}
		if m.Content == "" {
			t.Error("expected non-empty content")
		}
	}
}

func TestSearch_MultipleFiles(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-first.md", testutil.SampleADR1)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-second.md", testutil.SampleADR2)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0003-third.md", testutil.SampleADR3)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("Accepted")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}
