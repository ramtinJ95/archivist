package adrlog_test

import (
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestParseRecordExtractsMetadata(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	path := testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)

	rec, err := adrlog.ParseRecord(path)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Number != 1 {
		t.Errorf("Number = %d, want 1", rec.Number)
	}
	if rec.Title != "Record architecture decisions" {
		t.Errorf("Title = %q, want %q", rec.Title, "Record architecture decisions")
	}
	if rec.Date != "2024-01-15" {
		t.Errorf("Date = %q, want %q", rec.Date, "2024-01-15")
	}
	if len(rec.Status) != 1 || rec.Status[0] != "Accepted" {
		t.Errorf("Status = %v, want [Accepted]", rec.Status)
	}
}

func TestParseRecordMultipleStatusLines(t *testing.T) {
	content := `# 5. Use PostgreSQL

Date: 2024-02-01

## Status

Accepted

Superceded by [6. Use SQLite](0006-use-sqlite.md)

## Context

We need a database.

## Decision

Use PostgreSQL.

## Consequences

None.
`
	rec, err := adrlog.ParseRecordFromContent("0005-use-postgresql.md", content)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Number != 5 {
		t.Errorf("Number = %d, want 5", rec.Number)
	}
	if len(rec.Status) != 2 {
		t.Fatalf("Status has %d lines, want 2: %v", len(rec.Status), rec.Status)
	}
	if rec.Status[0] != "Accepted" {
		t.Errorf("Status[0] = %q, want %q", rec.Status[0], "Accepted")
	}
}

func TestExtractLeadingNumber(t *testing.T) {
	tests := []struct {
		filename string
		want     int
	}{
		{"0001-foo.md", 1},
		{"0012-bar.md", 12},
		{"0100-baz.md", 100},
		{"nope.md", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := adrlog.ExtractLeadingNumber(tt.filename)
		if got != tt.want {
			t.Errorf("ExtractLeadingNumber(%q) = %d, want %d", tt.filename, got, tt.want)
		}
	}
}
