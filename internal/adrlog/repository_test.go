package adrlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
)

func TestInitRepositoryDefaultDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ADR_DATE", "2024-01-15")

	relPath, err := adrlog.InitRepository(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join("doc", "adr", "0001-record-architecture-decisions.md")
	if relPath != want {
		t.Errorf("relPath = %q, want %q", relPath, want)
	}

	fullPath := filepath.Join(dir, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "# 1. Record architecture decisions") {
		t.Errorf("missing title in init ADR:\n%s", content)
	}
	if !strings.Contains(content, "Date: 2024-01-15") {
		t.Errorf("missing date in init ADR:\n%s", content)
	}

	dotFile := filepath.Join(dir, ".adr-dir")
	if _, err := os.Stat(dotFile); err == nil {
		t.Error(".adr-dir should not be created for default directory")
	}
}

func TestInitRepositoryCustomDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ADR_DATE", "2024-01-15")

	relPath, err := adrlog.InitRepository(dir, "architecture-log")
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join("architecture-log", "0001-record-architecture-decisions.md")
	if relPath != want {
		t.Errorf("relPath = %q, want %q", relPath, want)
	}

	dotFile := filepath.Join(dir, ".adr-dir")
	data, err := os.ReadFile(dotFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "architecture-log" {
		t.Errorf(".adr-dir = %q, want %q", string(data), "architecture-log")
	}

	adrPath := filepath.Join(dir, relPath)
	if _, err := os.Stat(adrPath); err != nil {
		t.Errorf("init ADR not created: %v", err)
	}
}

func TestInitRepositoryDiscoverable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ADR_DATE", "2024-01-15")

	_, err := adrlog.InitRepository(dir, "my-decisions")
	if err != nil {
		t.Fatal(err)
	}

	discovered, err := adrlog.DiscoverADRDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if discovered != "my-decisions" {
		t.Errorf("discovered %q, want %q", discovered, "my-decisions")
	}
}

func TestInitRepositoryUsesNextAvailableNumber(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ADR_DATE", "2024-01-15")

	firstPath, err := adrlog.InitRepository(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	secondPath, err := adrlog.InitRepository(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	wantFirst := filepath.Join("doc", "adr", "0001-record-architecture-decisions.md")
	if firstPath != wantFirst {
		t.Fatalf("firstPath = %q, want %q", firstPath, wantFirst)
	}

	wantSecond := filepath.Join("doc", "adr", "0002-record-architecture-decisions.md")
	if secondPath != wantSecond {
		t.Fatalf("secondPath = %q, want %q", secondPath, wantSecond)
	}

	if _, err := os.Stat(filepath.Join(dir, secondPath)); err != nil {
		t.Fatalf("expected second init ADR to be created: %v", err)
	}
}
