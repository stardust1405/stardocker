package src

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	width                 int
	height                int
	slectedContainerLogs  string
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
	Logs     string
	Children []Container
}

func InitListContainersModel(dockerClient client.SDKClient, width int, height int) listContainersModel {
	allContainers := FetchContainers(dockerClient)

	return listContainersModel{
		containers:   allContainers,
		help:         help.New(),
		keys:         keys,
		dockerClient: dockerClient,
		width:        width,
		height:       height,
	}
}

func (l listContainersModel) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("List Containers"), tickCmd())
}

func (l listContainersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		return l, nil

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
			selectedContainer := l.containers[l.cursor]
			if selectedContainer.Type == TypeContainer {
				l.slectedContainerLogs = GetContainerLogs(l.dockerClient, selectedContainer.ID)
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
			selectedContainer := l.containers[l.cursor]
			if selectedContainer.Type == TypeContainer {
				l.slectedContainerLogs = GetContainerLogs(l.dockerClient, selectedContainer.ID)
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
		case "q":
			return l, tea.Quit
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
		selectedContainer := l.containers[l.cursor]
		if selectedContainer.Type == TypeContainer {
			l.slectedContainerLogs = GetContainerLogs(l.dockerClient, l.containers[l.cursor].ID)
		}
		l.containers = FetchContainers(l.dockerClient)
		return l, tickCmd()
	}
	return l, nil
}

func (l listContainersModel) View() string {
	docLeft := strings.Builder{}

	title := lipgloss.PlaceHorizontal(l.width, lipgloss.Left, ContainerTitleStyle.Render("LIST CONTAINERS"))

	docLeft.WriteString(title)

	docLeft.WriteString("\n\n")

	contentDoc := strings.Builder{}

	for i, container := range l.containers {
		cursor := " "
		if l.cursor == i {
			cursor = "⏺"
		}

		contentDoc.WriteString(fmt.Sprintf("[%s] %s (%s)", cursor, container.Name, container.Type))
		if container.Type == TypeComposeStack {
			contentDoc.WriteString(fmt.Sprintf(" (%d)", len(container.Children)))
			for _, child := range container.Children {
				if child.State == containerTypes.StateRunning {
					contentDoc.WriteString(" ✅")
					break
				}
			}
		} else {
			contentDoc.WriteString(fmt.Sprintf("   [%s]", container.Status))
			if container.State == containerTypes.StateRunning {
				contentDoc.WriteString(" ✅")
			}
		}
		contentDoc.WriteString("\n")

		if l.cursor == i {
			if container.Type == TypeComposeStack {
				for j, child := range container.Children {
					nestedCursor1 := " "
					if l.nestedCursor1 == j && l.nestedCursorActivated {
						nestedCursor1 = ">"
					}
					contentDoc.WriteString(fmt.Sprintf("      [%s]%s   [%s]", nestedCursor1, child.Name, child.Status))
					if child.State == containerTypes.StateRunning {
						contentDoc.WriteString(" ✅")
					}
					contentDoc.WriteString("\n")
				}
			}
		}
		contentDoc.WriteString("\n")
	}

	content := lipgloss.PlaceHorizontal(l.width, lipgloss.Left, ContainerContentStyle.Render(contentDoc.String()))

	docLeft.WriteString(content)

	docRight := strings.Builder{}

	// logs, err := cli.ContainerLogs(ctx, "7ea3b2f057bbf5f9e5b199e29c7e21008d293b9815d74954dedfa8f50156c683", container.LogsOptions{ShowStdout: true, ShowStderr: true})
	// if err != nil {
	// 	panic(err)
	// }
	// defer logs.Close()

	// data, err := io.ReadAll(logs)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(data))

	docRight.WriteString(l.slectedContainerLogs)

	doc := strings.Builder{}

	doc.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Top,
		containerLeftContentStyle.Align(lipgloss.Left).Render(docLeft.String()),
		containerRightContentStyle.Align(lipgloss.Left).Render(docRight.String()),
	))

	// helpDoc := strings.Builder{}

	// helpView := l.help.View(l.keys)
	// height := 8 - strings.Count(helpView, "\n")

	// helpDoc.WriteString(strings.Repeat("\n", height) + helpView)

	return doc.String()
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

func GetContainerLogs(dockerClient client.SDKClient, containerID string) string {
	ctx := context.Background()
	logs, err := dockerClient.ContainerLogs(ctx, containerID, containerTypes.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "10"})
	if err != nil {
		panic(err)
	}
	defer logs.Close()

	data, err := io.ReadAll(logs)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}
