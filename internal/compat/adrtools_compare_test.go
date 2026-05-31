package compat_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type compatStep struct {
	name string
	args []string
}

type compatScenario struct {
	name         string
	initialFiles map[string]string
	steps        []compatStep
}

type commandResult struct {
	stdout string
	stderr string
	status int
}

func TestADRToolsCompatibility(t *testing.T) {
	adrTools := adrToolsCommand(t)
	archivist := buildArchivist(t)

	scenarios := []compatScenario{
		{
			name: "init",
			steps: []compatStep{
				{name: "init", args: []string{"init"}},
			},
		},
		{
			name: "new-list",
			steps: []compatStep{
				{name: "new first", args: []string{"new", "First", "Decision"}},
				{name: "new second", args: []string{"new", "Second", "Decision"}},
				{name: "list", args: []string{"list"}},
			},
		},
		{
			name: "link",
			steps: []compatStep{
				{name: "new first", args: []string{"new", "First", "Record"}},
				{name: "new second", args: []string{"new", "Second", "Record"}},
				{name: "new third", args: []string{"new", "Third", "Record"}},
				{name: "link amends", args: []string{"link", "3", "Amends", "1", "Amended by"}},
				{name: "link clarifies", args: []string{"link", "3", "Clarifies", "2", "Clarified by"}},
			},
		},
		{
			name: "inline-link",
			steps: []compatStep{
				{name: "new first", args: []string{"new", "First", "Record"}},
				{name: "new second", args: []string{"new", "Second", "Record"}},
				{name: "new linked third", args: []string{"new", "-l", "1:Amends:Amended by", "-l", "2:Clarifies:Clarified by", "Third", "Record"}},
			},
		},
		{
			name: "supersede",
			steps: []compatStep{
				{name: "init", args: []string{"init"}},
				{name: "new idea", args: []string{"new", "An", "idea", "that", "seems", "good", "at", "the", "time"}},
				{name: "new superseding idea", args: []string{"new", "-s", "2", "A", "better", "idea"}},
			},
		},
		{
			name: "generate-toc",
			steps: []compatStep{
				{name: "new first", args: []string{"new", "First", "Decision"}},
				{name: "new second", args: []string{"new", "Second", "Decision"}},
				{name: "new third", args: []string{"new", "Third", "Decision"}},
				{name: "generate toc", args: []string{"generate", "toc"}},
			},
		},
		{
			name: "generate-toc-options",
			initialFiles: map[string]string{
				"intro.md": "Intro text.\n\nMore intro.\n",
				"outro.md": "Outro text.\n",
			},
			steps: []compatStep{
				{name: "new first", args: []string{"new", "First", "Decision"}},
				{name: "new second", args: []string{"new", "Second", "Decision"}},
				{name: "generate toc with options", args: []string{"generate", "toc", "-i", "intro.md", "-o", "outro.md", "-p", "docs/adr/"}},
			},
		},
		{
			name: "generate-graph",
			steps: []compatStep{
				{name: "init", args: []string{"init"}},
				{name: "new idea", args: []string{"new", "An", "idea", "that", "seems", "good", "at", "the", "time"}},
				{name: "new superseding idea", args: []string{"new", "-s", "2", "A", "better", "idea"}},
				{name: "new working idea", args: []string{"new", "This", "will", "work"}},
				{name: "new final superseder", args: []string{"new", "-s", "3", "The", "end"}},
				{name: "generate graph", args: []string{"generate", "graph"}},
				{name: "generate graph with options", args: []string{"generate", "graph", "-p", "http://example.com/", "-e", ".xxx"}},
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			tempDir := t.TempDir()
			adrToolsRepo := filepath.Join(tempDir, "adr-tools")
			archivistRepo := filepath.Join(tempDir, "archivist")
			mustMkdir(t, adrToolsRepo)
			mustMkdir(t, archivistRepo)
			writeScenarioFiles(t, adrToolsRepo, scenario.initialFiles)
			writeScenarioFiles(t, archivistRepo, scenario.initialFiles)

			adrToolsTranscript := runScenario(t, adrTools, adrToolsRepo, scenario)
			archivistTranscript := runScenario(t, archivist, archivistRepo, scenario)

			assertEqualText(t, "stdout transcript", adrToolsTranscript.stdout, archivistTranscript.stdout)
			assertEqualText(t, "stderr transcript", adrToolsTranscript.stderr, archivistTranscript.stderr)

			adrToolsSnapshot := repositorySnapshot(t, adrToolsRepo)
			archivistSnapshot := repositorySnapshot(t, archivistRepo)
			assertEqualText(t, "repository snapshot", adrToolsSnapshot, archivistSnapshot)
		})
	}
}

func adrToolsCommand(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("ADR_TOOLS_CMD"); path != "" {
		return path
	}

	if dir := os.Getenv("ADR_TOOLS_DIR"); dir != "" {
		path := filepath.Join(dir, "src", "adr")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("ADR_TOOLS_DIR does not contain src/adr: %v", err)
		}
		return path
	}

	t.Skip("set ADR_TOOLS_DIR to an adr-tools checkout or set ADR_TOOLS_CMD")
	return ""
}

func buildArchivist(t *testing.T) string {
	t.Helper()

	root := repoRoot(t)
	out := filepath.Join(t.TempDir(), "archivist")
	cmd := exec.Command("go", "build", "-o", out, "./cmd/archivist")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build ./cmd/archivist: %v\n%s", err, stderr.String())
	}
	return out
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func runScenario(t *testing.T, commandPath, workDir string, scenario compatScenario) commandResult {
	t.Helper()

	var transcript commandResult
	for _, step := range scenario.steps {
		result := runCommand(t, commandPath, workDir, step.args)
		transcript.stdout += result.stdout
		transcript.stderr += result.stderr
		if result.status != 0 {
			t.Fatalf("%s failed with exit status %d\nstdout:\n%s\nstderr:\n%s", step.name, result.status, result.stdout, result.stderr)
		}
	}
	return transcript
}

func runCommand(t *testing.T, commandPath, workDir string, args []string) commandResult {
	t.Helper()

	cmd := exec.Command(commandPath, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"ADR_DATE=1992-01-12",
		"VISUAL=true",
		"EDITOR=true",
		"LC_ALL=C",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	status := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status = exitErr.ExitCode()
		} else {
			t.Fatalf("run %s %s: %v", commandPath, strings.Join(args, " "), err)
		}
	}

	return commandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		status: status,
	}
}

func repositorySnapshot(t *testing.T, root string) string {
	t.Helper()

	type fileSnapshot struct {
		path    string
		content string
	}

	var files []fileSnapshot
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, fileSnapshot{
			path:    filepath.ToSlash(rel),
			content: string(data),
		})
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })

	var snapshot strings.Builder
	for _, file := range files {
		snapshot.WriteString("--- ")
		snapshot.WriteString(file.path)
		snapshot.WriteString("\n")
		snapshot.WriteString(file.content)
		if !strings.HasSuffix(file.content, "\n") {
			snapshot.WriteString("\n")
		}
	}
	return snapshot.String()
}

func assertEqualText(t *testing.T, label, want, got string) {
	t.Helper()
	if want == got {
		return
	}
	t.Fatalf("%s mismatch (-adr-tools +archivist):\n%s", label, unifiedTextDiff(want, got))
}

func unifiedTextDiff(want, got string) string {
	wantLines := strings.SplitAfter(want, "\n")
	gotLines := strings.SplitAfter(got, "\n")

	var diff strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}

	for i := 0; i < max; i++ {
		var wantLine, gotLine string
		if i < len(wantLines) {
			wantLine = wantLines[i]
		}
		if i < len(gotLines) {
			gotLine = gotLines[i]
		}
		if wantLine == gotLine {
			continue
		}
		diff.WriteString(fmt.Sprintf("@@ line %d @@\n", i+1))
		if wantLine != "" {
			diff.WriteString("-")
			diff.WriteString(wantLine)
			if !strings.HasSuffix(wantLine, "\n") {
				diff.WriteString("\n")
			}
		}
		if gotLine != "" {
			diff.WriteString("+")
			diff.WriteString(gotLine)
			if !strings.HasSuffix(gotLine, "\n") {
				diff.WriteString("\n")
			}
		}
	}

	return diff.String()
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeScenarioFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}
