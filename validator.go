package configurator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ValidationTagName is the tag name for validation rules
const ValidationTagName = "validate"

// DefaultValidator provides basic validation for configuration objects
type DefaultValidator struct {
	// Rules maps field paths to validation functions
	Rules map[string]func(interface{}) error
	// UseTagValidation indicates whether to use tag-based validation
	UseTagValidation bool
}

// NewDefaultValidator creates a new default validator
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		Rules:            make(map[string]func(interface{}) error),
		UseTagValidation: true,
	}
}

// AddRule adds a validation rule for a specific field path
// fieldPath format: "Server.Port" for nested fields
func (v *DefaultValidator) AddRule(fieldPath string, rule func(interface{}) error) *DefaultValidator {
	v.Rules[fieldPath] = rule
	return v
}

// DisableTagValidation disables tag-based validation
func (v *DefaultValidator) DisableTagValidation() *DefaultValidator {
	v.UseTagValidation = false
	return v
}

// EnableTagValidation enables tag-based validation
func (v *DefaultValidator) EnableTagValidation() *DefaultValidator {
	v.UseTagValidation = true
	return v
}

// Validate validates the configuration
func (v *DefaultValidator) Validate(cfg interface{}) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Apply explicit validation rules
	for fieldPath, rule := range v.Rules {
		value, err := getFieldValue(cfg, fieldPath)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		if err := rule(value.Interface()); err != nil {
			return fmt.Errorf("validation failed for field %s: %w", fieldPath, err)
		}
	}

	// Apply tag-based validation if enabled
	if v.UseTagValidation {
		if err := v.validateTags(cfg); err != nil {
			return err
		}
	}

	return nil
}

// validateTags validates fields based on their tags
func (v *DefaultValidator) validateTags(cfg interface{}) error {
	value := reflect.ValueOf(cfg)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil // Only validate structs
	}

	// Process struct fields
	return v.validateStructFields(value, "")
}

// validateStructFields validates all fields in a struct recursively
func (v *DefaultValidator) validateStructFields(value reflect.Value, prefix string) error {
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Build the field path
		fieldPath := fieldType.Name
		if prefix != "" {
			fieldPath = prefix + "." + fieldPath
		}

		// Process tag validation
		tag := fieldType.Tag.Get(ValidationTagName)
		if tag != "" {
			if err := v.validateFieldByTag(field, fieldPath, tag); err != nil {
				return err
			}
		}

		// Recursively validate nested structs
		switch {
		case field.Kind() == reflect.Struct:
			if err := v.validateStructFields(field, fieldPath); err != nil {
				return err
			}
		case field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct:
			if err := v.validateStructFields(field.Elem(), fieldPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFieldByTag validates a field based on its validation tag
func (v *DefaultValidator) validateFieldByTag(field reflect.Value, fieldPath, tag string) error {
	// Process multiple validation rules (comma-separated)
	rules := strings.Split(tag, ",")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		// Parse the rule
		parts := strings.SplitN(rule, ":", 2)
		ruleName := parts[0]

		// Apply appropriate validation based on rule name
		switch ruleName {
		case "required":
			if err := RequiredRule()(field.Interface()); err != nil {
				return fmt.Errorf("validation failed for field %s: %w", fieldPath, err)
			}
		case "range":
			if len(parts) < 2 {
				return fmt.Errorf("invalid range rule for field %s: missing range values", fieldPath)
			}

			// Parse range values
			rangeValues := strings.Split(parts[1], "-")
			if len(rangeValues) != 2 {
				return fmt.Errorf("invalid range format for field %s: expected min-max", fieldPath)
			}

			min, err := strconv.ParseInt(rangeValues[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid range minimum for field %s: %w", fieldPath, err)
			}

			max, err := strconv.ParseInt(rangeValues[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid range maximum for field %s: %w", fieldPath, err)
			}

			if err := RangeRule(min, max)(field.Interface()); err != nil {
				return fmt.Errorf("validation failed for field %s: %w", fieldPath, err)
			}
		case "min":
			if len(parts) < 2 {
				return fmt.Errorf("invalid min rule for field %s: missing value", fieldPath)
			}

			min, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid min value for field %s: %w", fieldPath, err)
			}

			if err := MinRule(min)(field.Interface()); err != nil {
				return fmt.Errorf("validation failed for field %s: %w", fieldPath, err)
			}
		case "max":
			if len(parts) < 2 {
				return fmt.Errorf("invalid max rule for field %s: missing value", fieldPath)
			}

			max, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid max value for field %s: %w", fieldPath, err)
			}

			if err := MaxRule(max)(field.Interface()); err != nil {
				return fmt.Errorf("validation failed for field %s: %w", fieldPath, err)
			}
			// Add more validation rules as needed
		}
	}

	return nil
}

// getFieldValue returns the value of a field at the given path
func getFieldValue(obj interface{}, path string) (reflect.Value, error) {
	value := reflect.ValueOf(obj)

	// If obj is a pointer, get the underlying value
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Ensure we're dealing with a struct
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("expected struct, got %v", value.Kind())
	}

	// Split the path into parts
	parts := strings.Split(path, ".")

	// Navigate through the struct fields
	for i, part := range parts {
		// Get the field by name
		field := value.FieldByName(part)
		if !field.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %s not found at part %d of path %s", part, i, path)
		}

		// If this is the last part of the path, return the field
		if i == len(parts)-1 {
			return field, nil
		}

		// If field is a pointer, get the underlying value
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return reflect.Value{}, fmt.Errorf("field %s is nil at part %d of path %s", part, i, path)
			}
			field = field.Elem()
		}

		// If the next level isn't a struct, we can't continue
		if field.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("field %s is not a struct at part %d of path %s", part, i, path)
		}

		// Continue with the nested struct
		value = field
	}

	// This should never happen if the function is used correctly
	return reflect.Value{}, fmt.Errorf("invalid field path: %s", path)
}

// Common validation rules

// RequiredRule validates that a field is not empty
func RequiredRule() func(interface{}) error {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.String:
			if v.String() == "" {
				return fmt.Errorf("value is required")
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() == 0 {
				return fmt.Errorf("value is required")
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v.Uint() == 0 {
				return fmt.Errorf("value is required")
			}
		case reflect.Float32, reflect.Float64:
			if v.Float() == 0 {
				return fmt.Errorf("value is required")
			}
		case reflect.Slice, reflect.Map, reflect.Array:
			if v.Len() == 0 {
				return fmt.Errorf("value is required")
			}
		case reflect.Ptr, reflect.Interface:
			if v.IsNil() {
				return fmt.Errorf("value is required")
			}
		}

		return nil
	}
}

// RangeRule validates that a numeric field is within a range
func RangeRule(min, max int64) func(interface{}) error {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)

		var val int64
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val = v.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uval := v.Uint()
			if uval > uint64(max) {
				return fmt.Errorf("value %d is greater than maximum %d", uval, max)
			}
			val = int64(uval)
		default:
			return fmt.Errorf("value must be numeric")
		}

		if val < min {
			return fmt.Errorf("value %d is less than minimum %d", val, min)
		}
		if val > max {
			return fmt.Errorf("value %d is greater than maximum %d", val, max)
		}

		return nil
	}
}

// MinRule validates that a numeric field is at least the minimum value
func MinRule(min int64) func(interface{}) error {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := v.Int()
			if val < min {
				return fmt.Errorf("value %d is less than minimum %d", val, min)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := v.Uint()
			if val < uint64(min) {
				return fmt.Errorf("value %d is less than minimum %d", val, min)
			}
		case reflect.Float32, reflect.Float64:
			val := v.Float()
			if val < float64(min) {
				return fmt.Errorf("value %f is less than minimum %d", val, min)
			}
		case reflect.Slice, reflect.Map, reflect.Array:
			length := v.Len()
			if int64(length) < min {
				return fmt.Errorf("length %d is less than minimum %d", length, min)
			}
		default:
			return fmt.Errorf("value type does not support min validation")
		}

		return nil
	}
}

// MaxRule validates that a numeric field is at most the maximum value
func MaxRule(max int64) func(interface{}) error {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := v.Int()
			if val > max {
				return fmt.Errorf("value %d is greater than maximum %d", val, max)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := v.Uint()
			if val > uint64(max) {
				return fmt.Errorf("value %d is greater than maximum %d", val, max)
			}
		case reflect.Float32, reflect.Float64:
			val := v.Float()
			if val > float64(max) {
				return fmt.Errorf("value %f is greater than maximum %d", val, max)
			}
		case reflect.Slice, reflect.Map, reflect.Array:
			length := v.Len()
			if int64(length) > max {
				return fmt.Errorf("length %d is greater than maximum %d", length, max)
			}
		default:
			return fmt.Errorf("value type does not support max validation")
		}

		return nil
	}
}
