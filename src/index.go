package src

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/go-sdk/client"
)

// index Model

type indexModel struct {
	help         help.Model
	keys         keyMap
	dockerClient client.SDKClient
	width        int
	height       int
	table        table.Model
}

func InitIndexModel(dockerClient client.SDKClient) indexModel {
	columns := []table.Column{
		{Title: "Index", Width: 6},
		{Title: "Menu Item", Width: 15},
	}

	rows := []table.Row{
		{"0", "Containers"},
		{"1", "Images"},
		{"2", "Exit"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return indexModel{
		help:         help.New(),
		keys:         keys,
		dockerClient: dockerClient,
		table:        t,
	}
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	MarginLeft(4)

func (m indexModel) Init() tea.Cmd {
	return tea.SetWindowTitle("StarDocker")
}

func (m indexModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle help
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}

		switch msg.String() {

		case "q":
			return m, tea.Quit

		case "enter":
			row := m.table.SelectedRow()

			switch row[1] {

			case "Exit":
				return m, tea.Quit

			case "Containers":
				l := InitListContainersModel(m.dockerClient, m.width, m.height)
				return l, l.Init()

			case "Images":
				m := InitListImagesModel(m.dockerClient)
				return m, m.Init()
			}
		}
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m indexModel) View() string {
	doc := strings.Builder{}

	bigTitle := `
	███████╗████████╗ █████╗ ██████╗ ██████╗  ██████╗  ██████╗██╗  ██╗███████╗██████╗ 
	██╔════╝╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██╔═══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗
	███████╗   ██║   ███████║██████╔╝██║  ██║██║   ██║██║     █████╔╝ █████╗  ██████╔╝
	╚════██║   ██║   ██╔══██║██╔══██╗██║  ██║██║   ██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗
	███████║   ██║   ██║  ██║██║  ██║██████╔╝╚██████╔╝╚██████╗██║  ██╗███████╗██║  ██║
	╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
                                                                                  
	`

	// Some nice fonts
	// Terrace
	// Rubifont
	// ANSI Compact
	// Ansi Regular
	// ANSI Shadow
	// Delta corps preist 1
	// ASCII 12
	// Mono 12

	title := lipgloss.PlaceHorizontal(m.width, lipgloss.Left, TitleStyle.Render(bigTitle))

	doc.WriteString(title)

	doc.WriteString("\n\n")

	doc.WriteString(baseStyle.Render(m.table.View()) + "\n")

	helpView := m.help.View(m.keys)

	doc.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Left, HelpStyle.Render(helpView)))

	return doc.String()
}
