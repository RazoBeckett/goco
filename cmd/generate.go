package cmd

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var (
	apiKey         string
	model          string
	commitType     string
	breakingChange bool
	stagged        bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message using Gemini",

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			log.Fatalf("failed to create genai client: %v", err)
		}

		models, err := client.Models.List(ctx, nil)

		if err != nil {
			log.Fatalf("Failed to list model: %v", err)
		}

		var filtered []string
		re := regexp.MustCompile(`^gemini-\d+\.\d+`)
		for _, m := range models.Items {
			name := strings.TrimPrefix(m.Name, "models/")
			if re.MatchString(name) {
				filtered = append(filtered, name)
			}
		}

		if !slices.Contains(filtered, model) {
			var b strings.Builder
			for _, m := range filtered {
				fmt.Fprintf(&b, "%s\n", m)
			}
			log.Fatalf("Model not available\nAvailable Models: \n%s", b.String())
		}

		gitStatus := exec.Command("git", "status")

		gitStatusOutput, err := gitStatus.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var gitDiff *exec.Cmd

		if stagged {
			gitDiff = exec.Command("git", "diff", "--no-color", "--staged")

		} else {
			gitDiff = exec.Command("git", "diff", "--no-color")
		}

		gitDiffOutput, err := gitDiff.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		referLink := "https://gist.githubusercontent.com/qoomon/5dfcdf8eec66a051ecd85625518cfd13/raw/d7d529a329079616d47dcf100bd7d2d2c848e835/conventional-commits-cheatsheet.md"

		prompt := fmt.Sprintf("Generate Conventional Commit:\n\nGit Status: %s\n\nGit Diff: %s\n\nThings to do before resposeding, you won't responed anything rather than the commit message and commit description that's all i want, and make sure you read: %v", gitStatusOutput, gitDiffOutput, referLink)

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
	generateCmd.Flags().StringVarP(&model, "model", "m", "gemini-2.5-flash", "Gemini model to use")
	generateCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, chore, etc.)")
	generateCmd.Flags().BoolVarP(&breakingChange, "breaking-change", "b", false, "Mark commit as breaking change")
	generateCmd.Flags().BoolVarP(&stagged, "stagged", "s", false, "stagged changes")

	generateCmd.MarkFlagRequired("api-key")
	rootCmd.AddCommand(generateCmd)
}
