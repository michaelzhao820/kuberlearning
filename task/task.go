package task

import (
	"io"
	"log"
	"os"
	"time"

	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/docker/docker/api/types"
	"github.com/moby/moby/pkg/stdcopy"
)



//task that user wants to run
type Task struct {
	ID uuid.UUID
	ContainerID string
	Name string
	State State
	Image string
	CPU float64
	Memory int64
	Disk int64
	ExposedPorts nat.PortSet
	PortBindings map[string]string
	RestartPolicy string
	StartTime time.Time
	FinishTime time.Time
}


//Configuration of task (container settings)
type Config struct {
    Name          string
    AttachStdin   bool
    AttachStdout  bool
    AttachStderr  bool
    ExposedPorts  nat.PortSet
    Cmd           []string
    Image         string
    Cpu           float64
    Memory        int64
    Disk          int64
    Env           []string
    RestartPolicy string
}

func NewConfig(t *Task) *Config {
	return &Config{
		Name : t.Name,
		ExposedPorts: t.ExposedPorts,
		Image : t.Image,
		Cpu: t.CPU,
		Memory: t.Memory,
		Disk: t.Disk,
		RestartPolicy: t.RestartPolicy,
	}
} 

type Docker struct {
    Client *client.Client
    Config  Config
}
func NewDocker(c *Config) *Docker{
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	return &Docker{
		Client: dc,
		Config: *c,
	}

}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})
	if (err != nil) {
		log.Printf("Error pulling image %s:%v\n", d.Config.Image,err)
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout,reader)

	rp := container.RestartPolicy{
		Name: d.Config.RestartPolicy,
	}
	r := container.Resources{
		Memory: d.Config.Memory,
	}
	cc:= container.Config{
		Image: d.Config.Image,
		Tty : false,
		Env : d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}
	hc:= container.HostConfig{
		RestartPolicy: rp,
		Resources: r,
		PublishAllPorts: true,
	}
	resp, err := d.Client.ContainerCreate(ctx,&cc,&hc,nil,nil,d.Config.Name)
	if err != nil{
		log.Printf("Error getting container %s started: %v\n", d.Config.Image,err)
		return DockerResult{Error: err}
	}
	if err = d.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("Error starting container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}
	out, err := d.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return DockerResult{ContainerId: resp.ID, Action: "start", Result: "success"}
}


func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Attempting to stop container %v", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, nil)
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})
	if err != nil {
		log.Printf("Error removing container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}
