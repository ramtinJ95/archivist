package adrlog_test

import (
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestListFilesMatchesADRPattern(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)
	testutil.SeedADR(t, adrDir, "README.md", "not an ADR")
	testutil.SeedADR(t, adrDir, "notes.txt", "not an ADR")

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, err := repo.ListFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(files), files)
	}

	want0 := filepath.Join("doc", "adr", "0001-record-architecture-decisions.md")
	want1 := filepath.Join("doc", "adr", "0002-use-go-for-implementation.md")

	if files[0] != want0 {
		t.Errorf("files[0] = %q, want %q", files[0], want0)
	}
	if files[1] != want1 {
		t.Errorf("files[1] = %q, want %q", files[1], want1)
	}
}

func TestListFilesSortsLexicographically(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0003-third.md", testutil.SampleADR3)
	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-second.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, err := repo.ListFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatalf("got %d files, want 3", len(files))
	}

	for i := 0; i < len(files)-1; i++ {
		if files[i] >= files[i+1] {
			t.Errorf("files not sorted: %q >= %q", files[i], files[i+1])
		}
	}
}

func TestListFilesErrorsOnMissingDir(t *testing.T) {
	dir := t.TempDir()

	repo := &adrlog.Repository{
		CWD:    dir,
		ADRDir: "doc/adr",
	}

	_, err := repo.ListFiles()
	if err == nil {
		t.Fatal("expected error for missing ADR directory")
	}
}

func TestListFilesEmptyDir(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, err := repo.ListFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}

func TestListFilesMatchesFiveDigitADRPattern(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "9999-last-four-digit.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "10000-five-digits.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, err := repo.ListFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(files), files)
	}

	wantFiveDigit := filepath.Join("doc", "adr", "10000-five-digits.md")
	if files[0] != wantFiveDigit && files[1] != wantFiveDigit {
		t.Fatalf("expected %q in files, got %v", wantFiveDigit, files)
	}

	next, err := repo.NextNumber()
	if err != nil {
		t.Fatal(err)
	}
	if next != 10001 {
		t.Fatalf("NextNumber() = %d, want 10001", next)
	}
}
