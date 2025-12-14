package src

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/go-sdk/client"
	"github.com/muesli/reflow/wordwrap"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type logsModel struct {
	dockerClient  client.SDKClient
	containerID   string
	containerName string
	logs          string
	ready         bool
	viewport      viewport.Model
}

func InitLogsModel(dockerClient client.SDKClient, containerID string, containerName string) logsModel {
	return logsModel{
		dockerClient:  dockerClient,
		containerID:   containerID,
		containerName: containerName,
		logs:          GetContainerLogs(dockerClient, containerID, false),
	}
}

func (l logsModel) Init() tea.Cmd {
	return nil
}

func (l logsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" {
			return l, tea.Quit
		}

		switch msg.String() {
		case "esc":
			m := InitListContainersModel(l.dockerClient, l.viewport.Width, l.viewport.Height)
			return m, m.Init()
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(l.headerView())
		footerHeight := lipgloss.Height(l.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !l.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			l.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			l.viewport.YPosition = headerHeight
			l.viewport.SetContent(wordwrap.String(l.logs, msg.Width))
			l.viewport.GotoBottom()
			l.ready = true
		} else {
			l.viewport.Width = msg.Width
			l.viewport.Height = msg.Height - verticalMarginHeight
		}

	case tickMsg:
		l.logs = GetContainerLogs(l.dockerClient, l.containerID, true)
		l.viewport.SetContent(wordwrap.String(l.logs, l.viewport.Width))
		if l.viewport.ScrollPercent()*100 > 90 {
			l.viewport.GotoBottom()
		}
		return l, tickCmd()
	}

	// Handle keyboard and mouse events in the viewport
	l.viewport, cmd = l.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l logsModel) View() string {
	if !l.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", l.headerView(), l.viewport.View(), l.footerView())
}

func (l logsModel) headerView() string {
	title := titleStyle.Render(l.containerName)
	line := strings.Repeat("─", max(0, l.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (l logsModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", l.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, l.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func GetContainerLogs(dockerClient client.SDKClient, containerID string, refresh bool) string {
	ctx := context.Background()
	logs, err := dockerClient.ContainerLogs(ctx, containerID, containerTypes.LogsOptions{ShowStdout: true, ShowStderr: true, Since: "24h", Timestamps: true})
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
