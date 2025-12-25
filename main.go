package main
 
import (
    "log"
    "os"
    "strconv"
    "ring/task"
    "ring/worker/api"
    managerapi "ring/manager/api"
    "ring/task/state"
    "fmt"
    "time"

    "github.com/golang-collections/collections/queue"
    "github.com/google/uuid"
    "github.com/docker/docker/client"
    "github.com/joho/godotenv"

    "ring/worker"
    "ring/manager"
)

func createContainer() (*task.Docker, *task.DockerResult) {
    c := task.Config{
        Name:  "test-container-1",
        Image: "postgres:13",
        Env: []string{
            "POSTGRES_USER=cube",
            "POSTGRES_PASSWORD=secret",
        },
    }
 
    dc, _ := client.NewClientWithOpts(client.FromEnv)
    d := task.Docker{
        Client: dc,
        Config: c,
    }
 
    result := d.Run()
    if result.Error != nil {
        fmt.Printf("%v\n", result.Error)
        return nil, nil
    }
 
    fmt.Printf(
        "Container %s is running with config %v\n", result.ContainerId, c)
    return &d, &result
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
    result := d.Stop(id)
    if result.Error != nil {
        fmt.Printf("%v\n", result.Error)
        return nil
    }
 
    fmt.Printf(
        "Container %s has been stopped and removed\n", result.ContainerId)
    return &result
}

func main() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Printf("Error loading .env: %v", err)
    }
    whost := os.Getenv("RING_WORKER_HOST")
    if whost == "" {
        whost = "localhost"
    }
    wport, _ := strconv.Atoi(os.Getenv("RING_WORKER_PORT"))
    if wport == 0 {
        wport = 5555
    }

    mhost := os.Getenv("RING_MANAGER_HOST")
    if mhost == "" {
        mhost = "localhost"
    }
    mport, _ := strconv.Atoi(os.Getenv("RING_MANAGER_PORT"))
    if mport == 0 {
        mport = 5556
    }
 
    log.Printf("Host: %s, Port: %d", whost, wport)
 
    fmt.Println("Starting Ring worker")
 
    w := worker.Worker{
        Queue: *queue.New(),
        Db:    make(map[uuid.UUID]*task.Task),
    }
    wapi := api.Api{Address: whost, Port: wport, Worker: &w}
 
    go worker.RunTasks(&w)
    go w.CollectStats()
    go wapi.Start()

    workers := []string{fmt.Sprintf("%s:%d", whost, wport)}
    m := manager.New(workers)
    mapi := managerapi.Api{Address: mhost, Port: mport, Manager: m}

    go m.ProcessTasks();
    go m.UpdateTasks();

    for i := 0; i < 3; i++ {
        t := task.Task{
            ID:    uuid.New(),
            Name:  fmt.Sprintf("test-container-%d", i),
            State: state.Scheduled,
            Image: "strm/helloworld-http",
        }
        te := task.TaskEvent{
            ID:    uuid.New(),
            State: state.Running,
            Task:  t,
        }
        m.AddTask(te)
        m.SendWork()
    }

    mapi.Start()

    go func() {                                                       
        for {                                                         
            fmt.Printf("[Manager] Updating tasks from %d workers\n", len(m.Workers))
            m.UpdateTasks()                                           
            time.Sleep(15 * time.Second)
        }
    }()                                                               
 
    for {                                                             
        for _, t := range m.TaskDb {                                  
            fmt.Printf("[Manager] Task: id: %s, state: %d\n", t.ID, t.State)
            time.Sleep(15 * time.Second)
        }
    }
}