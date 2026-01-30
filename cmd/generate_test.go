package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGetStagedFiles creates a temporary git repository, stages a file and
// verifies getStagedFiles returns only the staged path.
func TestGetStagedFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "goco-test-repo-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// init git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, out: %s", err, out)
	}

	// create and stage file staged.txt
	stagedPath := filepath.Join(dir, "staged.txt")
	if err := os.WriteFile(stagedPath, []byte("staged"), 0o644); err != nil {
		t.Fatalf("write staged file: %v", err)
	}
	cmd = exec.Command("git", "add", "staged.txt")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v, out: %s", err, out)
	}

	// create but do not stage unstaged.txt
	unstagedPath := filepath.Join(dir, "unstaged.txt")
	if err := os.WriteFile(unstagedPath, []byte("unstaged"), 0o644); err != nil {
		t.Fatalf("write unstaged file: %v", err)
	}

	// call getStagedFiles
	files, err := getStagedFiles(dir)
	if err != nil {
		t.Fatalf("getStagedFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 staged file, got %d: %v", len(files), files)
	}

	if files[0] != "staged.txt" {
		t.Fatalf("unexpected staged file: %v", files[0])
	}
}
