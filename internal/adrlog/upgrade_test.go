package adrlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestUpgradeRepositoryConvertsDateFormat(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	oldDateADR := `# 1. First decision

Date: 15/01/2024

## Status

Accepted

## Context

## Decision

## Consequences
`
	testutil.SeedADR(t, adrDir, "0001-first-decision.md", oldDateADR)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	upgraded, err := repo.UpgradeRepository()
	if err != nil {
		t.Fatal(err)
	}

	if upgraded != 1 {
		t.Errorf("upgraded = %d, want 1", upgraded)
	}

	data, err := os.ReadFile(filepath.Join(adrDir, "0001-first-decision.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "Date: 2024-01-15") {
		t.Errorf("expected converted date, got:\n%s", content)
	}
	if strings.Contains(content, "15/01/2024") {
		t.Errorf("old date format still present:\n%s", content)
	}
}

func TestUpgradeRepositorySkipsAlreadyUpgraded(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	upgraded, err := repo.UpgradeRepository()
	if err != nil {
		t.Fatal(err)
	}

	if upgraded != 0 {
		t.Errorf("upgraded = %d, want 0 (already in ISO format)", upgraded)
	}
}

func TestUpgradeRepositoryPreservesContent(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	oldDateADR := `# 1. First decision

Date: 20/03/2023

## Status

Accepted

## Context

Important context here.

## Decision

Important decision here.

## Consequences

Important consequences here.
`
	testutil.SeedADR(t, adrDir, "0001-first.md", oldDateADR)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.UpgradeRepository()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(adrDir, "0001-first.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "Important context here.") {
		t.Errorf("content not preserved:\n%s", content)
	}
	if !strings.Contains(content, "Important decision here.") {
		t.Errorf("content not preserved:\n%s", content)
	}
}
