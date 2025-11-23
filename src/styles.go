package src

import "github.com/charmbracelet/lipgloss"

var TitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#020202ff")).
	Background(lipgloss.Color("#f9a318ff")).
	Padding(2, 3, 0, 0)

var ContentStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#020202ff")).
	Background(lipgloss.Color("#6bc6ffff")).
	Padding(1, 3, 0, 2).
	Width(30).Height(20).
	MarginLeft(5).
	Bold(true)

var HelpStyle = lipgloss.NewStyle().
	MarginLeft(5)
