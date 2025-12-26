# Ring

A distributed task orchestration system written in Go.

## Project Setup

1. **Prerequisites**:

   - Install Go (version 1.25.5 or later) from [golang.org](https://golang.org/dl/).
   - Install [Docker](https://docs.docker.com/desktop/setup/install/linux/).
   - Clone the repository and navigate to the project directory.

   ```bash
   git clone https://github.com/satyam-kr03/Ring.git && cd Ring
   ```

2. **Install Dependencies**:

   - Run `go mod tidy` to download and install Go modules listed in [go.mod](go.mod).

3. **Environment Configuration**:

   - Create a `.env` file in the root directory.
   - Set the following variables (defaults are used if not provided):
     - `RING_WORKER_HOST`: Host for the worker API (default: `localhost`).
     - `RING_WORKER_PORT`: Port for the worker API (default: `5555`).
     - `RING_MANAGER_HOST`: Host for the manager API (default: `localhost`).
     - `RING_MANAGER_PORT`: Port for the manager API (default: `5556`).

4. **Build and Run**:
   - Build the project: `go build -o ring main.go`.
   - Run the system: `./ring`. This starts the manager and worker components, which communicate via HTTP APIs defined in [manager/api/api.go](manager/api/api.go) and [worker/api/api.go](worker/api/api.go).
   - Ensure that the docker daemon is already running on your system.

## Architecture

The orchestrator consists of the following components:

- **Manager**: Acts as the central coordinator. It accepts task requests via its API ([manager/api/api.go](manager/api/api.go)), schedules tasks to workers using round-robin selection ([`manager.SelectWorker`](manager/manager.go)), and tracks task states in databases ([`manager.Manager.TaskDb`](manager/manager.go)). It periodically updates task statuses from workers.

- **Worker**: Executes tasks as Docker containers. Each worker has a queue for incoming tasks ([`worker.Worker.Queue`](worker/worker.go)), a database for task tracking ([`worker.Worker.Db`](worker/worker.go)), and collects system stats ([`worker/stats/stats.go`](worker/stats/stats.go)). Tasks transition through states defined in [task/state/state.go](task/state/state.go), with Docker operations handled in [task/task.go](task/task.go).

- **Task**: Represents a unit of work, including Docker image, resources, and state. Tasks are managed via events ([`task.TaskEvent`](task/task.go)) and run in containers with configurable resources.

- **API Layer**: Both manager and worker expose REST APIs using [Chi router](https://github.com/go-chi/chi) for task management (start, stop, list) and stats retrieval.

- **Scheduler**: An interface for future extensibility ([scheduler/scheduler.go](scheduler/scheduler.go)), currently using simple round-robin in the manager.

The system uses HTTP for inter-component communication, UUIDs for task identification, and relies on Docker for containerization. Nodes are represented by [`node.Node`](node/node.go).

References: [_Build an Orchestrator in Go (From Scratch)_](https://www.manning.com/books/build-an-orchestrator-in-go-from-scratch?new=true&experiment=B)
