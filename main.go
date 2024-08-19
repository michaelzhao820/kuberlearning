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
	whost := os.Getenv("KLEARN_HOST")
	wport, _ := strconv.Atoi(os.Getenv("KLEARN_PORT"))

	mhost := os.Getenv("KLEARN_MANAGER_HOST")
 	mport, _ := strconv.Atoi(os.Getenv("KLEARN_MANAGER_PORT"))


	fmt.Println("Starting kuberlearn worker")

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	wapi := worker.Api{Address: whost, Port: wport, Worker: &w}

	go w.RunTasks()
	go w.CollectStats()
	go wapi.Start()

	fmt.Println("Starting kuberlearn manager")
	workers := []string{fmt.Sprintf("%s:%d", whost, wport)}
	m:= manager.New(workers)
	mapi := manager.Api{Address: mhost, Port: mport, Manager: m}

	m.ProcessTasks()
	m.UpdateTasks()

	mapi.Start()
}