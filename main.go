package main
 
import (
    "ring/task"
    "ring/task/state"
    "fmt"
    "time"
 
    "github.com/golang-collections/collections/queue"
    "github.com/google/uuid"
    "github.com/docker/docker/client"
 
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
 
func main() {
    db := make(map[uuid.UUID]*task.Task)
    w := worker.Worker{
        Queue: *queue.New(),
        Db:    db,
    }
 
    t := task.Task{
        ID:    uuid.New(),
        Name:  "test-container-1",
        State: state.Scheduled,
        Image: "strm/helloworld-http",
    }
 
    fmt.Println("starting task")
    w.AddTask(t)
    result := w.RunTask()
    if result.Error != nil {
        panic(result.Error)
    }

    t.ContainerID = result.ContainerId
    fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerID)
    fmt.Println("Sleepy time")
    time.Sleep(time.Second * 30)
 
    fmt.Printf("stopping task %s\n", t.ID)
    t.State = state.Completed
    w.AddTask(t)
    result = w.RunTask()
    if result.Error != nil {
        panic(result.Error)
    }
}