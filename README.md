# Configurator

[![Go Reference](https://pkg.go.dev/badge/github.com/localrivet/configurator.svg)](https://pkg.go.dev/github.com/localrivet/configurator)
[![Go Report Card](https://goreportcard.com/badge/github.com/localrivet/configurator)](https://goreportcard.com/report/github.com/localrivet/configurator)

A flexible, extensible configuration management library for Go applications.

## Features

- Load configuration from multiple sources (files, environment variables, defaults, secrets)
- Support for JSON, YAML, and TOML file formats
- Tag-based and programmatic validation
- Support for any Go struct as a configuration object
- Type-safe configuration with automatic conversions
- Monitoring and observability hooks
- Easy to extend with custom providers and validators

## Installation

```bash
go get github.com/localrivet/configurator
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/localrivet/configurator"
)

// Define your configuration struct
type AppConfig struct {
    Server struct {
        Host string `json:"host" env:"SERVER_HOST" validate:"required"`
        Port int    `json:"port" env:"SERVER_PORT" validate:"range:1-65535"`
    } `json:"server"`
    Database struct {
        URL      string `json:"url" env:"DB_URL" validate:"required"`
        Username string `json:"username" env:"DB_USER"`
        Password string `json:"password" env:"DB_PASS" secret:"true"`
    } `json:"database"`
}

func main() {
    // Create a logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create a configuration object
    cfg := &AppConfig{}

    // Create a configurator with providers
    config := configurator.New(logger).
        WithProvider(configurator.NewDefaultProvider().
            WithDefault("Server.Port", 8080)).
        WithProvider(configurator.NewFileProvider("config.json")).
        WithProvider(configurator.NewEnvProvider("APP")).
        WithValidator(configurator.NewDefaultValidator())

    // Load configuration
    ctx := context.Background()
    if err := config.Load(ctx, cfg); err != nil {
        logger.Error("Failed to load configuration", "error", err)
        os.Exit(1)
    }

    // Use the configuration
    fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
```

### Different File Formats

```go
// JSON
configurator.NewFileProvider("config.json")
// or explicitly
configurator.NewJSONFileProvider("config.json")

// YAML
configurator.NewYAMLFileProvider("config.yaml")

// TOML
configurator.NewTOMLFileProvider("config.toml")

// Auto-detect format based on extension
configurator.NewFileProvider("config.yaml") // Will use YAML
```

### Tag-Based Validation

```go
// Define validation rules in struct tags
type Config struct {
    Port     int    `validate:"range:1-65535"`
    Name     string `validate:"required"`
    Count    int    `validate:"min:5"`
    MaxUsers int    `validate:"max:100"`
}

// The validator will automatically process these tags
validator := configurator.NewDefaultValidator()
config.WithValidator(validator)
```

### Programmatic Validation

```go
validator := configurator.NewDefaultValidator().
    AddRule("Server.Port", configurator.RangeRule(1, 65535)).
    AddRule("Database.URL", configurator.RequiredRule())

config.WithValidator(validator)
```

### Monitoring and Observability

```go
// Create an observable configurator
observableConfig := configurator.NewObservable(config).
    WithObserver(configurator.NewLoggingObserver(logger))

// Use the observable configurator instead of the regular one
if err := observableConfig.Load(ctx, cfg); err != nil {
    // handle error
}
```

### Creating Custom Providers

```go
type MyProvider struct {
    // your fields here
}

func (p *MyProvider) Name() string {
    return "my-provider"
}

func (p *MyProvider) Load(cfg interface{}) error {
    // your implementation here
    return nil
}

// Use it
config.WithProvider(&MyProvider{})
```

### Creating Custom Validators

```go
type MyValidator struct {
    // your fields here
}

func (v *MyValidator) Validate(cfg interface{}) error {
    // your implementation here
    return nil
}

// Use it
config.WithValidator(&MyValidator{})
```

### Creating Custom Observers

```go
type MyObserver struct {
    // your fields here
}

func (o *MyObserver) OnLoad(event configurator.LoadEvent) {
    // handle load event
}

func (o *MyObserver) OnValidate(event configurator.ValidationEvent) {
    // handle validation event
}

func (o *MyObserver) OnError(event configurator.ErrorEvent) {
    // handle error event
}

// Use it
observableConfig.WithObserver(&MyObserver{})
```

## Extensibility

The library is designed to be extensible. You can:

1. Create custom providers to load configuration from any source
2. Implement custom validators for domain-specific validation
3. Define custom validation rules for specific fields
4. Add monitoring through observers
5. Process configuration before and after loading

## License

MIT
