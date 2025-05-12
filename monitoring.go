package configurator

import (
	"context"
	"log/slog"
	"reflect"
	"time"
)

// Observer defines the interface for configuration observers
type Observer interface {
	// OnLoad is called after configuration is loaded
	OnLoad(event LoadEvent)
	// OnValidate is called after validation
	OnValidate(event ValidationEvent)
	// OnError is called when an error occurs
	OnError(event ErrorEvent)
}

// Event is the base interface for all events
type Event interface {
	// Timestamp returns the time when the event occurred
	Timestamp() time.Time
}

// LoadEvent represents a configuration load event
type LoadEvent struct {
	// When is the time when the event occurred
	When time.Time
	// Provider is the name of the provider that loaded the configuration
	Provider string
	// ConfigType is the type of the configuration object
	ConfigType string
	// Duration is how long the load operation took
	Duration time.Duration
}

// Timestamp returns the time when the event occurred
func (e LoadEvent) Timestamp() time.Time {
	return e.When
}

// ValidationEvent represents a validation event
type ValidationEvent struct {
	// When is the time when the event occurred
	When time.Time
	// Valid indicates whether the validation succeeded
	Valid bool
	// FailedRules is a list of rules that failed
	FailedRules []string
	// Duration is how long the validation took
	Duration time.Duration
}

// Timestamp returns the time when the event occurred
func (e ValidationEvent) Timestamp() time.Time {
	return e.When
}

// ErrorEvent represents an error event
type ErrorEvent struct {
	// When is the time when the event occurred
	When time.Time
	// Operation is the operation that failed
	Operation string
	// Error is the error that occurred
	Error error
}

// Timestamp returns the time when the event occurred
func (e ErrorEvent) Timestamp() time.Time {
	return e.When
}

// ObservableConfigurator extends Configurator with observability features
type ObservableConfigurator struct {
	*Configurator
	observers []Observer
}

// NewObservable creates a new ObservableConfigurator
func NewObservable(configurator *Configurator) *ObservableConfigurator {
	return &ObservableConfigurator{
		Configurator: configurator,
		observers:    make([]Observer, 0),
	}
}

// WithObserver adds an observer to the configurator
func (c *ObservableConfigurator) WithObserver(observer Observer) *ObservableConfigurator {
	c.observers = append(c.observers, observer)
	return c
}

// Load loads the configuration and notifies observers
func (c *ObservableConfigurator) Load(ctx context.Context, cfg interface{}) error {
	startTime := time.Now()
	var provider string

	// Get the type name of the config object
	cfgType := getTypeName(cfg)

	// Call the underlying Load method
	err := c.Configurator.Load(ctx, cfg)

	// Calculate duration
	duration := time.Since(startTime)

	if err != nil {
		// Notify observers of error
		c.notifyError("Load", err)
		return err
	}

	// Notify observers of successful load
	c.notifyLoad(provider, cfgType, duration)

	// Notify validation success (this would be more detailed in a real implementation)
	c.notifyValidation(true, nil, duration)

	return nil
}

// notifyLoad notifies observers of a load event
func (c *ObservableConfigurator) notifyLoad(provider, configType string, duration time.Duration) {
	event := LoadEvent{
		When:       time.Now(),
		Provider:   provider,
		ConfigType: configType,
		Duration:   duration,
	}

	for _, observer := range c.observers {
		observer.OnLoad(event)
	}
}

// notifyValidation notifies observers of a validation event
func (c *ObservableConfigurator) notifyValidation(valid bool, failedRules []string, duration time.Duration) {
	event := ValidationEvent{
		When:        time.Now(),
		Valid:       valid,
		FailedRules: failedRules,
		Duration:    duration,
	}

	for _, observer := range c.observers {
		observer.OnValidate(event)
	}
}

// notifyError notifies observers of an error event
func (c *ObservableConfigurator) notifyError(operation string, err error) {
	event := ErrorEvent{
		When:      time.Now(),
		Operation: operation,
		Error:     err,
	}

	for _, observer := range c.observers {
		observer.OnError(event)
	}
}

// getTypeName returns the type name of an object
func getTypeName(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	return reflect.TypeOf(obj).String()
}

// LoggingObserver is an Observer that logs events
type LoggingObserver struct {
	logger *slog.Logger
}

// NewLoggingObserver creates a new LoggingObserver
func NewLoggingObserver(logger *slog.Logger) *LoggingObserver {
	return &LoggingObserver{
		logger: logger,
	}
}

// OnLoad logs load events
func (o *LoggingObserver) OnLoad(event LoadEvent) {
	o.logger.Info("Configuration loaded",
		"provider", event.Provider,
		"configType", event.ConfigType,
		"duration", event.Duration.String())
}

// OnValidate logs validation events
func (o *LoggingObserver) OnValidate(event ValidationEvent) {
	if event.Valid {
		o.logger.Info("Configuration validated successfully",
			"duration", event.Duration.String())
	} else {
		o.logger.Error("Configuration validation failed",
			"failedRules", event.FailedRules,
			"duration", event.Duration.String())
	}
}

// OnError logs error events
func (o *LoggingObserver) OnError(event ErrorEvent) {
	o.logger.Error("Configuration error",
		"operation", event.Operation,
		"error", event.Error.Error())
}
