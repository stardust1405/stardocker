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
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/go-sdk/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tickMsg time.Time

// List Containers Model

type listContainersModel struct {
	help            help.Model
	keys            keyMap
	dockerClient    client.SDKClient
	width           int
	height          int
	table           table.Model
	ShowChildrenSet StringSet
}

const composeStackIdentifier = "com.docker.compose.project"

type ContainerType string

func (ct ContainerType) String() string {
	return string(ct)
}

const (
	TypeContainer    ContainerType = "container"
	TypeComposeStack ContainerType = "compose_stack"
)

var tableBaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	MarginLeft(1)

type Container struct {
	ID       string
	Name     string
	Type     ContainerType
	Image    string
	Ports    []containerTypes.Port
	Status   string
	State    containerTypes.ContainerState
	Children []Container
}

func InitListContainersModel(dockerClient client.SDKClient, width int, height int) listContainersModel {
	columns := []table.Column{
		{Title: "⏺", Width: 2},
		{Title: "Name", Width: 35},
		{Title: "Container ID", Width: 15},
		{Title: "Image", Width: 15},
		{Title: "Ports", Width: 15},
		{Title: "Status", Width: 32},
		{Title: "State", Width: 10},
		{Title: "Type", Width: 20},
	}

	// rows := getRows(allContainers)

	t := table.New(
		table.WithColumns(columns),
		// table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-10),
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

	return listContainersModel{
		help:            help.New(),
		keys:            keys,
		dockerClient:    dockerClient,
		width:           width,
		height:          height,
		table:           t,
		ShowChildrenSet: make(StringSet),
	}
}

func (l listContainersModel) Init() tea.Cmd {
	allContainers := FetchContainers(l.dockerClient)
	l.table.SetRows(l.getRows(allContainers))
	return tea.Batch(tea.SetWindowTitle("Containers"), tickCmd())
}

func (l listContainersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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

		case "enter":
			// return InitViewContainersModel(l.DB, l.Containers[l.cursor], l.ContainersDownloadPercent[l.cursor], "/home/stardust/Downloads"), nil
			row := l.table.SelectedRow()
			containerType := row[7]

			if containerType == TypeComposeStack.String() {
				showChildren := !l.ShowChildrenSet.Contains(row[1])
				if showChildren {
					l.ShowChildrenSet.Add(row[1])
				} else {
					l.ShowChildrenSet.Remove(row[1])
				}
				allContainers := FetchContainers(l.dockerClient)
				l.table.SetRows(l.getRows(allContainers))
			}

		case "q":
			return l, tea.Quit

		case "r":
			row := l.table.SelectedRow()
			containerID := strings.TrimSpace(row[2])
			containerType := strings.TrimSpace(row[7])
			containerState := strings.TrimSpace(row[6])

			if containerType == TypeContainer.String() {
				if containerState == containerTypes.StateRunning {
					StopContainer(l.dockerClient, containerID)
				}
				if containerState == containerTypes.StateExited {
					StartContainer(l.dockerClient, containerID)
				}
			}

		case "s":
			row := l.table.SelectedRow()
			containerID := strings.TrimSpace(row[2])
			containerType := strings.TrimSpace(row[7])

			if containerType == TypeContainer.String() {
				StartContainer(l.dockerClient, containerID)
			}

		case "d":
			row := l.table.SelectedRow()
			containerID := strings.TrimSpace(row[2])
			containerType := strings.TrimSpace(row[7])

			if containerType == TypeContainer.String() {
				StopContainer(l.dockerClient, containerID)
			}
		}

	case tickMsg:
		allContainers := FetchContainers(l.dockerClient)
		l.table.SetRows(l.getRows(allContainers))
		return l, tickCmd()
	}

	l.table, cmd = l.table.Update(msg)

	return l, cmd
}

func (l listContainersModel) View() string {
	doc := strings.Builder{}

	title := lipgloss.PlaceHorizontal(l.width, lipgloss.Left, ContainerTitleStyle.Render("CONTAINERS"))

	doc.WriteString(title)

	doc.WriteString("\n\n")

	doc.WriteString(tableBaseStyle.Render(l.table.View()) + "\n")

	helpDoc := strings.Builder{}

	helpView := l.help.View(l.keys)
	height := 8 - strings.Count(helpView, "\n")

	helpDoc.WriteString(strings.Repeat("\n", height) + helpView)

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
				Type:   TypeContainer,
				Image:  container.Image,
				Ports:  container.Ports,
				Status: container.Status,
				State:  container.State,
			})
		} else {
			allContainers = append(allContainers, Container{
				ID:     container.ID,
				Name:   strings.TrimLeft(container.Names[0], "/"),
				Image:  container.Image,
				Ports:  container.Ports,
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

func (l listContainersModel) getRows(containers []Container) []table.Row {
	rows := []table.Row{}

	for _, container := range containers {
		indicator := " "
		if container.State == containerTypes.StateRunning {
			indicator = "⏺"
		}
		for _, child := range container.Children {
			if child.State == containerTypes.StateRunning {
				indicator = "⏺"
			}
		}
		rows = append(rows, table.Row{
			indicator,
			container.Name,
			container.ID,
			container.Image,
			fmt.Sprintf("%v", container.Ports),
			container.Status,
			container.State,
			container.Type.String(),
		})
		if container.Type == TypeComposeStack && l.ShowChildrenSet.Contains(container.Name) {
			for _, child := range container.Children {
				nestedIndicator := " "
				if strings.TrimSpace(child.State) == containerTypes.StateRunning {
					nestedIndicator = "⏺"
				}
				rows = append(rows, table.Row{
					nestedIndicator,
					"  " + child.Name,
					"  " + child.ID,
					"  " + child.Image,
					"  " + fmt.Sprintf("%v", child.Ports),
					"  " + child.Status,
					"  " + child.State,
					"  " + child.Type.String(),
				})
			}
		}
	}

	return rows
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

// Define a set type using a map
type StringSet map[string]struct{}

// Function to add an element to the set
func (s StringSet) Add(value string) {
	s[value] = struct{}{}
}

// Function to remove an element from the set
func (s StringSet) Remove(value string) {
	delete(s, value)
}

// Function to check if an element is in the set
func (s StringSet) Contains(value string) bool {
	_, exists := s[value]
	return exists
}
