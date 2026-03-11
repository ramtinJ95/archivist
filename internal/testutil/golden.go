package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func GoldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("testdata", name+".golden")
}

func UpdateGolden(t *testing.T, path string, actual []byte) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, actual, 0o644); err != nil {
		t.Fatal(err)
	}
}

func ReadGolden(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
