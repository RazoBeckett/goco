package cli

import "github.com/charmbracelet/lipgloss"

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

	modelProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	modelItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			PaddingLeft(2)
)
