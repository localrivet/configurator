# Product Requirements Document: Universal Configuration Library

## 1. Overview

The Universal Configuration Library is designed to provide a robust, flexible way to manage application configuration from multiple sources. It enables developers to use their own configuration objects while benefiting from validation, external files, environment variables, secrets, and dynamic configuration.

## 2. Goals and Objectives

- Allow users to use their own configuration structs from any package
- Provide a consistent interface for loading configuration from multiple sources
- Enable validation of configuration objects
- Support different file formats, environment variables, and secrets management
- Minimize dependencies for easy integration

## 3. Core Features

### 3.1 Configuration Loading

- **Load from files**: Support JSON, YAML, TOML, and other formats
- **Environment variables**: Map env vars to struct fields with customizable prefix
- **Secrets management**: Load from mounted secrets, Vault, AWS Secrets Manager
- **Dynamic configuration**: Support for runtime configuration changes
- **Default values**: Fallback to sensible defaults

### 3.2 Configuration Validation

- Struct tag-based validation rules
- Programmatic validation via validator interface
- Nested validation for complex objects
- Custom validation functions

### 3.3 Provider Interface

- Pluggable provider architecture
- Default providers for common sources
- User-extensible for custom sources

## 4. Technical Requirements

### 4.1 API Design

```go
// Core interface for loading configuration
type Provider interface {
    Name() string
    Load(cfg interface{}) error
}

// Validator interface
type Validator interface {
    Validate(cfg interface{}) error
}

// Main configurator
type Configurator struct {
    // Manages providers and validation
}

// Create a new configurator
func New() *Configurator

// Add providers
func (c *Configurator) WithProvider(provider Provider) *Configurator

// Set validator
func (c *Configurator) WithValidator(validator Validator) *Configurator

// Load configuration
func (c *Configurator) Load(cfg interface{}) error
```

### 4.2 Provider Implementations

- File provider (JSON, YAML, TOML)
- Environment variable provider
- Secrets provider (file-based, Vault, AWS)
- Default value provider
- Dynamic value provider

## 5. Use Cases

### 5.1 Basic Usage

```go
type MyConfig struct {
    Server struct {
        Host string `json:"host" env:"SERVER_HOST" validate:"required"`
        Port int    `json:"port" env:"SERVER_PORT" validate:"range:1,65535"`
    }
    Database struct {
        URL      string `json:"url" env:"DB_URL" validate:"required"`
        Username string `json:"username" env:"DB_USER"`
        Password string `json:"password" env:"DB_PASS" secret:"true"`
    }
}

// Create config
cfg := &MyConfig{}

// Create configurator and load
configurator := config.New().
    WithProvider(config.NewFileProvider("config.json")).
    WithProvider(config.NewEnvProvider("APP")).
    WithProvider(config.NewSecretsProvider("/run/secrets")).
    WithValidator(config.NewDefaultValidator())

err := configurator.Load(cfg)
```

### 5.2 Advanced Usage

```go
// Custom provider
type APIProvider struct {
    // implementation details
}

func (p *APIProvider) Name() string { return "api" }
func (p *APIProvider) Load(cfg interface{}) error {
    // Implementation to load from API
}

// Custom validator
type MyValidator struct {
    // implementation details
}

func (v *MyValidator) Validate(cfg interface{}) error {
    // Custom validation logic
}

// Advanced configuration
configurator := config.New().
    WithProvider(config.NewFileProvider("config.json")).
    WithProvider(config.NewEnvProvider("APP")).
    WithProvider(&APIProvider{}).
    WithValidator(&MyValidator{})

err := configurator.Load(cfg)
```

## 6. Security Considerations

- Sensitive values should not be logged
- Support for encrypted configuration files
- Secure handling of secrets
- Protection against leaking sensitive values in error messages

## 7. Implementation Plan

### Phase 1: Core Framework

- Base interfaces (Provider, Validator)
- Configurator implementation
- File provider (JSON)
- Environment variable provider
- Basic validation

### Phase 2: Extended Features

- Additional file formats (YAML, TOML)
- Secrets provider
- Enhanced validation
- Documentation and examples

### Phase 3: Advanced Features

- Dynamic configuration
- Custom providers
- Monitoring and observability hooks
- Performance optimizations
