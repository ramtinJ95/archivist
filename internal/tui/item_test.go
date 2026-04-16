package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
)

func TestADRItemFilterValueIncludesContent(t *testing.T) {
	item := ADRItem{record: &adrlog.Record{
		Number:  2,
		Title:   "Use Go for implementation",
		Date:    "2024-01-16",
		Status:  []string{"Accepted"},
		Path:    filepath.Join("doc", "adr", "0002-use-go-for-implementation.md"),
		Content: "## Decision\n\nWe will use Go.",
	}}

	filterValue := item.FilterValue()

	if !strings.Contains(filterValue, "Use Go for implementation") {
		t.Fatalf("expected title in filter value, got %q", filterValue)
	}
	if !strings.Contains(filterValue, "0002-use-go-for-implementation.md") {
		t.Fatalf("expected filename in filter value, got %q", filterValue)
	}
	if !strings.Contains(filterValue, "We will use Go.") {
		t.Fatalf("expected ADR content in filter value, got %q", filterValue)
	}
}
