package configurator

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// EnvProvider loads configuration from environment variables
type EnvProvider struct {
	Prefix string
}

// NewEnvProvider creates a new environment provider
func NewEnvProvider(prefix string) *EnvProvider {
	return &EnvProvider{
		Prefix: prefix,
	}
}

// Name returns the provider name
func (p *EnvProvider) Name() string {
	return "environment"
}

// Load loads configuration from environment variables
func (p *EnvProvider) Load(cfg interface{}) error {
	return applyEnvVariables(cfg, p.Prefix)
}

// applyEnvVariables applies environment variables to the configuration
func applyEnvVariables(cfg interface{}, prefix string) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrInvalidConfig
	}
	return processStruct(v.Elem(), prefix, "")
}

// processStruct processes a struct's fields for environment variables
func processStruct(v reflect.Value, prefix, parent string) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the field tag for environment variable name
		var envTag string
		tag := fieldType.Tag.Get("env")
		if tag != "" {
			envTag = tag
		} else {
			// Default to field name if no tag
			envTag = fieldType.Name
		}

		// For nested structs, build the proper path
		fieldName := fieldType.Name
		path := fieldName
		if parent != "" {
			path = parent + "_" + fieldName
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.Struct:
			// Recurse into nested structs
			if err := processStruct(field, prefix, path); err != nil {
				return err
			}
			continue
		case reflect.Ptr:
			if field.IsNil() && field.Type().Elem().Kind() == reflect.Struct {
				// Create a new struct and set it
				newStruct := reflect.New(field.Type().Elem())
				field.Set(newStruct)
				// Process the new struct
				if err := processStruct(newStruct.Elem(), prefix, path); err != nil {
					return err
				}
			} else if !field.IsNil() && field.Type().Elem().Kind() == reflect.Struct {
				// Process the existing struct
				if err := processStruct(field.Elem(), prefix, path); err != nil {
					return err
				}
			}
			continue
		}

		// Construct the environment variable name
		envVarName := strings.ToUpper(envTag)
		if prefix != "" {
			envVarName = prefix + "_" + envVarName
		}

		// Get the value from environment
		envValue := os.Getenv(envVarName)
		if envValue == "" {
			continue
		}

		// Apply the value based on the field type
		if err := applyValueToField(field, envValue); err != nil {
			return fmt.Errorf("failed to apply environment variable %s: %w", envVarName, err)
		}
	}
	return nil
}

// applyValueToField applies a value to a field based on its type
func applyValueToField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type().String() == "time.Duration" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			if field.OverflowInt(intValue) {
				return fmt.Errorf("value %d overflows field %s", intValue, field.Type().String())
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		if field.OverflowUint(uintValue) {
			return fmt.Errorf("value %d overflows field %s", uintValue, field.Type().String())
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if field.OverflowFloat(floatValue) {
			return fmt.Errorf("value %f overflows field %s", floatValue, field.Type().String())
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Handle string slices (comma-separated values)
			values := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), 0, len(values))
			for _, v := range values {
				v = strings.TrimSpace(v)
				if v != "" {
					slice = reflect.Append(slice, reflect.ValueOf(v))
				}
			}
			field.Set(slice)
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type().String())
	}
	return nil
}
