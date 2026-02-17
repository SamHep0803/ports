package tui

import "github.com/charmbracelet/lipgloss"

var (
	paneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1)

	titleStyle = lipgloss.NewStyle().Bold(true)
	faintStyle = lipgloss.NewStyle().Faint(true)
)
