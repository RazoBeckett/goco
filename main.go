package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/razobeckett/goco/cmd"
)

const (
	exitOK      = 0
	exitError   = 1
	exitCancel  = 2
	exitAuth    = 4
	exitPending = 8
)

func main() {
	if err := cmd.Execute(); err != nil {
		var apiErr *cmd.APIError
		var gitErr *cmd.GitError
		var provErr *cmd.ProviderError
		var valErr *cmd.ValidationError
		var configErr *cmd.ConfigError

		switch {
		case errors.As(err, &valErr):
			fmt.Fprintf(os.Stderr, "Error: %s\n", valErr.Error())
			os.Exit(exitError)

		case errors.As(err, &apiErr):
			fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Error())
			os.Exit(exitError)

		case errors.As(err, &gitErr):
			fmt.Fprintf(os.Stderr, "Error: %s\n", gitErr.Error())
			os.Exit(exitError)

		case errors.As(err, &provErr):
			fmt.Fprintf(os.Stderr, "Error: %s\n", provErr.Error())
			os.Exit(exitError)

		case errors.As(err, &configErr):
			fmt.Fprintf(os.Stderr, "Error: %s\n", configErr.Error())
			os.Exit(exitError)

		case errors.Is(err, cmd.ErrNoEditor):
			fmt.Fprintf(os.Stderr, "Error: %s\n\nPlease set the EDITOR or VISUAL environment variable, or ensure a text editor (vim, nano, vi) is installed.\n", err.Error())
			os.Exit(exitError)

		case errors.Is(err, cmd.ErrGitRepository):
			fmt.Fprintf(os.Stderr, "Error: %s\n\nPlease run this command from within a git repository.\n", err.Error())
			os.Exit(exitError)

		case errors.Is(err, cmd.ErrNoStagedFiles):
			fmt.Fprintf(os.Stderr, "Error: %s\n\nStage your changes using 'git add <files>' before running goco.\n", err.Error())
			os.Exit(exitError)

		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	}
	os.Exit(exitOK)
}
