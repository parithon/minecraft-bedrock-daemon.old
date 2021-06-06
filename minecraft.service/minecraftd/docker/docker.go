package docker

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

const (
	dockerImage        = "parithon/minecraftd"
	portMappings       = "0.0.0.0:19132:19132/udp"
	containerName      = "minecraft-bedrock-server"
	heathcheckInterval = 15
)

var (
	ctx         context.Context
	cli         *client.Client
	containerID string
)

func getStopSignal(debug bool) string {
	if !debug {
		return "SIGQUIT"
	}
	return "SIGTERM"
}

func printContainerLogs() {
	out, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func Init(debug *bool) {
	ctx = context.Background()

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	log.Println("Pulling latest image from Docker hub...")

	if UpdateAvailable() {
		log.Println("Pulled latest image, starting container...")
	} else {
		log.Println("Image is alrady up to date, starting container...")
	}

	exposedPorts, portBindings, _ := nat.ParsePortSpecs([]string{
		portMappings,
	})

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        dockerImage,
		ExposedPorts: exposedPorts,
		AttachStdout: true,
		StopSignal:   getStopSignal(*debug),
		Healthcheck: &container.HealthConfig{
			Interval: time.Second * time.Duration(heathcheckInterval),
			Test: []string{
				"{'CMD-SHELL', 'minecraftd healthcheck'}",
			},
		},
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	containerID = resp.ID

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}

	log.Println("Container started")

	go printContainerLogs()
}

func Shutdown(signal *os.Signal) {
	log.Println("Stopping docker container...")
	timeout := time.Second * time.Duration(35)
	cli.ContainerStop(ctx, containerID, &timeout)
	log.Println("Docker container stopped")
	time.Sleep(time.Millisecond * time.Duration(1))
}

func Cleanup() {
	log.Println("Removing docker container...")
	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		log.Fatal(err)
	}
	log.Println("Docker container removed")
}

func Wait() {
	statusCH, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		{
			if err != nil {
				log.SetPrefix("ERROR ")
				log.Print(err)
			}
		}
	case <-statusCH:
	}
}

func UpdateAvailable() bool {
	out, err := cli.ImagePull(ctx, dockerImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer out.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	result := buf.String()
	if strings.Contains(result, "Image is up to date") {
		return false
	} else {
		return true
	}
}
