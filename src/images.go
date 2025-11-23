package src

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/go-sdk/client"
)

// List Images Model

type listImagesModel struct {
	cursor                    int
	containers                []string
	containersDownloadPercent map[int]int
	help                      help.Model
	keys                      keyMap
	dockerClient              client.SDKClient
}

func InitListImagesModel(dockerClient client.SDKClient) listImagesModel {
	list := []string{"Postgres", "Redis", "Kafka", "Star Trek", "Forza Horizon 5"}
	percent := make(map[int]int)
	for i := range list {
		percent[i] = rand.Intn(101)
	}
	return listImagesModel{
		containers:                list,
		containersDownloadPercent: percent,
		help:                      help.New(),
		keys:                      keys,
		dockerClient:              dockerClient,
	}
}

func (l listImagesModel) Init() tea.Cmd {
	return tea.SetWindowTitle("List Images")
}

func (l listImagesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle help
		switch {
		case key.Matches(msg, l.keys.Help):
			l.help.ShowAll = !l.help.ShowAll
		}

		switch msg.String() {
		case "esc":
			return InitHomeModel(l.dockerClient), nil
		case "up", "k":
			if l.cursor > 0 {
				l.cursor--
			}
		case "down", "j":
			if l.cursor < len(l.containers)-1 {
				l.cursor++
			}
		case "enter":
			// return InitViewImagesModel(l.DB, l.Images[l.cursor], l.ImagesDownloadPercent[l.cursor], "/home/stardust/Downloads"), nil
		}
	}
	return l, nil
}

func (l listImagesModel) View() string {
	s := "\nList Images\n\n"

	for i, Images := range l.containers {
		cursor := " "
		if l.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %d. %s (%d%%)\n", cursor, i+1, Images, l.containersDownloadPercent[i])

		status := "Downloading"
		if l.containersDownloadPercent[i] == 100 {
			status = "Downloaded"
		}

		s += "     Progress: "

		prog := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
		s += prog.ViewAs(float64(l.containersDownloadPercent[i]))
		s += "\n"
		s += fmt.Sprintf("     Status: %s\n\n", status)
	}

	helpView := l.help.View(l.keys)
	height := 8 - strings.Count(helpView, "\n")

	s += strings.Repeat("\n", height) + helpView

	return s
}
