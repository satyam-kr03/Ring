import(
    "ring/task"
    "fmt"
 
    "github.com/golang-collections/collections/queue"
    "github.com/google/uuid"
)
 
type Manager struct {
    Pending       queue.Queue
    TaskDb        map[string][]*task.Task
    EventDb       map[string][]*task.TaskEvent
    Workers       []string
    WorkerTaskMap map[string][]uuid.UUID
    TaskWorkerMap map[uuid.UUID]string
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

/*
The manager’s requirements are:

Accept requests from users to start and stop tasks -> Pending
Schedule tasks onto worker machines -> SelectWorker() and SendWork()
Keep track of tasks, their states, and the machine on which they run -> TaskDb, EventDb
*/