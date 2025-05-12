package configurator

import (
	"os"
)

// Provider represents a configuration provider
type Provider interface {
	// Name returns the provider name
	Name() string
	// Load loads configuration into the provided interface
	Load(into interface{}) error
}

// Helper functions

// fileExists checks if a file exists
func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
