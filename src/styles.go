package src

import "github.com/charmbracelet/lipgloss"

// Index Styles
var TitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#020202ff")).
	Background(lipgloss.Color("#f9a318ff")).
	MarginLeft(5).
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

// Containers Styles
var ContainerTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#020202ff")).
	Background(lipgloss.Color("#f9a318ff")).
	MarginLeft(1).
	Padding(1).
	Width(70)

var ContainerContentStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#020202ff")).
	Background(lipgloss.Color("#6bc6ffff")).
	Padding(1).
	Width(70).
	MarginLeft(1)
