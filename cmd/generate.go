package cmd

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var (
	apiKey         string
	model          string
	commitType     string
	breakingChange bool
)

// Valid Gemini models (update this list as needed)
var validModels = []string{
	"gemini-2.0-flash-exp",
	"gemini-1.5-flash",
	"gemini-1.5-pro",
	"gemini-1.0-pro",
}

func isValidModel(model string) bool {
	for _, valid := range validModels {
		if model == valid {
			return true
		}
	}
	return false
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message using Gemini",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate model name
		if !isValidModel(model) {
			return fmt.Errorf("invalid model '%s'. Valid models are: %s",
				model, strings.Join(validModels, ", "))
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			log.Fatalf("failed to create genai client: %v", err)
		}

		gitStatus := exec.Command("git", "status")

		gitStatusOutput, err := gitStatus.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		gitDiff := exec.Command("git", "diff", "--no-color", "--staged")

		gitDiffOutput, err := gitDiff.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		prompt := fmt.Sprintf("Generate Conventional Commit:\n\nGit Status: %s\n\nGit Diff: %s", gitStatusOutput, gitDiffOutput)

		fmt.Println(prompt)

		resp, err := client.Models.GenerateContent(
			ctx,
			model,
			genai.Text(prompt),
			nil,
		)
		if err != nil {
			log.Fatalf("Gemini API error: %v", err)
		}

		fmt.Println(resp.Text())
	},
}

func init() {
	generateCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "Gemini API key")
	generateCmd.Flags().StringVarP(&model, "model", "m", "gemini-2.0-flash-exp", "Gemini model to use")
	generateCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, chore, etc.)")
	generateCmd.Flags().BoolVarP(&breakingChange, "breaking-change", "b", false, "Mark commit as breaking change")

	generateCmd.MarkFlagRequired("api-key")
	rootCmd.AddCommand(generateCmd)
}
