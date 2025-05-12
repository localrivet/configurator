package configurator

import (
	"reflect"
	"strconv"
	"strings"
)

// DefaultProvider provides default configuration values
type DefaultProvider struct {
	// DefaultValues maps field paths to default values
	DefaultValues map[string]interface{}
}

// NewDefaultProvider creates a new default provider
func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{
		DefaultValues: make(map[string]interface{}),
	}
}

// WithDefault adds a default value for a specific field path
// fieldPath format: "Server.Port" for nested fields
func (p *DefaultProvider) WithDefault(fieldPath string, value interface{}) *DefaultProvider {
	p.DefaultValues[fieldPath] = value
	return p
}

// Name returns the provider name
func (p *DefaultProvider) Name() string {
	return "default"
}

// Load loads default values into the configuration
func (p *DefaultProvider) Load(cfg interface{}) error {
	if len(p.DefaultValues) == 0 {
		return nil // Nothing to do
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrInvalidConfig
	}

	// Apply default values
	for fieldPath, defaultValue := range p.DefaultValues {
		// Check if field exists and is settable
		field, err := getFieldByPath(v.Elem(), fieldPath)
		if err != nil {
			continue // Skip fields that don't exist
		}

		// Skip if field is already set
		if isZeroValue(field) {
			// Set default value if compatible
			if err := setFieldValue(field, defaultValue); err != nil {
				continue // Skip incompatible values
			}
		}
	}

	return nil
}

// isZeroValue checks if a field has its zero/empty value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	default:
		return false
	}
}

// setFieldValue sets a value on a field, converting types if necessary
func setFieldValue(field reflect.Value, value interface{}) error {
	// Skip if field is not settable
	if !field.CanSet() {
		return ErrFieldNotSettable
	}

	// Get the value as reflect.Value
	val := reflect.ValueOf(value)

	// Convert value if needed and possible
	if field.Kind() != val.Kind() {
		var converted bool
		// Special case: string conversion to other types
		if val.Kind() == reflect.String {
			converted = convertFromString(field, val.String())
		} else {
			// Try other conversions like int to float, etc.
			converted = tryConversion(field, val)
		}

		if !converted {
			return ErrIncompatibleType
		}
	} else {
		// Direct assignment for matching types
		field.Set(val)
	}

	return nil
}

// Helper functions for setting field values
func convertFromString(field reflect.Value, strValue string) bool {
	switch field.Kind() {
	case reflect.Bool:
		if b, err := strconv.ParseBool(strValue); err == nil {
			field.SetBool(b)
			return true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			if field.OverflowInt(i) {
				return false
			}
			field.SetInt(i)
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(strValue, 10, 64); err == nil {
			if field.OverflowUint(u) {
				return false
			}
			field.SetUint(u)
			return true
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(strValue, 64); err == nil {
			if field.OverflowFloat(f) {
				return false
			}
			field.SetFloat(f)
			return true
		}
	}
	return false
}

func tryConversion(field reflect.Value, val reflect.Value) bool {
	switch {
	// Int -> other numeric types
	case val.Kind() >= reflect.Int && val.Kind() <= reflect.Int64:
		i := val.Int()
		switch field.Kind() {
		case reflect.Float32, reflect.Float64:
			if !field.OverflowFloat(float64(i)) {
				field.SetFloat(float64(i))
				return true
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if i >= 0 && !field.OverflowUint(uint64(i)) {
				field.SetUint(uint64(i))
				return true
			}
		}
	// Uint -> other numeric types
	case val.Kind() >= reflect.Uint && val.Kind() <= reflect.Uint64:
		u := val.Uint()
		switch field.Kind() {
		case reflect.Float32, reflect.Float64:
			if !field.OverflowFloat(float64(u)) {
				field.SetFloat(float64(u))
				return true
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if !field.OverflowInt(int64(u)) {
				field.SetInt(int64(u))
				return true
			}
		}
	// Float -> other numeric types (with truncation)
	case val.Kind() == reflect.Float32 || val.Kind() == reflect.Float64:
		f := val.Float()
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if !field.OverflowInt(int64(f)) {
				field.SetInt(int64(f))
				return true
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if f >= 0 && !field.OverflowUint(uint64(f)) {
				field.SetUint(uint64(f))
				return true
			}
		}
	}
	return false
}

// getFieldByPath gets a field by its path (e.g., "Server.Port")
func getFieldByPath(structValue reflect.Value, path string) (reflect.Value, error) {
	// Split the path into parts
	parts := strings.Split(path, ".")

	// Start with the struct value
	value := structValue

	// Navigate through the struct fields
	for i, part := range parts {
		// Get the field by name
		field := value.FieldByName(part)
		if !field.IsValid() {
			return reflect.Value{}, ErrFieldNotFound
		}

		// If this is the last part of the path, return the field
		if i == len(parts)-1 {
			return field, nil
		}

		// If field is a pointer, get the underlying value
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return reflect.Value{}, ErrFieldNotFound
			}
			field = field.Elem()
		}

		// If the next level isn't a struct, we can't continue
		if field.Kind() != reflect.Struct {
			return reflect.Value{}, ErrFieldNotFound
		}

		// Continue with the nested struct
		value = field
	}

	// This should never happen if the function is used correctly
	return reflect.Value{}, ErrFieldNotFound
}
