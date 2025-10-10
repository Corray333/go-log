# go-log

A beautifully colored, structured logging library for Go applications built on top of the standard `log/slog` package.

[![Go Reference](https://pkg.go.dev/badge/github.com/corray333/go-log.svg)](https://pkg.go.dev/github.com/corray333/go-log)
[![Go Report Card](https://goreportcard.com/badge/github.com/corray333/go-log)](https://goreportcard.com/report/github.com/corray333/go-log)

## Features

- Custom colored output for different log levels
- Built on top of Go's standard `log/slog` package
- Automatic file and line number tracking for error logs
- HTTP middleware for Chi router
- Thread-safe logging with proper synchronization
- Structured logging with JSON attributes
- Easy integration with existing slog-based applications

## Installation

```bash
go get github.com/corray333/go-log
```

## Quick Start

### Basic Usage

```go
package main

import (
    "log/slog"
    golog "github.com/corray333/go-log"
)

func main() {
    // Setup the custom logger as default
    golog.SetupCustomLogger()

    // Use standard slog functions
    slog.Info("Application started")
    slog.Debug("Debug information")
    slog.Warn("Warning message")
    slog.Error("Error occurred")
}
```

### Custom Handler

```go
package main

import (
    "log/slog"
    golog "github.com/corray333/go-log"
)

func main() {
    // Create a custom handler with options
    handler := golog.NewHandler(&slog.HandlerOptions{
        Level:     slog.LevelDebug,
        AddSource: false,
    })

    // Create and set the logger
    logger := slog.New(handler)
    slog.SetDefault(logger)

    slog.Info("Custom logger initialized")
}
```

### HTTP Middleware

```go
package main

import (
    "log/slog"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    golog "github.com/corray333/go-log"
    logmiddleware "github.com/corray333/go-log/middleware"
)

func main() {
    // Setup logger
    golog.SetupCustomLogger()
    logger := slog.Default()

    // Setup router
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(logmiddleware.NewLoggerMiddleware(logger))

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        slog.Info("Handling request")
        w.Write([]byte("Hello, World!"))
    })

    http.ListenAndServe(":8080", r)
}
```

## Log Output

The logger produces beautifully colored output:

- **DEBUG**: Dark gray
- **INFO**: Cyan
- **WARN**: Light yellow
- **ERROR**: Light red (includes file and line number)

Example output:
```
[2025-10-10 13:45:23.123] INFO: Application started {}
[2025-10-10 13:45:23.456] ERROR: Database connection failed {
  "error": "connection refused",
  "file": "/path/to/file.go",
  "line": 42
}
```

## Advanced Usage

### Structured Logging

```go
slog.Info("User logged in",
    slog.String("user_id", "12345"),
    slog.String("ip", "192.168.1.1"),
    slog.Duration("session_duration", time.Hour),
)
```

### Context-Aware Logging

```go
logger := slog.With(
    slog.String("service", "auth"),
    slog.String("version", "1.0.0"),
)

logger.Info("Service initialized")
```

### Error Tracking

Error logs automatically include the file name and line number where the error was logged:

```go
slog.Error("Failed to process request",
    slog.String("error", err.Error()),
    slog.String("request_id", reqID),
)
```

## API Reference

### Handler

```go
func NewHandler(opts *slog.HandlerOptions) *Handler
```

Creates a new custom handler with the specified options. If `opts` is nil, default options are used.

### Logger Setup

```go
func SetupCustomLogger()
```

Creates and sets up a custom logger with INFO level as the default logger for the application.

### Middleware

```go
func NewLoggerMiddleware(log *slog.Logger) func(next http.Handler) http.Handler
```

Creates an HTTP middleware that logs:
- Request method, path, and remote address
- User agent and request ID
- Response status code and size
- Request duration

## Requirements

- Go 1.25.2 or higher
- `github.com/go-chi/chi/v5` (for middleware)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).

## Author

[corray333](https://github.com/corray333)
