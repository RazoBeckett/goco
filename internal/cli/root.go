package cli

import (
	"github.com/razobeckett/goco/internal/config"
	"github.com/razobeckett/goco/internal/git"
	"github.com/spf13/cobra"
)

type dependencies struct {
	configLoader *config.Loader
	repo         *git.Repository
}

func NewRootCmd() *cobra.Command {
	deps := dependencies{
		configLoader: config.NewLoader(),
		repo:         git.NewRepository(""),
	}

	cmd := &cobra.Command{
		Use:     "goco",
		Short:   "Generate Conventional Commit messages with AI",
		Long:    "GoCo generates Conventional Commit messages from your git changes using an AI provider, with Fang-powered help, errors, completions, and manpages.",
		Example: "  goco\n  goco generate --provider openai --model gpt-4o\n  goco generate --staged --verbose --custom-instructions \"focus on API changes\"",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddGroup(&cobra.Group{ID: "main", Title: "Main Commands"})

	cmd.AddCommand(newGenerateCmd(deps))

	return cmd
}
