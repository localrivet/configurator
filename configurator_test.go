package configurator

import (
	"context"
	"os"
	"testing"

	"log/slog"
)

// TestConfig is a test configuration structure
type TestConfig struct {
	Server struct {
		Host string `json:"host" env:"SERVER_HOST" validate:"required"`
		Port int    `json:"port" env:"SERVER_PORT" validate:"range:1-65535"`
	} `json:"server"`
	Database struct {
		URL      string `json:"url" env:"DB_URL" validate:"required"`
		Username string `json:"username" env:"DB_USER" validate:"required"`
		Password string `json:"password" env:"DB_PASS" secret:"true" validate:"required"`
	} `json:"database"`
}

// TestObserver implements the Observer interface for testing
type TestObserver struct {
	LoadCalled      bool
	ValidateCalled  bool
	ErrorCalled     bool
	ValidationValid bool
}

func (o *TestObserver) OnLoad(event LoadEvent) {
	o.LoadCalled = true
}

func (o *TestObserver) OnValidate(event ValidationEvent) {
	o.ValidateCalled = true
	o.ValidationValid = event.Valid
}

func (o *TestObserver) OnError(event ErrorEvent) {
	o.ErrorCalled = true
}

func TestDefaultProvider(t *testing.T) {
	cfg := &TestConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set up a default provider
	defaultProvider := NewDefaultProvider().
		WithDefault("Server.Host", "localhost").
		WithDefault("Server.Port", 8080).
		WithDefault("Database.URL", "mysql://localhost:3306/testdb").
		WithDefault("Database.Username", "testuser").
		WithDefault("Database.Password", "testpass")

	// Create configurator with default provider
	configurator := New(logger).WithProvider(defaultProvider)

	// Load configuration
	err := configurator.Load(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify values
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected Server.Host to be 'localhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected Server.Port to be 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL != "mysql://localhost:3306/testdb" {
		t.Errorf("Expected Database.URL to be 'mysql://localhost:3306/testdb', got '%s'", cfg.Database.URL)
	}
	if cfg.Database.Username != "testuser" {
		t.Errorf("Expected Database.Username to be 'testuser', got '%s'", cfg.Database.Username)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("Expected Database.Password to be 'testpass', got '%s'", cfg.Database.Password)
	}
}

func TestEnvProvider(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_SERVER_HOST", "testhost")
	os.Setenv("TEST_SERVER_PORT", "9090")
	os.Setenv("TEST_DB_URL", "postgres://localhost:5432/testdb")
	os.Setenv("TEST_DB_USER", "pguser")
	os.Setenv("TEST_DB_PASS", "pgpass")
	defer func() {
		os.Unsetenv("TEST_SERVER_HOST")
		os.Unsetenv("TEST_SERVER_PORT")
		os.Unsetenv("TEST_DB_URL")
		os.Unsetenv("TEST_DB_USER")
		os.Unsetenv("TEST_DB_PASS")
	}()

	cfg := &TestConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create configurator with env provider
	configurator := New(logger).WithProvider(NewEnvProvider("TEST"))

	// Load configuration
	err := configurator.Load(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify values
	if cfg.Server.Host != "testhost" {
		t.Errorf("Expected Server.Host to be 'testhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected Server.Port to be 9090, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL != "postgres://localhost:5432/testdb" {
		t.Errorf("Expected Database.URL to be 'postgres://localhost:5432/testdb', got '%s'", cfg.Database.URL)
	}
	if cfg.Database.Username != "pguser" {
		t.Errorf("Expected Database.Username to be 'pguser', got '%s'", cfg.Database.Username)
	}
	if cfg.Database.Password != "pgpass" {
		t.Errorf("Expected Database.Password to be 'pgpass', got '%s'", cfg.Database.Password)
	}
}

func TestObservability(t *testing.T) {
	cfg := &TestConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set up a default provider
	defaultProvider := NewDefaultProvider().
		WithDefault("Server.Host", "localhost").
		WithDefault("Server.Port", 8080).
		WithDefault("Database.URL", "mysql://localhost:3306/testdb").
		WithDefault("Database.Username", "testuser").
		WithDefault("Database.Password", "testpass")

	// Create a test observer
	observer := &TestObserver{}

	// Create configurator with default provider and wrap with observable
	configurator := New(logger).WithProvider(defaultProvider)
	observableConfig := NewObservable(configurator).WithObserver(observer)

	// Load configuration
	err := observableConfig.Load(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify observer was called
	if !observer.LoadCalled {
		t.Error("Observer's OnLoad was not called")
	}
	if !observer.ValidateCalled {
		t.Error("Observer's OnValidate was not called")
	}
	if !observer.ValidationValid {
		t.Error("Validation should have been successful")
	}
}

func TestFileProvider(t *testing.T) {
	// Create a temporary JSON config file
	jsonConfig := `{
		"server": {
			"host": "filehost",
			"port": 7070
		},
		"database": {
			"url": "mysql://filehost:3306/filedb",
			"username": "fileuser",
			"password": "filepass"
		}
	}`
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(jsonConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	cfg := &TestConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create configurator with file provider
	configurator := New(logger).WithProvider(NewJSONFileProvider(tmpFile.Name()))

	// Load configuration
	err = configurator.Load(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify values
	if cfg.Server.Host != "filehost" {
		t.Errorf("Expected Server.Host to be 'filehost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 7070 {
		t.Errorf("Expected Server.Port to be 7070, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL != "mysql://filehost:3306/filedb" {
		t.Errorf("Expected Database.URL to be 'mysql://filehost:3306/filedb', got '%s'", cfg.Database.URL)
	}
	if cfg.Database.Username != "fileuser" {
		t.Errorf("Expected Database.Username to be 'fileuser', got '%s'", cfg.Database.Username)
	}
	if cfg.Database.Password != "filepass" {
		t.Errorf("Expected Database.Password to be 'filepass', got '%s'", cfg.Database.Password)
	}
}

func TestValidation(t *testing.T) {
	// Test with missing required fields
	cfg := &TestConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set up a default provider with incomplete config
	defaultProvider := NewDefaultProvider().
		WithDefault("Server.Port", 8080) // Missing required fields

	// Create configurator with default provider and validator
	configurator := New(logger).
		WithProvider(defaultProvider).
		WithValidator(NewDefaultValidator())

	// Load configuration - should fail validation
	err := configurator.Load(context.Background(), cfg)
	if err == nil {
		t.Fatal("Expected validation to fail, but it passed")
	}

	// Now provide all required fields
	defaultProvider = NewDefaultProvider().
		WithDefault("Server.Host", "localhost").
		WithDefault("Server.Port", 8080).
		WithDefault("Database.URL", "mysql://localhost:3306/testdb").
		WithDefault("Database.Username", "testuser").
		WithDefault("Database.Password", "testpass")

	// Create configurator with complete config
	configurator = New(logger).
		WithProvider(defaultProvider).
		WithValidator(NewDefaultValidator())

	// Load configuration - should pass validation
	err = configurator.Load(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Validation failed when it should have passed: %v", err)
	}
}
