package adrlog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestAddLinkGolden(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	sourcePath := testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	targetPath := testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)

	if err := adrlog.AddLink(sourcePath, targetPath, "Clarifies", "Clarified by"); err != nil {
		t.Fatal(err)
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "link/link-source", sourceData)

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "link/link-target", targetData)
}

func TestSupersedeGolden(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	newPath, err := repo.CreateADR(adrlog.CreateOptions{
		Title:      "Use Rust instead",
		Date:       "2024-03-20",
		Supersedes: []string{"2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	newData, err := os.ReadFile(filepath.Join(dir, newPath))
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "supersede/supersede-new", newData)

	oldData, err := os.ReadFile(filepath.Join(adrDir, "0002-use-go-for-implementation.md"))
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "supersede/supersede-old", oldData)
}
