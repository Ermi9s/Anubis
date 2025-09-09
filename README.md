# Anubis: Configuration and User Guide

Anubis is a Go library designed to integrate with services, consume logs from a message queue, and store them in a user specified database. It provides a robust mechanism for log management, exposing remote procedure calls (RPCs) that can be accessed programmatically or via a pre built CLI tool. This document provides a comprehensive guide for configuring and using Anubis, including supported databases, queues, configuration details, and usage instructions.


## Supported Technologies

Anubis supports the following database and queue technologies:

- **Database**: PostgreSQL
- **Queue**: RabbitMQ

## Configuration Guide

Configuration for Anubis must be specified in a YAML file, with one queue and one database configuration. The key names in the YAML file must match exactly as outlined below.

>### Database Configuration

>#### PostgreSQL

The PostgreSQL configuration specifies connection details for the database where logs will be stored.

```yaml
postgres:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "anubis"
  sslmode: "disable"
  timezone: "UTC"
```

- **host**: The database server address (e.g., `localhost` or a remote IP).
- **port**: The database port (default: `5432` for PostgreSQL).
- **user**: The database user.
- **password**: The user’s password.
- **dbname**: The name of the database (e.g., `anubis`).
- **sslmode**: SSL mode for the connection (e.g., `disable`, `require`).
- **timezone**: The timezone for timestamps (e.g., `UTC`).

### Queue Configuration

#### RabbitMQ

The RabbitMQ configuration defines the queue from which Anubis consumes log events.

```yaml
rabbit_mq:
  address: "amqp://localhost:5672/"
  queue_name: "audit_events"
  durable: true
  auto_delete: false
  exclusive: false
  no_wait: false
  args:
    x-message-ttl: 60000
    x-dead-letter-exchange: "dead_letter_exchange"
```

- **address**: The RabbitMQ server address (use `amqp://` protocol, e.g., `amqp://localhost:5672/`).
- **queue_name**: The name of the queue (e.g., `audit_events`).
- **durable**: If `true`, the queue persists after a broker restart.
- **auto_delete**: If `true`, the queue is deleted when no consumers are connected.
- **exclusive**: If `true`, the queue is exclusive to the connection.
- **no_wait**: If `true`, queue declaration does not wait for confirmation.
- **args**: Optional queue arguments (e.g., `x-message-ttl` for message time-to-live in milliseconds, `x-dead-letter-exchange` for dead-letter queue routing).

### Optional Configuration

Additional settings can be included to control Anubis’s behavior.

```yaml
read_concurrency: 5

#for http 
enable_http: true #this will expose an http GET /get url to get logs with all the filters 
port: "8081" #if http is enabled post must be specified 

```

- **read_concurrency**: The number of threads reading from the queue. Adjust based on workload (default: `5`).

## User Guide

After configuring Anubis, integrate it into your Go application and use the CLI tool or RPC client to interact with logs.

### Integrating Anubis into a Service

```bash
go get github.com/Ermi9s/Anubis/config 
```

To use Anubis, import the package and initialize it with the path to your configuration YAML file.

```go
package main

import (
    "github.com/Ermi9s/Anubis/config"
)

func main() {
    config.HostConfig("/path/to/config.yaml")
    // Anubis starts consuming from the queue and storing logs in the database.
}
```

- **config.HostConfig(configPath)**: Initializes Anubis with the specified YAML configuration file. This starts the queue consumer and database storage processes.

### RPC Server and CLI Tool

Anubis runs an RPC server upon calling `HostConfig`, enabling log retrieval and filtering. The Anubis CLI tool, a pre-built binary, leverages these RPCs for user-friendly log access. Developers can also create custom RPC clients to interact with the server.

#### RPC Server Implementation

The RPC server allows querying logs with flexible filtering options. Below is the server code structure:

```go
package rpcserver

import (
    "anubis/internal/model"
    "anubis/internal/repository"
    "time"
)

type RpcServer struct {
    Repository *repository.Repository
}

func NewRpcServer(repository *repository.Repository) *RpcServer {
    return &RpcServer{
        Repository: repository,
    }
}

type Args struct {
    Action      *string
    Status      *string
    ActorID     *string
    ActorType   *string
    StartTime   *time.Time
    EndTime     *time.Time
    ServiceName *string
    Page        int
    PageSize    int
    SortBy      string
    SortOrder   string
}

type Response struct {
    Data       []model.AuditEvent
    Pagination model.Pagination
}

func (rpc *RpcServer) FindLog(args *Args, response *Response) error {
    filter := model.AuditEventFilter{
        Action:      args.Action,
        Status:      args.Status,
        ActorID:     args.ActorID,
        ActorType:   args.ActorType,
        StartTime:   args.StartTime,
        EndTime:     args.EndTime,
        ServiceName: args.ServiceName,
        Page:        args.Page,
        PageSize:    args.PageSize,
        SortBy:      args.SortBy,
        SortOrder:   args.SortOrder,
    }

    paginatedResponse, err := rpc.Repository.FindAudit(filter)
    if err != nil {
        return err
    }

    response.Data = paginatedResponse.Data
    response.Pagination = paginatedResponse.Pagination

    return nil
}
```

- **Args**: Specifies filter criteria (e.g., `Action`, `Status`, `StartTime`) and pagination settings.
- **Response**: Returns a list of `AuditEvent` objects and pagination metadata.
- **FindLog**: Queries the database using the provided filter and populates the response.

#### Example RPC Client

Developers can create an RPC client to query logs programmatically. Below is an example:

```go
package main

import (
    "fmt"
    "log"
    "net/rpc"
    "time"
)

type Args struct {
    Action      *string
    Status      *string
    ActorID     *string
    ActorType   *string
    StartTime   *time.Time
    EndTime     *time.Time
    ServiceName *string
    Page        int
    PageSize    int
    SortBy      string
    SortOrder   string
}

type AuditEvent struct {
    ID          string
    Action      string
    Status      string
    ActorID     string
    ActorType   string
    ServiceName string
    Timestamp   time.Time
}

type Pagination struct {
    Page      int
    PageSize  int
    Total     int
    TotalPage int
}

type Response struct {
    Data       []AuditEvent
    Pagination Pagination
}

func main() {
    // Connect to the RPC server
    client, err := rpc.Dial("tcp", "localhost:1234")
    if err != nil {
        log.Fatalf("Failed to connect to RPC server: %v", err)
    }
    defer client.Close()

    // Example filter
    action := "LOGIN"
    args := Args{
        Action:    &action,
        Page:      1,
        PageSize:  10,
        SortBy:    "timestamp",
        SortOrder: "desc",
    }

    var response Response
    err = client.Call("RpcServer.FindLog", args, &response)
    if err != nil {
        log.Fatalf("Failed to call RPC: %v", err)
    }

    // Print results
    for _, event := range response.Data {
        fmt.Printf("[%s] %s by %s (%s)\n", event.Timestamp.Format(time.RFC3339), event.Action, event.ActorID, event.ServiceName)
    }
    fmt.Printf("Page %d of %d (Total: %d)\n", response.Pagination.Page, response.Pagination.TotalPage, response.Pagination.Total)
}
```

- **rpc.Dial**: Establishes a connection to the Anubis RPC server (default port: `1234`).
- **client.Call**: Invokes the `FindLog` method with filter arguments.
- **Response**: Displays filtered logs and pagination details.

### CLI Tool Usage

The Anubis CLI tool provides a user friendly interface to query logs. Download the pre-built binary and use commands to filter logs based on fields like `Action`, `Status`, `ActorID`, etc. Refer to the CLI documentation for specific commands and options.

> #### Usage

After the anubis binary is downloaded you can run it by using,

> .\anubis --help

> .\anubis find

or you can make it a system wide executable command.

> sudo mv ./anubis /usr/local/bin/

and run it anywhere using command

> anubis

##### further documentation and user guide can be found on `anubis --help` and `anubis find --help`


# Docker Image 

```yaml

  anubis:
    image: ermi9s/anubis:latest
    container_name: anubis
    ports:
      - "8081:8081"

    volumes:
      - ./config.yaml:/config.yaml:ro #needs to be mounted 
    environment:
      CONFIG_PATH: /config.yaml #specify the mounted config file
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    networks:
      - app-network

```

