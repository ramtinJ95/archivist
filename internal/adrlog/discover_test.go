package adrlog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/testutil"
)

func TestDiscoverDefaultDocADR(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	got, err := adrlog.DiscoverADRDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "doc/adr" {
		t.Errorf("got %q, want %q", got, "doc/adr")
	}
}

func TestDiscoverDotADRDir(t *testing.T) {
	dir := testutil.TempRepoWithDotADRDir(t, "architecture-log")

	got, err := adrlog.DiscoverADRDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "architecture-log" {
		t.Errorf("got %q, want %q", got, "architecture-log")
	}
}

func TestDiscoverFromNestedSubdirectory(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	nested := filepath.Join(dir, "services", "foo")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := adrlog.DiscoverADRDir(nested)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("..", "..", "doc", "adr")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiscoverDotADRDirFromNestedSubdirectory(t *testing.T) {
	dir := testutil.TempRepoWithDotADRDir(t, "decisions")

	nested := filepath.Join(dir, "src", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := adrlog.DiscoverADRDir(nested)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("..", "..", "decisions")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiscoverFallbackToDefault(t *testing.T) {
	dir := t.TempDir()

	got, err := adrlog.DiscoverADRDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "doc/adr" {
		t.Errorf("got %q, want %q", got, "doc/adr")
	}
}

func TestDotADRDirTakesPrecedenceOverDocADR(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	customDir := filepath.Join(dir, "my-adrs")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dotFile := filepath.Join(dir, ".adr-dir")
	if err := os.WriteFile(dotFile, []byte("my-adrs"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := adrlog.DiscoverADRDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "my-adrs" {
		t.Errorf("got %q, want %q", got, "my-adrs")
	}
}

func TestOpenRepository(t *testing.T) {
	dir := testutil.TempRepoWithADRDir(t, "doc/adr")

	repo, err := adrlog.OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	if repo.ADRDir != "doc/adr" {
		t.Errorf("ADRDir = %q, want %q", repo.ADRDir, "doc/adr")
	}
	if repo.CWD != dir {
		t.Errorf("CWD = %q, want %q", repo.CWD, dir)
	}
}
