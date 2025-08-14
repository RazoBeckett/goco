package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var (
	apiKey         string
	model          string
	commitType     string
	breakingChange bool
	stagged        bool
	verbose        bool
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			MarginBottom(1)

	noteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			MarginTop(1)

	statusHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true).
				Background(lipgloss.Color("#065F46")).
				Padding(0, 1)

	diffHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true).
			Background(lipgloss.Color("#1E3A8A")).
			Padding(0, 1)

	statusBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Padding(1).
			MarginBottom(1).
			Width(80)

	diffBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B82F6")).
			Padding(1).
			MarginBottom(1).
			Width(80)

	commitMessageHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).
					Background(lipgloss.Color("#059669")).
					Padding(0, 1)

	commitMessageBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#10B981")).
				Padding(1).
				MarginBottom(1).
				Width(80)
)

type spinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case string:
		if msg == "done" {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

func newSpinnerModel(message string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	return spinnerModel{
		spinner: s,
		message: message,
	}
}

func promptForApiKey(envVar string) (string, error) {
	var apiKey string

	fmt.Println(titleStyle.Render("🔑 Gemini API Key Required"))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Please enter your Gemini API key:").
				Description("Your key will be set for this session only").
				Value(&apiKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("API key cannot be empty")
					}
					return nil
				}),
		),
	)

	err := form.Run()
	if err != nil {
		return "", err
	}

	// Set for current session
	os.Setenv(envVar, apiKey)

	// Show helpful note
	note := fmt.Sprintf(`
%s

To avoid this prompt in the future, add this to your shell profile:
  export %s="your-api-key-here"
  
For bash: ~/.bashrc or ~/.bash_profile
For zsh: ~/.zshrc
For fish: ~/.config/fish/config.fish`,
		noteStyle.Render("✅ API key set for this session!"), envVar)

	fmt.Println(note)
	fmt.Println()

	return apiKey, nil
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message using Gemini",

	Run: func(cmd *cobra.Command, args []string) {
		// Use flag value if provided, otherwise get from config
		if apiKey == "" {
			apiKey = GetConfig().GetGeminiApiKey()
		}

		// If still no API key, prompt user interactively
		if apiKey == "" {
			envVar := GetConfig().General.ApiKeyGeminiEnvVariable
			if envVar == "" {
				envVar = "GOCO_GEMINI_KEY"
			}

			promptedKey, err := promptForApiKey(envVar)
			if err != nil {
				log.Fatalf("Failed to get API key: %v", err)
			}
			apiKey = promptedKey
		}

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

		prompt := fmt.Sprintf(
			"Generate a Conventional Commit based strictly on the following:\n\n"+
				"Git Status:\n%s\n\n"+
				"Git Diff:\n%s\n\n"+
				"Before responding, you MUST:\n"+
				"- Read: %v\n"+
				"- ONLY output the commit message and description.\n"+
				"- DO NOT include markdown, code blocks, quotes, or any formatting.\n"+
				"- Output MUST be plain text only.\n"+
				"- Do not add extra explanations, notes, or commentary.\n"+
				"- The first line is the commit summary, the rest is the description.\n"+
				"- Follow Conventional Commit standards exactly.\n"+
				"- No extra lines before or after the commit message.\n",
			gitStatusOutput,
			gitDiffOutput,
			referLink,
		)

		if verbose {
			// Show git status in a green box
			statusBox := statusBoxStyle.Render(string(gitStatusOutput))
			fmt.Println(statusHeaderStyle.Render("📊 Git Status"))
			fmt.Println(statusBox)

			// Show git diff in a blue box
			diffBox := diffBoxStyle.Render(string(gitDiffOutput))
			fmt.Println(diffHeaderStyle.Render("📝 Git Diff"))
			fmt.Println(diffBox)
		}

		// Start spinner during API call
		spinnerProgram := tea.NewProgram(newSpinnerModel("Generating commit message..."))

		// Run spinner in goroutine
		done := make(chan bool)
		go func() {
			spinnerProgram.Run()
			done <- true
		}()

		// Make API call
		resp, err := client.Models.GenerateContent(
			ctx,
			model,
			genai.Text(prompt),
			nil,
		)

		// Stop spinner
		spinnerProgram.Send("done")
		spinnerProgram.Quit()
		<-done // Wait for spinner to finish

		if err != nil {
			log.Fatalf("Gemini API error: %v", err)
		}

		commitMessage := resp.Text()

		// Show the commit message in a beautiful green box
		fmt.Println(commitMessageHeaderStyle.Render("✅ Generated Commit Message"))
		fmt.Println(commitMessageBoxStyle.Render(commitMessage))

		if err := exec.Command("git", "add", "-u").Run(); err != nil {
			log.Fatalf("Failed to stage changes %v", err)
		}

		final := exec.Command("git", "commit", "-m", commitMessage)
		final.Stdout = os.Stdout
		final.Stderr = os.Stderr

		if err := final.Run(); err != nil {
			log.Fatalf("Failed to commit changes %v", err)
		}

	},
}

func init() {
	generateCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "Gemini API key")
	generateCmd.Flags().StringVarP(&model, "model", "m", "gemini-2.5-flash", "Gemini model to use")
	generateCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, chore, etc.)")
	generateCmd.Flags().BoolVarP(&breakingChange, "breaking-change", "b", false, "Mark commit as breaking change")
	generateCmd.Flags().BoolVarP(&stagged, "stagged", "s", false, "stagged changes")
	generateCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed output including prompts")

	rootCmd.AddCommand(generateCmd)
}
