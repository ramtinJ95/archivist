package adrlog_test

import (
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func setupRefsRepo(t *testing.T) *adrlog.Repository {
	t.Helper()
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)
	testutil.SeedADR(t, adrDir, "0003-use-cobra-for-cli.md", testutil.SampleADR3)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestResolveRefByNumber(t *testing.T) {
	repo := setupRefsRepo(t)

	got, err := repo.ResolveRef("2")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("doc", "adr", "0002-use-go-for-implementation.md")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveRefByPaddedNumber(t *testing.T) {
	repo := setupRefsRepo(t)

	got, err := repo.ResolveRef("0003")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("doc", "adr", "0003-use-cobra-for-cli.md")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveRefByFullFilename(t *testing.T) {
	repo := setupRefsRepo(t)

	got, err := repo.ResolveRef("0001-record-architecture-decisions.md")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("doc", "adr", "0001-record-architecture-decisions.md")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveRefByPartialFilename(t *testing.T) {
	repo := setupRefsRepo(t)

	got, err := repo.ResolveRef("cobra")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("doc", "adr", "0003-use-cobra-for-cli.md")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveRefNotFound(t *testing.T) {
	repo := setupRefsRepo(t)

	_, err := repo.ResolveRef("999")
	if err == nil {
		t.Fatal("expected error for non-existent ref")
	}
}

func TestResolveRecord(t *testing.T) {
	repo := setupRefsRepo(t)

	rec, err := repo.ResolveRecord("1")
	if err != nil {
		t.Fatal(err)
	}

	if rec.Number != 1 {
		t.Errorf("Number = %d, want 1", rec.Number)
	}
	if rec.Title != "Record architecture decisions" {
		t.Errorf("Title = %q, want %q", rec.Title, "Record architecture decisions")
	}
}
