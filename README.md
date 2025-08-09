# Container

A lightweight dependency injection and module lifecycle management container for Go applications.

## Overview

This package provides a simple yet powerful dependency injection container that manages module initialization, configuration, startup, and graceful shutdown. It's designed to help organize and manage the lifecycle of different components in your Go applications.

## Features

- **Dependency Injection**: Bind and resolve dependencies by name
- **Module Lifecycle Management**: Initialize, start, and stop modules
- **Configuration Management**: Global configuration support for modules
- **Graceful Shutdown**: Built-in support for graceful application shutdown
- **Interface-based Design**: Clean separation of concerns through interfaces

## Installation

```bash
go get github.com/recodextech/container
```

## Quick Start

```go
package main

import (
    "github.com/recodextech/container"
)

func main() {
    // Create a new container
    c := container.NewContainer()
    
    // Bind modules
    c.Bind("database", &DatabaseModule{})
    c.Bind("api", &APIModule{})
    
    // Initialize modules
    c.Init("database", "api")
    
    // Start modules (blocking call)
    c.Start("database", "api")
}
```

## Core Interfaces

### Container

The main container interface provides methods for dependency management:

```go
type Container interface {
    Init(modules ...string)           // Initialize modules
    Bind(typ string, obj any)         // Bind a dependency
    Resolve(name string) any          // Resolve a dependency
    GetGlobalConfig(typ string) any   // Get global configuration
}
```

### AppContainer

Extended container interface with additional lifecycle methods:

```go
type AppContainer interface {
    Container
    SetModuleGlobalConfig(configs ...ModuleConfig) error
    Start(modules ...string)          // Start modules
    Shutdown(modules ...string)       // Gracefully shutdown modules
}
```

### Module Interfaces

#### Initable
Modules that need initialization should implement this interface:

```go
type Initable interface {
    Init(Container) error
}
```

#### Runnable
Modules that run continuously (like servers) should implement this interface:

```go
type Runnable interface {
    Initable
    Run() error
}
```

#### Stoppable
Modules that need graceful shutdown should implement this interface:

```go
type Stoppable interface {
    Stop() error
}
```

#### Validator
Modules that need validation should implement this interface:

```go
type Validator interface {
    Validator() error
}
```

## Usage Examples

### Basic Module

```go
type DatabaseModule struct {
    config *DatabaseConfig
    db     *sql.DB
}

// Implement Initable
func (d *DatabaseModule) Init(c container.Container) error {
    // Get configuration from container
    config := c.GetGlobalConfig("database").(*DatabaseConfig)
    d.config = config
    
    // Initialize database connection
    db, err := sql.Open("postgres", config.ConnectionString)
    if err != nil {
        return err
    }
    d.db = db
    
    return nil
}

// Implement Runnable (if needed)
func (d *DatabaseModule) Run() error {
    // Keep connection alive, handle reconnections, etc.
    return nil
}

// Implement Stoppable
func (d *DatabaseModule) Stop() error {
    return d.db.Close()
}
```

### HTTP Server Module

```go
type APIModule struct {
    server *http.Server
    db     *DatabaseModule
}

func (a *APIModule) Init(c container.Container) error {
    // Resolve dependencies
    a.db = c.Resolve("database").(*DatabaseModule)
    
    // Setup HTTP server
    mux := http.NewServeMux()
    mux.HandleFunc("/health", a.healthHandler)
    
    a.server = &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    
    return nil
}

func (a *APIModule) Run() error {
    return a.server.ListenAndServe()
}

func (a *APIModule) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return a.server.Shutdown(ctx)
}

func (a *APIModule) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### Configuration Management

```go
type DatabaseConfig struct {
    Host     string `env:"DB_HOST" validate:"required"`
    Port     int    `env:"DB_PORT" validate:"required"`
    Database string `env:"DB_NAME" validate:"required"`
}

func main() {
    c := container.NewContainer()
    
    // Set module configurations
    dbConfig := &DatabaseConfig{}
    err := c.SetModuleGlobalConfig(
        container.ModuleConfig{
            Key:   "database",
            Value: dbConfig,
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Bind and initialize modules
    c.Bind("database", &DatabaseModule{})
    c.Init("database")
    
    // Start modules
    c.Start("database")
}
```

### Complete Application Example

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/recodextech/container"
)

func main() {
    // Create container
    c := container.NewContainer()
    
    // Setup configurations
    dbConfig := &DatabaseConfig{}
    apiConfig := &APIConfig{}
    
    err := c.SetModuleGlobalConfig(
        container.ModuleConfig{Key: "database", Value: dbConfig},
        container.ModuleConfig{Key: "api", Value: apiConfig},
    )
    if err != nil {
        log.Fatal("Failed to set configurations:", err)
    }
    
    // Bind modules
    c.Bind("database", &DatabaseModule{})
    c.Bind("api", &APIModule{})
    
    // Initialize modules in dependency order
    c.Init("database", "api")
    
    // Setup graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        
        log.Println("Shutting down...")
        c.Shutdown("api", "database") // Shutdown in reverse order
    }()
    
    // Start modules (this blocks until shutdown)
    c.Start("database", "api")
}
```

## Configuration Integration

The container integrates with [goconf](https://github.com/wgarunap/goconf) for configuration management, supporting:

- Environment variables
- Validation
- Multiple configuration sources

## Module Startup Order

Modules are started in the order they are provided to the `Start()` method. For graceful shutdown, reverse the order in the `Shutdown()` method to ensure dependencies are cleaned up properly.

## Error Handling

- Initialization errors cause panics to fail fast during startup
- Runtime errors in modules should be handled gracefully within the module
- Shutdown errors are logged but don't cause panics

## Thread Safety

The container uses mutex locks to ensure thread-safe access to internal maps and data structures.

## Dependencies

- [goconf](https://github.com/wgarunap/goconf) - Configuration management

## License

This project is open source. Please check the license file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
