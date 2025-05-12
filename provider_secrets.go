package configurator

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// SecretsProvider loads configuration from mounted secrets
type SecretsProvider struct {
	MountPath string
}

// NewSecretsProvider creates a new secrets provider
func NewSecretsProvider(mountPath string) *SecretsProvider {
	return &SecretsProvider{
		MountPath: mountPath,
	}
}

// Name returns the provider name
func (p *SecretsProvider) Name() string {
	return "secrets"
}

// Load loads configuration from mounted secrets
func (p *SecretsProvider) Load(cfg interface{}) error {
	if p.MountPath == "" || !dirExists(p.MountPath) {
		return nil
	}

	// Walk through the directory entries
	entries, err := os.ReadDir(p.MountPath)
	if err != nil {
		return fmt.Errorf("failed to read secrets directory: %w", err)
	}

	// Process each entry
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Get the file path
		filePath := p.MountPath + "/" + entry.Name()

		// Read the file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read secret file %s: %w", filePath, err)
		}

		// The file name is the key, the content is the value
		secretKey := entry.Name()
		secretValue := string(content)

		// Apply the secret value based on the key
		if err := applySecret(cfg, secretKey, secretValue); err != nil {
			// Log error but continue with other secrets
			fmt.Printf("Warning: failed to apply secret %s: %v\n", secretKey, err)
		}
	}
	return nil
}

// applySecret applies a secret value to a configuration field
func applySecret(cfg interface{}, secretKey, secretValue string) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrInvalidConfig
	}

	// Convert secret key to field path
	// Example: "DB_PASSWORD" -> "Database.Password"
	// This is a simple implementation - more sophisticated mapping might be needed
	fieldPath := secretKeyToFieldPath(secretKey)

	// Try to find and set the field
	field, err := getFieldValue(cfg, fieldPath)
	if err != nil {
		return err
	}

	// Set the field value
	return setFieldValue(field, secretValue)
}

// secretKeyToFieldPath converts a secret key to a field path
// Example: "DB_PASSWORD" -> "Database.Password"
func secretKeyToFieldPath(key string) string {
	// This is a simple implementation - adjust as needed
	parts := strings.Split(key, "_")
	for i, part := range parts {
		if i == 0 {
			parts[i] = strings.Title(strings.ToLower(part))
		} else {
			parts[i] = strings.Title(strings.ToLower(part))
		}
	}

	return strings.Join(parts, ".")
}
