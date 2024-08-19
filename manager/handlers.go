package manager

import (
	"encoding/json"
	"fmt"
	"kuberlearning/task"
	"log"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)


func (a*Api) StartTaskHandler (w http.ResponseWriter,r *http.Request){
	d:=json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	te := task.Task{}
	err := d.Decode(&te) 
	if err != nil {
		msg := fmt.Sprintf("Error body: %v\n", err)
        log.Printf(msg)
        w.WriteHeader(400)
        e := ErrResponse{
            HTTPStatusCode: 400,
            Message:        msg,
        }
        json.NewEncoder(w).Encode(e)
        return
	}
	a.Manager.AddTask(te)
	log.Printf("Added task %v\n", te.ID)        
    w.WriteHeader(201)                               
    json.NewEncoder(w).Encode(te) 

}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(a.Manager.GetTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request){
	taskID := chi.URLParam(r,"taskID")
	if taskID == "" {                                                   
        log.Printf("No taskID passed in request.\n")
        w.WriteHeader(400)
    }

	tID,_ := uuid.Parse(taskID)
	_, ok := a.Manager.TaskDb[tID] 
	if !ok {
		log.Printf("No task with ID %v found", tID)
        w.WriteHeader(404)
	}

	taskToStop := a.Manager.TaskDb[tID]
	//taskCopy is a copy, can't mess with the one in db
	taskCopy := *taskToStop 
	taskCopy.State = task.Completed
	a.Manager.AddTask(taskCopy)

	log.Printf("Added task %v to stop container %v\n", taskToStop.ID,
	taskToStop.ContainerID)                                       
	w.WriteHeader(204)   

}