package adrlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestAddLinkReciprocal(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	sourcePath := testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	targetPath := testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)

	err := adrlog.AddLink(sourcePath, targetPath, "Clarifies", "Clarified by")
	if err != nil {
		t.Fatal(err)
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	sourceContent := string(sourceData)

	if !strings.Contains(sourceContent, "Clarifies [2. Use Go for implementation](0002-use-go-for-implementation.md)") {
		t.Errorf("source missing forward link:\n%s", sourceContent)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	targetContent := string(targetData)

	if !strings.Contains(targetContent, "Clarified by [1. Record architecture decisions](0001-record-architecture-decisions.md)") {
		t.Errorf("target missing reverse link:\n%s", targetContent)
	}
}

func TestSupersedeViaCreateADR(t *testing.T) {
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

	newFullPath := filepath.Join(dir, newPath)
	newData, err := os.ReadFile(newFullPath)
	if err != nil {
		t.Fatal(err)
	}
	newContent := string(newData)

	if !strings.Contains(newContent, "Supercedes [2. Use Go for implementation](0002-use-go-for-implementation.md)") {
		t.Errorf("new ADR missing supercedes line:\n%s", newContent)
	}

	oldPath := filepath.Join(adrDir, "0002-use-go-for-implementation.md")
	oldData, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	oldContent := string(oldData)

	if !strings.Contains(oldContent, "Superceded by [3. Use Rust instead](0003-use-rust-instead.md)") {
		t.Errorf("old ADR missing superceded by line:\n%s", oldContent)
	}

	if strings.Contains(oldContent, "\nAccepted\n") {
		t.Errorf("old ADR should have Accepted removed:\n%s", oldContent)
	}
}

func TestLinkViaCreateADR(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	newPath, err := repo.CreateADR(adrlog.CreateOptions{
		Title: "Use Go",
		Date:  "2024-03-20",
		Links: []adrlog.LinkSpec{
			{Target: "1", ForwardLabel: "Amends", ReverseLabel: "Amended by"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	newFullPath := filepath.Join(dir, newPath)
	newData, err := os.ReadFile(newFullPath)
	if err != nil {
		t.Fatal(err)
	}
	newContent := string(newData)

	if !strings.Contains(newContent, "Amends [1. Record architecture decisions](0001-record-architecture-decisions.md)") {
		t.Errorf("new ADR missing forward link:\n%s", newContent)
	}

	oldPath := filepath.Join(adrDir, "0001-record-architecture-decisions.md")
	oldData, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	oldContent := string(oldData)

	if !strings.Contains(oldContent, "Amended by [2. Use Go](0002-use-go.md)") {
		t.Errorf("old ADR missing reverse link:\n%s", oldContent)
	}
}

func TestAddLinkFailsWithoutStatusHeading(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	noStatusContent := "# 1. Test ADR\n\nDate: 2024-01-15\n\n## Context\n\nSome context.\n"
	source := testutil.SeedADR(t, adrDir, "0001-test.md", noStatusContent)

	targetContent := "# 2. Target ADR\n\nDate: 2024-01-16\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"
	target := testutil.SeedADR(t, adrDir, "0002-target.md", targetContent)

	err := adrlog.AddLink(source, target, "Amends", "Amended by")
	if err == nil {
		t.Fatal("expected error when source has no ## Status heading")
	}
	if !strings.Contains(err.Error(), "no ## Status heading") {
		t.Errorf("expected error about missing Status heading, got: %v", err)
	}
}

func TestAddLinkPreservesExistingContent(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	sourcePath := testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)
	targetPath := testutil.SeedADR(t, adrDir, "0002-second.md", testutil.SampleADR2)

	err := adrlog.AddLink(sourcePath, targetPath, "Related to", "Related to")
	if err != nil {
		t.Fatal(err)
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	sourceContent := string(sourceData)

	if !strings.Contains(sourceContent, "## Context") {
		t.Errorf("original content sections missing:\n%s", sourceContent)
	}
	if !strings.Contains(sourceContent, "## Decision") {
		t.Errorf("original content sections missing:\n%s", sourceContent)
	}
	if !strings.Contains(sourceContent, "## Consequences") {
		t.Errorf("original content sections missing:\n%s", sourceContent)
	}
}

func TestAddLinkRollsBackSourceWhenTargetUpdateFails(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	sourcePath := testutil.SeedADR(t, adrDir, "0001-source.md", testutil.SampleADR1)
	targetPath := testutil.SeedADR(t, adrDir, "0002-target.md", "# 2. Target ADR\n\nDate: 2024-01-16\n\n## Context\n\nSome context.\n")

	originalSourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	err = adrlog.AddLink(sourcePath, targetPath, "Clarifies", "Clarified by")
	if err == nil {
		t.Fatal("expected AddLink to fail when target has no ## Status heading")
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(sourceData) != string(originalSourceData) {
		t.Fatalf("source ADR changed after failed link:\n%s", string(sourceData))
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(targetData), "Clarified by") {
		t.Fatalf("target ADR should not be mutated on failure:\n%s", string(targetData))
	}
}
