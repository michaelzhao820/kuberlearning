package task

import (
	"github.com/google/uuid"
	"github.com/docker/go-connections/nat"
	"time"
)

type State int

const (
	Pending State = iota
	Scheduled
	Completed
	Running
	Failed
)

//task that user wants to run
type Task struct {
	ID uuid.UUID
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

//wrapper around task that allows for transition of states
type TaskEvent struct {
	ID uuid.UUID
	State State
	Timestamp time.Time
	Task Task
}

