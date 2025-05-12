package configurator

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
)

// Common errors
var (
	ErrInvalidConfig    = errors.New("invalid configuration: must be a pointer to a struct")
	ErrLoadFailed       = errors.New("failed to load configuration")
	ErrValidation       = errors.New("configuration validation failed")
	ErrFieldNotSettable = errors.New("field is not settable")
	ErrIncompatibleType = errors.New("incompatible type for field")
	ErrFieldNotFound    = errors.New("field not found in configuration")
)

// Validator validates a configuration
type Validator interface {
	Validate(cfg interface{}) error
}

// Configurator handles loading configuration from multiple sources
type Configurator struct {
	providers []Provider
	validator Validator
	logger    *slog.Logger
}

// New creates a new Configurator
func New(logger *slog.Logger) *Configurator {
	return &Configurator{
		providers: make([]Provider, 0),
		logger:    logger,
	}
}

// WithProvider adds a provider to the configurator
func (c *Configurator) WithProvider(provider Provider) *Configurator {
	c.providers = append(c.providers, provider)
	return c
}

// WithValidator sets the validator for the configurator
func (c *Configurator) WithValidator(validator Validator) *Configurator {
	c.validator = validator
	return c
}

// Load loads configuration from all registered providers into the provided config object
func (c *Configurator) Load(ctx context.Context, cfg interface{}) error {
	// Ensure cfg is a pointer to a struct
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrInvalidConfig
	}

	// Load configuration from providers
	for _, provider := range c.providers {
		if c.logger != nil {
			c.logger.Info("Loading configuration from provider", "provider", provider.Name())
		}
		if err := provider.Load(cfg); err != nil {
			return err
		}
	}

	// Validate the configuration if a validator is set
	if c.validator != nil {
		if err := c.validator.Validate(cfg); err != nil {
			return err
		}
	}

	return nil
}

// DefaultLoad provides a simplified way to load configuration
func DefaultLoad(ctx context.Context, configPath string, envPrefix string, cfg interface{}, logger *slog.Logger) error {
	configurator := New(logger)

	// Add default providers
	configurator.WithProvider(NewDefaultProvider())

	if configPath != "" {
		configurator.WithProvider(NewFileProvider(configPath))
	}

	if envPrefix != "" {
		configurator.WithProvider(NewEnvProvider(envPrefix))
	}

	return configurator.Load(ctx, cfg)
}
