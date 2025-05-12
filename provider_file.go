package configurator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// FileFormat represents the format of a configuration file
type FileFormat int

const (
	// FormatJSON represents JSON format
	FormatJSON FileFormat = iota
	// FormatYAML represents YAML format
	FormatYAML
	// FormatTOML represents TOML format
	FormatTOML
	// FormatAuto automatically detects the format based on file extension
	FormatAuto
)

// FileProvider loads configuration from a file
type FileProvider struct {
	Path   string
	Format FileFormat
}

// NewFileProvider creates a new file provider with format auto-detection
func NewFileProvider(path string) *FileProvider {
	return &FileProvider{
		Path:   path,
		Format: FormatAuto,
	}
}

// NewJSONFileProvider creates a new JSON file provider
func NewJSONFileProvider(path string) *FileProvider {
	return &FileProvider{
		Path:   path,
		Format: FormatJSON,
	}
}

// NewYAMLFileProvider creates a new YAML file provider
func NewYAMLFileProvider(path string) *FileProvider {
	return &FileProvider{
		Path:   path,
		Format: FormatYAML,
	}
}

// NewTOMLFileProvider creates a new TOML file provider
func NewTOMLFileProvider(path string) *FileProvider {
	return &FileProvider{
		Path:   path,
		Format: FormatTOML,
	}
}

// Name returns the provider name
func (p *FileProvider) Name() string {
	return "file"
}

// Load loads configuration from a file
func (p *FileProvider) Load(cfg interface{}) error {
	if p.Path == "" {
		return nil
	}

	// Check if file exists
	if !fileExists(p.Path) {
		return fmt.Errorf("configuration file not found: %s", p.Path)
	}

	// Read file content
	data, err := os.ReadFile(p.Path)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Determine format if auto-detection is enabled
	format := p.Format
	if format == FormatAuto {
		format = detectFormatFromExtension(p.Path)
	}

	// Decode based on format
	switch format {
	case FormatJSON:
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to decode JSON configuration: %w", err)
		}
	case FormatYAML:
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to decode YAML configuration: %w", err)
		}
	case FormatTOML:
		if err := toml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to decode TOML configuration: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format")
	}

	return nil
}

// detectFormatFromExtension detects the file format from the file extension
func detectFormatFromExtension(path string) FileFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	case ".toml":
		return FormatTOML
	default:
		// Default to JSON if unknown
		return FormatJSON
	}
}

// SaveToFile is a utility function to save any config to a file with the given format
func SaveToFile(cfg interface{}, path string, format FileFormat) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for configuration file: %w", err)
	}

	// If format is auto, detect from extension
	if format == FormatAuto {
		format = detectFormatFromExtension(path)
	}

	var data []byte
	var err error

	// Encode based on format
	switch format {
	case FormatJSON:
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal configuration to JSON: %w", err)
		}
	case FormatYAML:
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal configuration to YAML: %w", err)
		}
	case FormatTOML:
		// TOML doesn't have a direct way to marshal to bytes, so we'll use a temporary file
		tmpFile, err := os.CreateTemp("", "config-*.toml")
		if err != nil {
			return fmt.Errorf("failed to create temporary file for TOML encoding: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if err := toml.NewEncoder(tmpFile).Encode(cfg); err != nil {
			return fmt.Errorf("failed to marshal configuration to TOML: %w", err)
		}

		// Read the encoded content
		if _, err := tmpFile.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek in temporary file: %w", err)
		}
		data, err = os.ReadFile(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("failed to read encoded TOML: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// LoadFromFile is a utility function to load any config from a file
func LoadFromFile(cfg interface{}, path string) error {
	provider := NewFileProvider(path)
	return provider.Load(cfg)
}

// FindConfigFile searches for a config file with the given name in the current directory and parent directories
func FindConfigFile(filename string) (string, error) {
	// Start with current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Try to find the file in current or any parent directory
	for {
		configPath := filepath.Join(dir, filename)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("configuration file %s not found", filename)
}
