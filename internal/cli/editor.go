package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func editCommitMessage(message string) (string, error) {
	tmpFile, err := os.CreateTemp("", "goco-commit-*.txt")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(message); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	editor, err := resolveEditor()
	if err != nil {
		return "", err
	}

	editCmd := exec.Command(editor, tmpPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return "", fmt.Errorf("run editor %q: %w", editor, err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("read edited message: %w", err)
	}

	trimmed := strings.TrimSpace(string(edited))
	if trimmed == "" {
		return message, nil
	}

	return trimmed, nil
}

func resolveEditor() (string, error) {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor, nil
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor, nil
	}

	for _, editor := range []string{"vim", "nano", "vi"} {
		if _, err := exec.LookPath(editor); err == nil {
			return editor, nil
		}
	}

	return "", fmt.Errorf("no text editor available; set EDITOR or VISUAL")
}
