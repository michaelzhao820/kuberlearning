package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"kuberlearning/node"
	"kuberlearning/scheduler"
	"kuberlearning/task"
	"kuberlearning/worker"
	"log"
	"net/http"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

/*
Manager duties :
1. Accept tasks from user
2. Schedule Tasks to Workers
3. Keep Track of Tasks
*/

type Manager struct {
	Pending queue.Queue
    TaskDb map[uuid.UUID]*task.Task
    Workers []string
    WorkerTaskMap map[string][]uuid.UUID // input worker name, get task uuids
    TaskWorkerMap map[uuid.UUID]string //inut tasks uuid, get worker name
    LastWorker int
    WorkerNodes []*node.Node
    Scheduler scheduler.Scheduler
}

func New(workers []string, schedulerType string) *Manager {
    taskDb := make(map[uuid.UUID]*task.Task)
    workerTaskMap := make(map[string][]uuid.UUID)
    taskWorkerMap := make(map[uuid.UUID]string)
    var nodes []*node.Node
    for _,worker := range workers {
        workerTaskMap[worker] = []uuid.UUID{}

        nAPI := fmt.Sprintf("http://%v", worker)
        n:= node.NewNode(worker,nAPI,"worker node")
        nodes = append(nodes,n)
    }
    var s scheduler.Scheduler
    switch (schedulerType){
        case "roundrobin" :
            s = &scheduler.RoundRobin{Name : "roundrobin"}
        default:
            s = &scheduler.RoundRobin{Name: "roundrobin"}
    }
    return &Manager{
        Pending: *queue.New(),
        Workers: workers,
        TaskDb: taskDb,
        WorkerTaskMap: workerTaskMap,
        TaskWorkerMap: taskWorkerMap,
        WorkerNodes: nodes,
        Scheduler: s,
    }
}

func (m *Manager) SelectWorker(t task.Task) (*node.Node, error){

    possibleWorkers := m.Scheduler.SelectCanidateNodes(t,m.WorkerNodes)
    if possibleWorkers == nil {
        msg := fmt.Sprintf("No available candidates match resource request for task %v", t.ID)
        err := errors.New(msg)
        return nil, err
    }
    scores := m.Scheduler.Score(t,possibleWorkers)
    return m.Scheduler.Pick(scores,possibleWorkers),nil
}

func (m *Manager) UpdateTasks() {
	for {
		log.Println("Checking for task updates from workers")
		m.updateTasks()
		log.Println("Task updates completed")
		log.Println("Sleeping for 15 seconds")
		time.Sleep(15 * time.Second)
	}
}
 

// Getting stat updates from the workers
func (m *Manager) updateTasks() {
    for _, worker := range m.Workers{
        log.Printf("Checking worker %v for task updates", worker)
        url := fmt.Sprintf("http://%s/tasks",worker)
        resp, err := http.Get(url)
        if err != nil {
            log.Printf("Error connecting to %v: %v\n", worker, err)
        }
        d:= json.NewDecoder(resp.Body)
        var tasks []*task.Task
        if err := d.Decode(&tasks); err != nil {
            fmt.Println("Error decoding JSON:", err)
            return
        }
        for _, task := range tasks {
            log.Printf("Attempting to update task %v\n", task.ID)
            _, ok := m.TaskDb[task.ID]
            if !ok {
                log.Printf("Task with ID %s not found\n", task.ID)
                return
            }
            if m.TaskDb[task.ID].State != task.State{
                m.TaskDb[task.ID].State = task.State
            }
            m.TaskDb[task.ID].StartTime = task.StartTime
            m.TaskDb[task.ID].FinishTime = task.FinishTime
            m.TaskDb[task.ID].ContainerID = task.ContainerID
        }
    }
}

func (m *Manager) ProcessTasks() {
	for {
		log.Println("Processing any tasks in the queue")
		m.SendWork()
		log.Println("Sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}
}

func (m *Manager) stopTask(worker string, taskID string) {
	client := &http.Client{}
	url := fmt.Sprintf("http://%s/tasks/%s", worker, taskID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("error creating request to delete task %s: %v", taskID, err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error connecting to worker at %s: %v", url, err)
		return
	}

	if resp.StatusCode != 204 {
		log.Printf("Error sending request: %v", err)
		return
	}

	log.Printf("task %s has been scheduled to be stopped", taskID)
}
 
func (m *Manager) SendWork() {
    if m.Pending.Len() > 0 {

        e := m.Pending.Dequeue()
        t := e.(task.Task)
        log.Printf("Pulled %v off pending queue\n", t)

        taskWorker, ok := m.TaskWorkerMap[t.ID]
        if ok {
            persistedTask := m.TaskDb[t.ID]
            if t.State == task.Completed && task.ValidStateTransition(persistedTask.State,t.State){
                m.stopTask(taskWorker,t.ID.String())
                return
            }
            log.Printf("invalid request: existing task %s is in state %v and cannot transition to the completed state", persistedTask.ID.String(), persistedTask.State)
			return
        }

        w,err := m.SelectWorker(t)
        if err != nil {
            log.Printf("error selecting worker for task %s: %v", t.ID, err)
			return
        }
        log.Printf("[manager] selected worker %s for task %s", w.Name, t.ID)
        
        m.WorkerTaskMap[w.Name] = append(m.WorkerTaskMap[w.Name],t.ID)
        m.TaskWorkerMap[t.ID] = w.Name
        m.TaskDb[t.ID] = &t

        data,err := json.Marshal(t)

        if err != nil {
            log.Printf("Unable to marshal task object: %v.\n", t)
        }
        url := fmt.Sprintf("http://%s/tasks",w.Name)
        //API call to send task to a worker
        resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
        if err != nil {
            log.Printf("Error connecting to %v: %v\n", w, err)
            m.Pending.Enqueue(t)
            return
        }
        d := json.NewDecoder(resp.Body)
        if resp.StatusCode != http.StatusCreated {
            e := worker.ErrResponse{}
            err := d.Decode(&e)
            if err != nil {
            fmt.Printf("Error decoding response: %s\n", err.Error())
            return
            }
            log.Printf("Response error (%d): %s", e.HTTPStatusCode, e.Message)
            return
        }
    } else {
        log.Println("No work in the queue")
    }
       
}

func (m *Manager) GetTasks() []*task.Task{
    tasks := []*task.Task{}
    for _,task := range m.TaskDb { 
        tasks = append(tasks, task)
    }
    return tasks
}

func (m *Manager) AddTask(t task.Task) {
    m.Pending.Enqueue(t)
}