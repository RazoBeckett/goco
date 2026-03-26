package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrNoChanges = errors.New("no changes detected in the repository")

type Repository struct {
	dir string
}

func NewRepository(dir string) *Repository {
	return &Repository{dir: dir}
}

func (r *Repository) Status(ctx context.Context) (string, error) {
	return r.output(ctx, "status", "--short", "--branch")
}

func (r *Repository) Diff(ctx context.Context, staged bool) (string, error) {
	args := []string{"diff", "--no-color"}
	if staged {
		args = append(args, "--staged")
	}
	return r.output(ctx, args...)
}

func (r *Repository) EnsureChanges(ctx context.Context) (string, error) {
	status, err := r.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("read git status: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(status), "\n")
	if len(lines) <= 1 {
		return "", ErrNoChanges
	}
	return status, nil
}

func (r *Repository) CurrentBranch(ctx context.Context) (string, error) {
	out, err := r.output(ctx, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("read current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (r *Repository) CreateBranch(ctx context.Context, name string) error {
	if _, err := r.output(ctx, "checkout", "-b", name); err != nil {
		return fmt.Errorf("create branch %q: %w", name, err)
	}
	return nil
}

func (r *Repository) StageTracked(ctx context.Context) error {
	if _, err := r.output(ctx, "add", "-u"); err != nil {
		return fmt.Errorf("stage tracked changes: %w", err)
	}
	return nil
}

func (r *Repository) StagedFiles(ctx context.Context) ([]string, error) {
	out, err := r.output(ctx, "diff", "--name-only", "--cached")
	if err != nil {
		return nil, fmt.Errorf("list staged files: %w", err)
	}
	files := strings.Fields(strings.TrimSpace(out))
	if len(files) == 0 {
		return nil, ErrNoChanges
	}
	return files, nil
}

func (r *Repository) Commit(ctx context.Context, message string, onlyFiles []string) error {
	args := []string{"commit", "-m", message}
	if len(onlyFiles) > 0 {
		args = append(args, "--only", "--")
		args = append(args, onlyFiles...)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}
	return nil
}

func (r *Repository) output(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, msg)
	}

	return stdout.String(), nil
}
