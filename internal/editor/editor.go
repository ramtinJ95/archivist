package editor

import (
	"fmt"
	"os"
	"os/exec"
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

	if editorCmd == "true" {
		return nil
	}

	cmd := exec.Command(editorCmd, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", editorCmd, err)
	}
	return nil
}
