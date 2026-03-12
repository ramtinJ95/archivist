package adrlog_test

import (
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestValidate_AllValid(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-valid.md", testutil.SampleADR1)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-valid.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d: %v", len(issues), issues)
	}
}

func TestValidate_MissingTitle(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-no-title.md", `
Date: 2024-01-15

## Status

Accepted

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Message != "missing title" {
		t.Errorf("expected 'missing title', got %q", issues[0].Message)
	}
}

func TestValidate_MissingDate(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-no-date.md", `# 1. No date ADR

## Status

Accepted

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Message != "missing date" {
		t.Errorf("expected 'missing date', got %q", issues[0].Message)
	}
}

func TestValidate_EmptyStatusSection(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-no-status.md", `# 1. No status ADR

Date: 2024-01-15

## Status

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Message != "empty status section" {
		t.Errorf("expected 'empty status section', got %q", issues[0].Message)
	}
	if issues[0].Severity != "warning" {
		t.Errorf("expected severity 'warning', got %q", issues[0].Severity)
	}
}

func TestValidate_MissingStatusHeading(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-no-heading.md", `# 1. No status heading

Date: 2024-01-15

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Message != "missing ## Status heading" {
		t.Errorf("expected 'missing ## Status heading', got %q", issues[0].Message)
	}
	if issues[0].Severity != "error" {
		t.Errorf("expected severity 'error', got %q", issues[0].Severity)
	}
}

func TestValidate_BrokenRelationTarget(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-with-relation.md", `# 1. ADR with relation

Date: 2024-01-15

## Status

Accepted

Superseded by [2. New approach](0002-new-approach.md)

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, issue := range issues {
		if issue.Message == "broken relation target: 0002-new-approach.md" {
			found = true
			if issue.Severity != "error" {
				t.Errorf("expected severity 'error', got %q", issue.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected broken relation target issue, got %v", issues)
	}
}

func TestValidate_ValidRelationTarget(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-with-relation.md", `# 1. ADR with relation

Date: 2024-01-15

## Status

Accepted

Superseded by [2. Use Go](0002-use-go.md)

## Context

Some context.
`)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-use-go.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	for _, issue := range issues {
		if issue.Message == "broken relation target: 0002-use-go.md" {
			t.Errorf("did not expect broken relation issue, but got %v", issue)
		}
	}
}

func TestValidate_FileNumberMismatch(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0005-mismatched.md", `# 3. Mismatched number

Date: 2024-01-15

## Status

Accepted

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, issue := range issues {
		if issue.Message == "filename number 5 does not match title number 3" {
			found = true
			if issue.Severity != "warning" {
				t.Errorf("expected severity 'warning', got %q", issue.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected number mismatch issue, got %v", issues)
	}
}

func TestValidate_DuplicateADRNumber(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-foo.md", `# 2. Foo

Date: 2024-01-16

## Status

Accepted

## Context

Some context.
`)
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0002-bar.md", `# 2. Bar

Date: 2024-01-16

## Status

Accepted

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	var duplicateIssues []adrlog.ValidationIssue
	for _, issue := range issues {
		if issue.Message == "duplicate ADR number 2" {
			duplicateIssues = append(duplicateIssues, issue)
		}
	}

	if len(duplicateIssues) != 2 {
		t.Fatalf("expected 2 duplicate-number warnings, got %d: %v", len(duplicateIssues), issues)
	}
	for _, issue := range duplicateIssues {
		if issue.Severity != "warning" {
			t.Errorf("expected severity 'warning', got %q", issue.Severity)
		}
	}
}

func TestValidate_MultipleIssues(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-broken.md", `Something without structure

## Context

Some context.
`)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	issues, err := repo.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) < 2 {
		t.Fatalf("expected at least 2 issues, got %d: %v", len(issues), issues)
	}
}
