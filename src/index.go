package src

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/go-sdk/client"
)

// index Model

type indexModel struct {
	cursor           int
	menuItems        []string
	selectedMenuItem string
	help             help.Model
	keys             keyMap
	dockerClient     client.SDKClient
	width            int
	height           int
}

func InitIndexModel(dockerClient client.SDKClient) indexModel {
	return indexModel{
		menuItems:    []string{"List Containers", "List Images", "Exit"},
		help:         help.New(),
		keys:         keys,
		dockerClient: dockerClient,
	}
}

func (m indexModel) Init() tea.Cmd {
	return tea.SetWindowTitle("StarDocker")
}

func (m indexModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}

		case "q":
			return m, tea.Quit

		case "enter":
			m.selectedMenuItem = m.menuItems[m.cursor]
			switch m.selectedMenuItem {

			case "Exit":
				return m, tea.Quit

			case "List Containers":
				m := InitListContainersModel(m.dockerClient, m.width, m.height)
				return m, m.Init()

			case "List Images":
				m := InitListImagesModel(m.dockerClient)
				return m, m.Init()
			}
		}
	}

	return m, nil
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

	// Content Lines
	contentDoc := strings.Builder{}

	for i, item := range m.menuItems {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		contentDoc.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
	}

	content := lipgloss.PlaceHorizontal(m.width, lipgloss.Left, ContentStyle.Render(contentDoc.String()))

	doc.WriteString(content)

	doc.WriteString("\n\n")

	if m.selectedMenuItem == "Exit" {
		doc.WriteString("Bye!")
	}

	doc.WriteString("\n\n")

	helpView := m.help.View(m.keys)

	doc.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Left, HelpStyle.Render(helpView)))

	return doc.String()
}
