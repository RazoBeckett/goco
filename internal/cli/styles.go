package cli

import (
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func terminalWidth() int {
	if s := os.Getenv("COLUMNS"); s != "" {
		if w, err := strconv.Atoi(s); err == nil && w > 0 {
			if w > 120 {
				return 120
			}
			return w
		}
	}
	return 80
}

func boxStyle(borderColor, textColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Foreground(lipgloss.Color(textColor)).
		Padding(1).
		MarginBottom(1).
		Width(terminalWidth())
}

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

	statusBoxStyle = boxStyle(electricOrange, tangerineShock)
	diffBoxStyle   = boxStyle(tangerineShock, electricOrange)

	commitMessageHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(creamGleam)).
					Bold(true).
					Background(lipgloss.Color(electricOrange)).
					Padding(0, 1)

	commitMessageBoxStyle = boxStyle(mangoVolt, electricOrange)

	modelProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(electricOrange)).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	modelItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tangerineShock)).
			PaddingLeft(2)
)
