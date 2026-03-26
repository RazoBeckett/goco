package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	promptTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(electricOrange)).
				Bold(true)

	promptDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(mangoVolt)).
				Italic(true)

	promptErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(lipstickRed))
)

type apiKeyPromptDoneMsg struct{}

type apiKeyPromptModel struct {
	input       textinput.Model
	title       string
	description string
	err         error
	submitted   bool
}

func newAPIKeyPromptModel(providerName, envVar string) apiKeyPromptModel {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "Paste API key"
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(electricOrange))
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(creamGleam))
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(mangoVolt))
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(electricOrange))

	return apiKeyPromptModel{
		input:       input,
		title:       fmt.Sprintf("Enter your %s API key", providerName),
		description: fmt.Sprintf("This sets %s for the current session only.", envVar),
	}
}

func (m apiKeyPromptModel) Init() tea.Cmd {
	return m.input.Focus()
}

func (m apiKeyPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			if value == "" {
				m.err = fmt.Errorf("API key cannot be empty")
				return m, nil
			}
			m.submitted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Err != nil {
		m.err = m.input.Err
	}
	return m, cmd
}

func (m apiKeyPromptModel) View() string {
	var parts []string
	parts = append(parts, promptTitleStyle.Render(m.title))
	parts = append(parts, promptDescriptionStyle.Render(m.description))
	parts = append(parts, m.input.View())
	if m.err != nil {
		parts = append(parts, promptErrorStyle.Render(m.err.Error()))
	}
	return strings.Join(parts, "\n")
}

func runAPIKeyPrompt(providerName, envVar string) (string, error) {
	program := tea.NewProgram(newAPIKeyPromptModel(providerName, envVar))
	model, err := program.Run()
	if err != nil {
		return "", err
	}

	prompt, ok := model.(apiKeyPromptModel)
	if !ok || !prompt.submitted {
		return "", tea.ErrProgramKilled
	}

	return strings.TrimSpace(prompt.input.Value()), nil
}

type confirmPromptModel struct {
	help      help.Model
	keys      confirmPromptKeyMap
	title     string
	selected  int
	submitted bool
}

type confirmPromptKeyMap struct {
	Left   key.Binding
	Right  key.Binding
	Submit key.Binding
	Choose key.Binding
}

func (k confirmPromptKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Submit}
}

func (k confirmPromptKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Left, k.Right, k.Submit}}
}

func newConfirmPromptModel(title string) confirmPromptModel {
	keys := confirmPromptKeyMap{
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↵", "confirm"),
		),
		Choose: key.NewBinding(
			key.WithKeys("y", "n"),
			key.WithHelp("y/n", "choose"),
		),
	}

	h := help.New()
	h.Styles.ShortKey = promptDescriptionStyle
	h.Styles.ShortDesc = promptDescriptionStyle
	h.Styles.ShortSeparator = promptDescriptionStyle

	return confirmPromptModel{
		title: title,
		keys:  keys,
		help:  h,
	}
}

func (m confirmPromptModel) Init() tea.Cmd {
	return nil
}

func (m confirmPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Left):
			m.selected = 0
		case key.Matches(msg, m.keys.Right):
			m.selected = 1
		case key.Matches(msg, m.keys.Submit):
			m.submitted = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Choose):
			if msg.String() == "y" {
				m.selected = 0
			} else {
				m.selected = 1
			}
			m.submitted = true
			return m, tea.Quit
		case msg.String() == "ctrl+c" || msg.String() == "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m confirmPromptModel) View() string {
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(creamGleam)).
		Background(lipgloss.Color(electricOrange)).
		Bold(true).
		Padding(0, 2)
	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(mangoVolt)).
		Padding(0, 2)

	yesStyle := unselectedStyle
	noStyle := unselectedStyle
	if m.selected == 0 {
		yesStyle = selectedStyle
	} else {
		noStyle = selectedStyle
	}

	return strings.Join([]string{
		promptTitleStyle.Render(m.title),
		lipgloss.JoinHorizontal(lipgloss.Left, yesStyle.Render("Yes"), "  ", noStyle.Render("No")),
		m.help.ShortHelpView(m.keys.ShortHelp()),
	}, "\n")
}

func runConfirmPrompt(title string) (bool, error) {
	program := tea.NewProgram(newConfirmPromptModel(title))
	model, err := program.Run()
	if err != nil {
		return false, err
	}

	prompt, ok := model.(confirmPromptModel)
	if !ok || !prompt.submitted {
		return false, nil
	}

	return prompt.selected == 0, nil
}
