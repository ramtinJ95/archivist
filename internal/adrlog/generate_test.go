package adrlog_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestGenerateTOC(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	toc, err := repo.GenerateTOC(adrlog.TOCOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(toc, "# Architecture Decision Records\n") {
		t.Errorf("TOC missing heading:\n%s", toc)
	}
	if !strings.Contains(toc, "* [1. Record architecture decisions](0001-record-architecture-decisions.md)") {
		t.Errorf("TOC missing first entry:\n%s", toc)
	}
	if !strings.Contains(toc, "* [2. Use Go for implementation](0002-use-go-for-implementation.md)") {
		t.Errorf("TOC missing second entry:\n%s", toc)
	}
}

func TestGenerateTOCWithIntroOutro(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	toc, err := repo.GenerateTOC(adrlog.TOCOptions{
		Intro: "This is the intro.",
		Outro: "This is the outro.",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(toc, "This is the intro.") {
		t.Errorf("TOC missing intro:\n%s", toc)
	}
	if !strings.Contains(toc, "This is the outro.") {
		t.Errorf("TOC missing outro:\n%s", toc)
	}
}

func TestGenerateTOCWithLinkPrefix(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	toc, err := repo.GenerateTOC(adrlog.TOCOptions{LinkPrefix: "doc/adr/"})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(toc, "(doc/adr/0001") {
		t.Errorf("TOC missing link prefix:\n%s", toc)
	}
}

func TestGenerateGraph(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-record-architecture-decisions.md", testutil.SampleADR1)
	testutil.SeedADR(t, adrDir, "0002-use-go-for-implementation.md", testutil.SampleADR2)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	graph, err := repo.GenerateGraph(adrlog.GraphOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(graph, "digraph {\n") {
		t.Errorf("graph missing digraph opener:\n%s", graph)
	}
	if !strings.Contains(graph, "node [shape=plaintext]") {
		t.Errorf("graph missing node shape:\n%s", graph)
	}
	if !strings.Contains(graph, ".html") {
		t.Errorf("graph missing default .html extension:\n%s", graph)
	}
	if !strings.Contains(graph, "style=\"dotted\"") {
		t.Errorf("graph missing dotted chronological edges:\n%s", graph)
	}
	if !strings.HasSuffix(graph, "}\n") {
		t.Errorf("graph missing closing brace:\n%s", graph)
	}
}

func TestGenerateGraphWithOptions(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	testutil.SeedADR(t, adrDir, "0001-first.md", testutil.SampleADR1)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	graph, err := repo.GenerateGraph(adrlog.GraphOptions{
		LinkPrefix:    "http://example.com/",
		LinkExtension: ".pdf",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(graph, "http://example.com/") {
		t.Errorf("graph missing link prefix:\n%s", graph)
	}
	if !strings.Contains(graph, ".pdf") {
		t.Errorf("graph missing custom extension:\n%s", graph)
	}
}

func TestGenerateGraphExcludesReverseLinks(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	supercededADR := `# 1. First decision

Date: 2024-01-15

## Status

Superceded by [2. Second decision](0002-second-decision.md)

## Context

## Decision

## Consequences
`
	supercedingADR := `# 2. Second decision

Date: 2024-01-16

## Status

Accepted

Supercedes [1. First decision](0001-first-decision.md)

## Context

## Decision

## Consequences
`
	testutil.SeedADR(t, adrDir, "0001-first-decision.md", supercededADR)
	testutil.SeedADR(t, adrDir, "0002-second-decision.md", supercedingADR)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	graph, err := repo.GenerateGraph(adrlog.GraphOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(graph, "\"Superceded by\"") {
		t.Errorf("graph should exclude reverse links ending in ' by':\n%s", graph)
	}
	if !strings.Contains(graph, "\"Supercedes\"") {
		t.Errorf("graph should include forward relation:\n%s", graph)
	}
}

func TestGenerateGraphEscapesQuotedTitlesAndLabels(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")
	adrDir := filepath.Join(dir, "doc", "adr")

	sourceADR := `# 1. Source decision

Date: 2024-01-15

## Status

Clarifies "why" [2. Uses "quoted" title](0002-uses-quoted-title.md)

## Context

## Decision

## Consequences
`
	targetADR := `# 2. Uses "quoted" title

Date: 2024-01-16

## Status

Accepted

## Context

## Decision

## Consequences
`
	testutil.SeedADR(t, adrDir, "0001-source-decision.md", sourceADR)
	testutil.SeedADR(t, adrDir, "0002-uses-quoted-title.md", targetADR)

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	graph, err := repo.GenerateGraph(adrlog.GraphOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(graph, `label="2. Uses \"quoted\" title"`) {
		t.Fatalf("graph should escape quoted node labels:\n%s", graph)
	}
	if !strings.Contains(graph, `label="Clarifies \"why\""`) {
		t.Fatalf("graph should escape quoted edge labels:\n%s", graph)
	}
}
