package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/stardust1405/stardocker/src"

	tea "github.com/charmbracelet/bubbletea"
	vanillaDockerClient "github.com/docker/docker/client"
	"github.com/docker/go-sdk/client"
)

func main() {
	ctx := context.TODO()

	// Try creating a Docker client
	cli, err := vanillaDockerClient.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	// Check if Docker is alive
	_, err = cli.Ping(ctx)
	if err != nil {
		fmt.Println("Docker daemon not running. Starting Docker Desktop...")

		// Start Docker Desktop headless
		exec.Command("open", "-a", "Docker", "--args", "--unattended").Start()

		// Wait for daemon to come up
		for {
			_, err = cli.Ping(ctx)
			if err == nil {
				fmt.Println("Docker is ready!")
				break
			}
			fmt.Println("Waiting for Docker daemon...")
			time.Sleep(2 * time.Second)
		}
	} else {
		fmt.Println("Docker already running.")
	}

	// Close the client when done
	err = cli.Close()
	if err != nil {
		panic(err)
	}

	dockerClient, err := client.New(ctx)
	if err != nil {
		panic(err)
	}

	// Close the docker client when done
	defer dockerClient.Close()

	p := tea.NewProgram(src.InitHomeModel(dockerClient), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error encountered, terminating: %v", err)
		os.Exit(1)
	}

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
}
