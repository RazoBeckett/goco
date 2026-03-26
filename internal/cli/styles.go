package cli

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(electricOrange)).
			Bold(true).
			MarginBottom(1)

	noteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tangerineShock)).
			Italic(true).
			MarginTop(1)

	statusHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(creamGleam)).
				Bold(true).
				Background(lipgloss.Color(electricOrange)).
				Padding(0, 1)

	diffHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(creamGleam)).
			Bold(true).
			Background(lipgloss.Color(tangerineShock)).
			Padding(0, 1)

	statusBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(electricOrange)).
			Foreground(lipgloss.Color(tangerineShock)).
			Padding(1).
			MarginBottom(1).
			Width(80)

	diffBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(tangerineShock)).
			Foreground(lipgloss.Color(electricOrange)).
			Padding(1).
			MarginBottom(1).
			Width(80)

	commitMessageHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(creamGleam)).
					Bold(true).
					Background(lipgloss.Color(electricOrange)).
					Padding(0, 1)

	commitMessageBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(mangoVolt)).
				Foreground(lipgloss.Color(electricOrange)).
				Padding(1).
				MarginBottom(1).
				Width(80)

	modelProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(electricOrange)).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	modelItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tangerineShock)).
			PaddingLeft(2)
)
