package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/localrivet/configurator"
)

// AppConfig represents the application configuration
type AppConfig struct {
	Server struct {
		Host string `json:"host" env:"SERVER_HOST" validate:"required"`
		Port int    `json:"port" env:"SERVER_PORT" validate:"range:1-65535"`
	} `json:"server"`
	Database struct {
		URL      string `json:"url" env:"DB_URL" validate:"required"`
		Username string `json:"username" env:"DB_USER" validate:"required"`
		Password string `json:"password" env:"DB_PASS" secret:"true" validate:"required"`
	} `json:"database"`
	Logging struct {
		Level  string `json:"level" env:"LOG_LEVEL" validate:"required"`
		Format string `json:"format" env:"LOG_FORMAT" validate:"required"`
	} `json:"logging"`
}

func main() {
	// Create a new logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a configuration object
	cfg := &AppConfig{}

	// Set up default values with a DefaultProvider
	defaultProvider := configurator.NewDefaultProvider().
		WithDefault("Server.Port", 8080).
		WithDefault("Logging.Level", "info").
		WithDefault("Logging.Format", "json")

	// Set up a validator
	validator := configurator.NewDefaultValidator()
	// The validator will automatically use struct tags (validate:"required", etc.)

	// Create a new configurator
	config := configurator.New(logger).
		WithProvider(defaultProvider).
		WithProvider(configurator.NewYAMLFileProvider("config.yaml")). // Use YAML provider
		WithProvider(configurator.NewEnvProvider("APP")).
		WithValidator(validator)

	// Wrap with observable configurator for monitoring
	observableConfig := configurator.NewObservable(config).
		WithObserver(configurator.NewLoggingObserver(logger))

	// Load configuration
	ctx := context.Background()
	if err := observableConfig.Load(ctx, cfg); err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Print the configuration
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database URL: %s\n", cfg.Database.URL)
	fmt.Printf("Logging: %s/%s\n", cfg.Logging.Level, cfg.Logging.Format)

	// Create example config files in different formats
	createExampleConfigFiles(cfg, logger)
}

// createExampleConfigFiles creates example configuration files in different formats
func createExampleConfigFiles(cfg *AppConfig, logger *slog.Logger) {
	// Set some example values
	cfg.Server.Host = "localhost"
	cfg.Server.Port = 8080
	cfg.Database.URL = "mysql://localhost:3306/mydb"
	cfg.Database.Username = "user"
	cfg.Database.Password = "password"
	cfg.Logging.Level = "info"
	cfg.Logging.Format = "json"

	// Save in JSON format
	if err := configurator.SaveToFile(cfg, "example/config.json", configurator.FormatJSON); err != nil {
		logger.Error("Failed to save JSON config", "error", err)
	}

	// Save in YAML format
	if err := configurator.SaveToFile(cfg, "example/config.yaml", configurator.FormatYAML); err != nil {
		logger.Error("Failed to save YAML config", "error", err)
	}

	// Save in TOML format
	if err := configurator.SaveToFile(cfg, "example/config.toml", configurator.FormatTOML); err != nil {
		logger.Error("Failed to save TOML config", "error", err)
	}

	logger.Info("Created example configuration files in JSON, YAML, and TOML formats")
}
