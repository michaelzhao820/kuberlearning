package main

import (
	"fmt"
	"kuberlearning/manager"
	"kuberlearning/task"
	"kuberlearning/worker"
	"os"
	"strconv"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	whost := os.Getenv("KLEARN_WORKER_HOST")
	wport, _ := strconv.Atoi(os.Getenv("KLEARN_WORKER_PORT"))

	mhost := os.Getenv("KLEARN_MANAGER_HOST")
	mport, _ := strconv.Atoi(os.Getenv("KLEARN_MANAGER_PORT"))

	fmt.Println("Starting kuberlearning worker")

	w1 := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	wapi1 := worker.Api{Address: whost, Port: wport, Worker: &w1}

	w2 := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	wapi2 := worker.Api{Address: whost, Port: wport + 1, Worker: &w2}

	w3 := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	wapi3 := worker.Api{Address: whost, Port: wport + 2, Worker: &w3}

	go w1.RunTasks()
	go wapi1.Start()

	go w2.RunTasks()
	go wapi2.Start()

	go w3.RunTasks()
	go wapi3.Start()

	fmt.Println("Starting kuberlearning manager")

	workers := []string{
		fmt.Sprintf("%s:%d", whost, wport),
		fmt.Sprintf("%s:%d", whost, wport+1),
		fmt.Sprintf("%s:%d", whost, wport+2),
	}
	m := manager.New(workers, "roundrobin")
	mapi := manager.Api{Address: mhost, Port: mport, Manager: m}

	go m.ProcessTasks()
	go m.UpdateTasks()

	mapi.Start()

}
