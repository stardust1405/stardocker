package src

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/go-sdk/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tickMsg time.Time

// List Containers Model

type listContainersModel struct {
	cursor                int
	nestedCursor1         int
	nestedCursorActivated bool
	containers            []Container
	help                  help.Model
	keys                  keyMap
	dockerClient          client.SDKClient
}

const composeStackIdentifier = "com.docker.compose.project"

type ContainerType string

const (
	TypeContainer    ContainerType = "container"
	TypeComposeStack ContainerType = "compose_stack"
)

type Container struct {
	ID       string
	Name     string
	Type     ContainerType
	Status   string
	State    containerTypes.ContainerState
	Children []Container
}

func InitListContainersModel(dockerClient client.SDKClient) listContainersModel {
	allContainers := FetchContainers(dockerClient)

	return listContainersModel{
		containers:   allContainers,
		help:         help.New(),
		keys:         keys,
		dockerClient: dockerClient,
	}
}

func (l listContainersModel) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("List Containers"), tickCmd())
}

func (l listContainersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle help
		switch {
		case key.Matches(msg, l.keys.Help):
			l.help.ShowAll = !l.help.ShowAll
		}

		switch msg.String() {
		case "esc":
			m := InitIndexModel(l.dockerClient)
			return m, m.Init()

		case "up", "k":
			if l.nestedCursorActivated {
				if l.nestedCursor1 > 0 {
					l.nestedCursor1--
				}
			} else {
				if l.cursor > 0 {
					l.cursor--
				}
			}
		case "down", "j":
			if l.nestedCursorActivated {
				if l.nestedCursor1 < len(l.containers[l.cursor].Children)-1 {
					l.nestedCursor1++
				}
			} else {
				if l.cursor < len(l.containers)-1 {
					l.cursor++
				}
			}
		case "right":
			if l.containers[l.cursor].Type == TypeComposeStack {
				l.nestedCursorActivated = true
				if l.nestedCursor1 >= len(l.containers[l.cursor].Children) {
					l.nestedCursor1 = 0
				}
			}
		case "left":
			if l.nestedCursorActivated {
				l.nestedCursorActivated = false
			}
		case "enter":
			// return InitViewContainersModel(l.DB, l.Containers[l.cursor], l.ContainersDownloadPercent[l.cursor], "/home/stardust/Downloads"), nil
		case "r":
			l.containers = FetchContainers(l.dockerClient)
		case "s":
			selectedContainer := l.containers[l.cursor]

			if selectedContainer.Type == TypeContainer {
				StartContainer(l.dockerClient, selectedContainer.ID)
			}

			if selectedContainer.Type == TypeComposeStack {
				if l.nestedCursorActivated {
					selectedNestedContainer := selectedContainer.Children[l.nestedCursor1]
					StartContainer(l.dockerClient, selectedNestedContainer.ID)
				}
			}
		case "d":
			selectedContainer := l.containers[l.cursor]

			if selectedContainer.Type == TypeContainer {
				StopContainer(l.dockerClient, selectedContainer.ID)
			}

			if selectedContainer.Type == TypeComposeStack {
				if l.nestedCursorActivated {
					selectedNestedContainer := selectedContainer.Children[l.nestedCursor1]
					StopContainer(l.dockerClient, selectedNestedContainer.ID)
				}
			}
		}
	case tickMsg:
		l.containers = FetchContainers(l.dockerClient)
		return l, tickCmd()
	}
	return l, nil
}

func (l listContainersModel) View() string {
	s := "\nContainers\n\n"

	for i, container := range l.containers {
		cursor := "  "
		if l.cursor == i {
			cursor = "🐬"
		}

		s += fmt.Sprintf("%s %d. %s (%s)", cursor, i+1, container.Name, container.Type)
		if container.Type == TypeComposeStack {
			s += fmt.Sprintf(" (%d)", len(container.Children))
			for _, child := range container.Children {
				if child.State == containerTypes.StateRunning {
					s += " 🟢"
					break
				}
			}
		} else {
			s += fmt.Sprintf("   [%s]", container.Status)
			if container.State == containerTypes.StateRunning {
				s += " 🟢"
			}
		}
		s += "\n"

		if l.cursor == i {
			if container.Type == TypeComposeStack {
				for j, child := range container.Children {
					nestedCursor1 := "   "
					if l.nestedCursor1 == j && l.nestedCursorActivated {
						nestedCursor1 = "⛵️"
					}
					s += fmt.Sprintf("      %s%s   [%s]", nestedCursor1, child.Name, child.Status)
					if child.State == containerTypes.StateRunning {
						s += " 🟢"
					}
					s += "\n"
				}
			}
		}
		s += "\n"
	}

	helpView := l.help.View(l.keys)
	height := 8 - strings.Count(helpView, "\n")

	s += strings.Repeat("\n", height) + helpView

	return s
}

func FetchContainers(dockerClient client.SDKClient) []Container {
	ctx := context.TODO()

	containers, err := dockerClient.ContainerList(ctx, containerTypes.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	// List to store all containers
	allContainers := make([]Container, 0)

	// Map to store compose containers
	composeContainers := make(map[string][]Container)

	for _, container := range containers {
		// Check if container is part of a compose stack
		if _, ok := container.Labels[composeStackIdentifier]; ok {
			composeProjectName := container.Labels[composeStackIdentifier]
			composeContainers[composeProjectName] = append(composeContainers[composeProjectName], Container{
				ID:     container.ID,
				Name:   strings.TrimLeft(container.Names[0], "/"),
				Type:   TypeComposeStack,
				Status: container.Status,
				State:  container.State,
			})
		} else {
			allContainers = append(allContainers, Container{
				ID:     container.ID,
				Name:   strings.TrimLeft(container.Names[0], "/"),
				Type:   TypeContainer,
				Status: container.Status,
				State:  container.State,
			})
		}
	}

	// Add compose containers to all containers
	for composeStackName, containers := range composeContainers {
		composeStack := Container{
			ID:   primitive.NewObjectID().Hex(),
			Name: composeStackName,
			Type: TypeComposeStack,
		}
		sort.Slice(containers, func(i, j int) bool { return containers[i].Name < containers[j].Name })

		composeStack.Children = containers
		allContainers = append(allContainers, composeStack)
	}

	sort.Slice(allContainers, func(i, j int) bool { return allContainers[i].Name < allContainers[j].Name })

	return allContainers
}

func StartContainer(dockerClient client.SDKClient, containerID string) {
	ctx := context.Background()
	err := dockerClient.ContainerStart(ctx, containerID, containerTypes.StartOptions{})
	if err != nil {
		panic(err)
	}
}

func StopContainer(dockerClient client.SDKClient, containerID string) {
	ctx := context.Background()
	err := dockerClient.ContainerStop(ctx, containerID, containerTypes.StopOptions{})
	if err != nil {
		panic(err)
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
