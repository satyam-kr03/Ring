package worker
 
import (
    "errors"
    "fmt"
    "log"
    "time"
    "github.com/google/uuid"
    "github.com/golang-collections/collections/queue"
 
    "ring/task"
    "ring/task/state"
)
 
type Worker struct {
    Name      string
    Queue     queue.Queue
    Db        map[uuid.UUID]*task.Task
    TaskCount int
}

func (w *Worker) CollectStats() {
    fmt.Println("I will collect stats")
}
 
func (w *Worker) RunTask() task.DockerResult {
    t := w.Queue.Dequeue()                          
    if t == nil {
        log.Println("No tasks in the queue")
        return task.DockerResult{Error: nil}
    }
 
    taskQueued := t.(task.Task)    
    taskPersisted := w.Db[taskQueued.ID]                  
    if taskPersisted == nil {
        taskPersisted = &taskQueued
        w.Db[taskQueued.ID] = &taskQueued
    }
 
    var result task.DockerResult
    if state.ValidStateTransition(taskPersisted.State, taskQueued.State) {
        switch taskQueued.State {
        case state.Scheduled:
            result = w.StartTask(taskQueued)             
        case state.Completed:
            result = w.StopTask(taskQueued)              
        default:
            result.Error = errors.New("We should not get here")
        }
    } else {
        err := fmt.Errorf("Invalid transition from %v to %v",
            taskPersisted.State, taskQueued.State)
        result.Error = err                               
    }
    return result                                        
}
 
func (w *Worker) StartTask(t task.Task) task.DockerResult {
    t.StartTime = time.Now().UTC()
    config := task.NewConfig(&t)
    d := task.NewDocker(config)
    result := d.Run()
    if result.Error != nil {
        log.Printf("Err running task %v: %v\n", t.ID, result.Error)
        t.State = state.Failed
        w.Db[t.ID] = &t
        return result
    }
 
    t.ContainerID = result.ContainerId
    t.State = state.Running
    w.Db[t.ID] = &t
 
    return result
}
 
func (w *Worker) StopTask(t task.Task) task.DockerResult {
    config := task.NewConfig(&t)
    d := task.NewDocker(config)
 
    result := d.Stop(t.ContainerID)
    if result.Error != nil {
        log.Printf("Error stopping container %v: %v\n", t.ContainerID, result.Error)
    }
    t.FinishTime = time.Now().UTC()
    t.State = state.Completed
    w.Db[t.ID] = &t
    log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)
 
    return result
}

func (w *Worker) AddTask(t task.Task) {
    w.Queue.Enqueue(t)
}

/*
The worker’s requirements are:

Run tasks as Docker containers -> Db
Accept tasks to run from a manager -> Queue
Provide relevant statistics to the manager for the purpose of scheduling tasks -> TaskCount
Keep track of its tasks and their state -> Db
*/