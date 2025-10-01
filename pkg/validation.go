package binigo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validator handles request validation
type Validator struct {
	data   map[string]interface{}
	rules  map[string][]string
	errors map[string][]string
}

// NewValidator creates a new validator instance
func NewValidator(data map[string]interface{}, rules map[string][]string) *Validator {
	return &Validator{
		data:   data,
		rules:  rules,
		errors: make(map[string][]string),
	}
}

// Validate runs validation and returns whether it passes
func (v *Validator) Validate() bool {
	for field, fieldRules := range v.rules {
		value := v.data[field]
		
		for _, rule := range fieldRules {
			if !v.validateRule(field, value, rule) {
				break // Stop on first error for this field
			}
		}
	}
	
	return len(v.errors) == 0
}

// validateRule validates a single rule
func (v *Validator) validateRule(field string, value interface{}, rule string) bool {
	parts := strings.SplitN(rule, ":", 2)
	ruleName := parts[0]
	var ruleValue string
	if len(parts) > 1 {
		ruleValue = parts[1]
	}
	
	switch ruleName {
	case "required":
		return v.validateRequired(field, value)
	case "email":
		return v.validateEmail(field, value)
	case "min":
		length, _ := strconv.Atoi(ruleValue)
		return v.validateMin(field, value, length)
	case "max":
		length, _ := strconv.Atoi(ruleValue)
		return v.validateMax(field, value, length)
	case "numeric":
		return v.validateNumeric(field, value)
	case "alpha":
		return v.validateAlpha(field, value)
	case "alpha_num":
		return v.validateAlphaNum(field, value)
	case "in":
		values := strings.Split(ruleValue, ",")
		return v.validateIn(field, value, values)
	case "confirmed":
		return v.validateConfirmed(field, value)
	case "url":
		return v.validateURL(field, value)
	case "regex":
		return v.validateRegex(field, value, ruleValue)
	case "unique":
		// Would need database connection to implement
		return true
	case "exists":
		// Would need database connection to implement
		return true
	default:
		return true
	}
}

// Individual validation methods

func (v *Validator) validateRequired(field string, value interface{}) bool {
	if value == nil {
		v.addError(field, fmt.Sprintf("The %s field is required", field))
		return false
	}
	
	str, ok := value.(string)
	if ok && strings.TrimSpace(str) == "" {
		v.addError(field, fmt.Sprintf("The %s field is required", field))
		return false
	}
	
	return true
}

func (v *Validator) validateEmail(field string, value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		v.addError(field, fmt.Sprintf("The %s must be a valid email address", field))
		return false
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		v.addError(field, fmt.Sprintf("The %s must be a valid email address", field))
		return false
	}
	
	return true
}

func (v *Validator) validateMin(field string, value interface{}, min int) bool {
	str, ok := value.(string)
	if !ok {
		return true
	}
	
	if len(str) < min {
		v.addError(field, fmt.Sprintf("The %s must be at least %d characters", field, min))
		return false
	}
	
	return true
}

func (v *Validator) validateMax(field string, value interface{}, max int) bool {
	str, ok := value.(string)
	if !ok {
		return true
	}
	
	if len(str) > max {
		v.addError(field, fmt.Sprintf("The %s must not exceed %d characters", field, max))
		return false
	}
	
	return true
}

func (v *Validator) validateNumeric(field string, value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		str := value.(string)
		if _, err := strconv.ParseFloat(str, 64); err != nil {
			v.addError(field, fmt.Sprintf("The %s must be numeric", field))
			return false
		}
		return true
	default:
		v.addError(field, fmt.Sprintf("The %s must be numeric", field))
		return false
	}
}

func (v *Validator) validateAlpha(field string, value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		v.addError(field, fmt.Sprintf("The %s must contain only letters", field))
		return false
	}
	
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		v.addError(field, fmt.Sprintf("The %s must contain only letters", field))
		return false
	}
	
	return true
}

func (v *Validator) validateAlphaNum(field string, value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		v.addError(field, fmt.Sprintf("The %s must contain only letters and numbers", field))
		return false
	}
	
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(str) {
		v.addError(field, fmt.Sprintf("The %s must contain only letters and numbers", field))
		return false
	}
	
	return true
}

func (v *Validator) validateIn(field string, value interface{}, allowedValues []string) bool {
	str, ok := value.(string)
	if !ok {
		v.addError(field, fmt.Sprintf("The %s is invalid", field))
		return false
	}
	
	for _, allowed := range allowedValues {
		if str == strings.TrimSpace(allowed) {
			return true
		}
	}
	
	v.addError(field, fmt.Sprintf("The %s must be one of: %s", field, strings.Join(allowedValues, ", ")))
	return false
}

func (v *Validator) validateConfirmed(field string, value interface{}) bool {
	confirmField := field + "_confirmation"
	confirmValue := v.data[confirmField]
	
	if value != confirmValue {
		v.addError(field, fmt.Sprintf("The %s confirmation does not match", field))
		return false
	}
	
	return true
}

func (v *Validator) validateURL(field string, value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		v.addError(field, fmt.Sprintf("The %s must be a valid URL", field))
		return false
	}
	
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(str) {
		v.addError(field, fmt.Sprintf("The %s must be a valid URL", field))
		return false
	}
	
	return true
}

func (v *Validator) validateRegex(field string, value interface{}, pattern string) bool {
	str, ok := value.(string)
	if !ok {
		v.addError(field, fmt.Sprintf("The %s format is invalid", field))
		return false
	}
	
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return true // Invalid regex pattern, skip
	}
	
	if !regex.MatchString(str) {
		v.addError(field, fmt.Sprintf("The %s format is invalid", field))
		return false
	}
	
	return true
}

// addError adds an error message for a field
func (v *Validator) addError(field, message string) {
	if v.errors[field] == nil {
		v.errors[field] = make([]string, 0)
	}
	v.errors[field] = append(v.errors[field], message)
}

// Errors returns validation errors
func (v *Validator) Errors() map[string][]string {
	return v.errors
}

// FirstError returns the first error message
func (v *Validator) FirstError() string {
	for _, messages := range v.errors {
		if len(messages) > 0 {
			return messages[0]
		}
	}
	return ""
}

// HasErrors checks if there are any validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Context validation helper
func (c *Context) ValidateJSON(rules map[string][]string) (*Validator, error) {
	var data map[string]interface{}
	if err := c.Bind(&data); err != nil {
		return nil, err
	}
	
	validator := NewValidator(data, rules)
	validator.Validate()
	
	return validator, nil
}

// ValidateRequest validates request and returns errors if any
func (c *Context) ValidateRequest(rules map[string][]string) error {
	validator, err := c.ValidateJSON(rules)
	if err != nil {
		return c.Error("Invalid JSON", 400)
	}
	
	if validator.HasErrors() {
		return c.Status(422).JSON(Map{
			"success": false,
			"message": "Validation failed",
			"errors":  validator.Errors(),
		})
	}
	
	return nil
}

// Example usage in controller:
/*
func (uc *UserController) Store(ctx *framework.Context) error {
	rules := map[string][]string{
		"name":     {"required", "min:3", "max:50"},
		"email":    {"required", "email"},
		"password": {"required", "min:8", "confirmed"},
		"age":      {"numeric", "min:18"},
		"role":     {"in:admin,user,guest"},
	}
	
	if err := ctx.ValidateRequest(rules); err != nil {
		return err
	}
	
	// Continue with valid data...
}
*/