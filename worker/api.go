package worker

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ErrResponse struct {
	HTTPStatusCode int
	Message        string
}

/*
Address and Port define local IP address
of where API runs
Worker serves as reference that this API wraps
*/
type Api struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
}

func (a *Api) initRouter() {
    a.Router = chi.NewRouter()                         
    a.Router.Route("/tasks", func(r chi.Router) {      
        r.Post("/", a.StartTaskHandler)                
        r.Get("/", a.GetTasksHandler)
        r.Route("/{taskID}", func(r chi.Router) {      
            r.Delete("/", a.StopTaskHandler)
        })
    })
    a.Router.Route("/stats", func(r chi.Router) {
        r.Get("/", a.GetStatsHandler)
    })
}

func (a *Api) Start() {
    a.initRouter()
    http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}