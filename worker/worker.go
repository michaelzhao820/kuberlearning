package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"

	"kuberlearning/task"
)

/*
Worker duties :
1. Run tasks as Docker containers
2. Accept tasks from manager
3. Provide statistics to manager
4> keep track of its tasks
*/
type Worker struct {
	Name string
	Queue queue.Queue
	Db map[uuid.UUID]*task.Task
	Stats *Stats
	TaskCount int
}
func (w *Worker) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, t := range w.Db {
		tasks = append(tasks, t)
	}
	return tasks
}

func (w *Worker) CollectStats() {
	for {
        log.Println("Collecting stats")
        w.Stats = GetStats()
        w.Stats.TaskCount = w.TaskCount
        time.Sleep(15 * time.Second)
    }
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}
 
func (w *Worker) RunTask() task.DockerResult {
    t := w.Queue.Dequeue()

    if t == nil {
        log.Println("No tasks in the queue")
        return task.DockerResult{Error: nil}
    }
    //convert interface to task.Task type
    taskQueued := t.(task.Task)
	//checking if it is already in the database
    taskPersisted := w.Db[taskQueued.ID]
    if taskPersisted == nil {
        taskPersisted = &taskQueued
        w.Db[taskQueued.ID] = &taskQueued
    }

    var result task.DockerResult
    if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {

        switch taskQueued.State {
        case task.Scheduled:
           result =  w.StartTask(taskQueued)
		case task.Completed:
			result = w.StopTask(taskQueued)
        default:
            result.Error = errors.New("failed")
        }
    } else {
        err := fmt.Errorf("invalid transition from %v to %v",
            taskPersisted.State, taskQueued.State)
        result.Error = err
    }
    return result
}

 
func (w *Worker) StartTask(t task.Task) task.DockerResult {
    t.StartTime = time.Now().UTC()

	//Docker settings
	config := task.NewConfig(&t)
	d:= task.NewDocker(config)
	//Running of docker container
	result := d.Run()
	if result.Error != nil{
		log.Printf("Err running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}
	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t
	return result


}
 
func (w *Worker) StopTask(t task.Task) task.DockerResult {
    config := task.NewConfig(&t)
	d:= task.NewDocker(config)

	result:=d.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n",t.ID, result.Error )
	}
	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v\n",
	 t.ContainerID, t.ID)

	return result	
}


