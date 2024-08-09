package manager

import(
    "kuberlearning/task"
    "fmt"
 
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
    TaskDb map[string][]*task.Task 
    EventDb map[string][]*task.TaskEvent
    Workers []string
    WorkerTaskMap map[string][]uuid.UUID // input worker name, get task uuids
    TaskWorkerMap map[uuid.UUID]string //inut tasks uuid, get worker name
}

func (m *Manager) SelectWorker() {
    fmt.Println("I will select an appropriate worker")
}
 
func (m *Manager) UpdateTasks() {
    fmt.Println("I will update tasks")
}
 
func (m *Manager) SendWork() {
    fmt.Println("I will send work to workers")
}