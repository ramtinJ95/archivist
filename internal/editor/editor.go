package editor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ResolveEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return ""
}

func LaunchEditor(path string) error {
	editorCmd := ResolveEditor()
	if editorCmd == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", editorCmd+" \"$1\"", "sh", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", editorCmd, err)
	}
	return nil
}

func ResolvePager() string {
	if p := os.Getenv("ADR_PAGER"); p != "" {
		return p
	}
	if p := os.Getenv("PAGER"); p != "" {
		return p
	}
	return ""
}

func LaunchPager(w io.Writer, content string) error {
	pagerCmd := ResolvePager()
	if pagerCmd == "" {
		fmt.Fprint(w, content)
		return nil
	}

	cmd := exec.Command("sh", "-c", pagerCmd)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pager %q: %w", pagerCmd, err)
	}
	return nil
}
