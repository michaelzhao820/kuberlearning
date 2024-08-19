package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
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
}

func New(workers []string) *Manager {
    taskDb := make(map[uuid.UUID]*task.Task)
    workerTaskMap := make(map[string][]uuid.UUID)
    taskWorkerMap := make(map[uuid.UUID]string)
    for worker := range workers {
    workerTaskMap[workers[worker]] = []uuid.UUID{}
    }
    return &Manager{
    Pending: *queue.New(),
    Workers: workers,
    TaskDb: taskDb,
    WorkerTaskMap: workerTaskMap,
    TaskWorkerMap: taskWorkerMap,
    }
}

func (m *Manager) SelectWorker() string{
    if m.LastWorker+1 < len(m.Workers){
        m.LastWorker+=1
        return m.Workers[m.LastWorker]
    }else{
        m.LastWorker = 0
        return m.Workers[m.LastWorker]
    }
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
 
func (m *Manager) SendWork() {
    if m.Pending.Len() > 0 {
        w:= m.SelectWorker()

        e := m.Pending.Dequeue()
        t := e.(task.Task)
        log.Printf("Pulled %v off pending queue\n", t)
        
        m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w],t.ID)
        m.TaskWorkerMap[t.ID] = w
        m.TaskDb[t.ID] = &t

        data,err := json.Marshal(t)

        if err != nil {
            log.Printf("Unable to marshal task object: %v.\n", t)
        }
        url := fmt.Sprintf("http://%s/tasks",w)
        fmt.Printf("http://%s/tasks",w)
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

func (m *Manager) GetTasks () []*task.Task{
    tasks := []*task.Task{}
    for _,task := range m.TaskDb { 
        tasks = append(tasks, task)
    }
    return tasks
}

func (m *Manager) AddTask(t task.Task) {
    m.Pending.Enqueue(t)
}