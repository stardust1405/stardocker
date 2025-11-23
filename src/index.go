package src

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/go-sdk/client"
)

// Home Model

type homeModel struct {
	cursor           int
	menuItems        []string
	selectedMenuItem string
	help             help.Model
	keys             keyMap
	dockerClient     client.SDKClient
}

func InitHomeModel(dockerClient client.SDKClient) homeModel {
	return homeModel{
		menuItems:    []string{"List Containers", "List Images", "Exit"},
		help:         help.New(),
		keys:         keys,
		dockerClient: dockerClient,
	}
}

func (m homeModel) Init() tea.Cmd {
	return tea.SetWindowTitle("StarDocker")
}

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		case "enter":
			m.selectedMenuItem = m.menuItems[m.cursor]
			switch m.selectedMenuItem {
			case "Exit":
				return m, tea.Quit
			case "List Containers":
				m := InitListContainersModel(m.dockerClient)
				return m, m.Init()
			case "List Images":
				m := InitListImagesModel(m.dockerClient)
				return m, m.Init()
			}
		}
	}

	return m, nil
}

func (m homeModel) View() string {
	s := "\n******** StarDocker welcomes you! **********\n\n"

	for i, item := range m.menuItems {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, item)
	}

	if m.selectedMenuItem == "Exit" {
		s += "\nBye!\n"
	}

	helpView := m.help.View(m.keys)
	height := 8 - strings.Count(helpView, "\n")

	s += strings.Repeat("\n", height) + helpView

	return s
}
