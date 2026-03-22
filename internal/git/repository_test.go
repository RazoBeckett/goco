package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepositoryStagedFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "goco-test-repo-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	initCmd := exec.Command("git", "init")
	initCmd.Dir = dir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, out: %s", err, out)
	}

	stagedPath := filepath.Join(dir, "staged.txt")
	if err := os.WriteFile(stagedPath, []byte("staged"), 0o644); err != nil {
		t.Fatalf("write staged file: %v", err)
	}

	addCmd := exec.Command("git", "add", "staged.txt")
	addCmd.Dir = dir
	if out, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v, out: %s", err, out)
	}

	unstagedPath := filepath.Join(dir, "unstaged.txt")
	if err := os.WriteFile(unstagedPath, []byte("unstaged"), 0o644); err != nil {
		t.Fatalf("write unstaged file: %v", err)
	}

	repo := NewRepository(dir)
	files, err := repo.StagedFiles(context.Background())
	if err != nil {
		t.Fatalf("StagedFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 staged file, got %d: %v", len(files), files)
	}

	if files[0] != "staged.txt" {
		t.Fatalf("unexpected staged file: %s", files[0])
	}
}
