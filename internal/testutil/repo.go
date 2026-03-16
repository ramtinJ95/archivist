package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TempRepoWithADRDir(t *testing.T, adrDir string) string {
	t.Helper()
	dir := TempRepo(t)

	fullADRPath := filepath.Join(dir, adrDir)
	if err := os.MkdirAll(fullADRPath, 0o755); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TempRepoWithDotADRDir(t *testing.T, adrDir string) string {
	t.Helper()
	dir := TempRepoWithADRDir(t, adrDir)

	dotFile := filepath.Join(dir, ".adr-dir")
	if err := os.WriteFile(dotFile, []byte(adrDir+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
