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

func TestValidate_MissingStatus(t *testing.T) {
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
	if issues[0].Message != "missing status" {
		t.Errorf("expected 'missing status', got %q", issues[0].Message)
	}
	if issues[0].Severity != "warning" {
		t.Errorf("expected severity 'warning', got %q", issues[0].Severity)
	}
}

func TestValidate_MultipleIssues(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	testutil.SeedADR(t, filepath.Join(dir, "doc/adr"), "0001-broken.md", `Something without structure

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

	if len(issues) < 2 {
		t.Fatalf("expected at least 2 issues, got %d: %v", len(issues), issues)
	}
}
