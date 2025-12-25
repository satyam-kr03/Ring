package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ring/manager"
	"ring/task"
	"ring/task/state"
)

type ErrResponse struct {
    HTTPStatusCode int    `json:"-"`
    Message        string `json:"message"`
}

type Api struct {
    Address string
    Port    int
    Manager *manager.Manager
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
}

func (a *Api) Start() {
    a.initRouter()
    http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
    d := json.NewDecoder(r.Body)                                          
    d.DisallowUnknownFields()
 
    te := task.TaskEvent{}
    err := d.Decode(&te)                                                  
    if err != nil {                                                       
        msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
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
    log.Printf("Added task %v\n", te.Task.ID)
    w.WriteHeader(201)                                                    
    json.NewEncoder(w).Encode(te.Task)                                    
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(a.Manager.GetTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
    taskID := chi.URLParam(r, "taskID")                
    if taskID == "" {
        log.Printf("No taskID passed in request.\n")
        w.WriteHeader(400)
    }
 
    tID, _ := uuid.Parse(taskID)
    taskToStop, ok := a.Manager.TaskDb[tID]            
    if !ok {
        log.Printf("No task with ID %v found", tID)
        w.WriteHeader(404)
    }
 
    te := task.TaskEvent{                              
        ID:        uuid.New(),
		State:     state.Completed,
        Timestamp: time.Now(),
    }
 
    taskCopy := *taskToStop                            
    taskCopy.State = state.Completed
    te.Task = taskCopy
    a.Manager.AddTask(te)                              
 
    log.Printf("Added task event %v to stop task %v\n", te.ID, taskToStop.ID)
    w.WriteHeader(204)                                 
}