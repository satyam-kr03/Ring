package main
 
import (
    "log"
    "os"
    "strconv"
    "ring/task"
    "ring/worker/api"
    "fmt"
    "time"
 
    "github.com/golang-collections/collections/queue"
    "github.com/google/uuid"
    "github.com/docker/docker/client"
    "github.com/joho/godotenv"
 
    "ring/worker"
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

func runTasks(w *worker.Worker) {
    for {
        if w.Queue.Len() != 0 {
            result := w.RunTask()
            if result.Error != nil {
                log.Printf("Error running task: %v\n", result.Error)
            }
        } else {
            log.Printf("No tasks to process currently.\n")
        }
        log.Println("Sleeping for 10 seconds.")
        time.Sleep(10 * time.Second)
    }
 
}
 
func main() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Printf("Error loading .env: %v", err)
    }
    host := os.Getenv("RING_HOST")
    port, _ := strconv.Atoi(os.Getenv("RING_PORT"))
 
    log.Printf("Host: %s, Port: %d", host, port)
 
    fmt.Println("Starting Ring worker")
 
    w := worker.Worker{
        Queue: *queue.New(),
        Db:    make(map[uuid.UUID]*task.Task),
    }
    api := api.Api{Address: host, Port: port, Worker: &w}
 
    go runTasks(&w)
    go w.CollectStats()
    api.Start()
}