package config

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// LoadConfig loads environment variables into the config struct based on tags
func LoadConfig(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	
	// Config must be a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}
	
	// Get the actual struct value
	val = val.Elem()
	typ := val.Type()
	
	// Iterate through all fields in the struct
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Skip if field is not exportable
		if !field.CanSet() {
			continue
		}
		
		// Get env tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}
		
		// Parse the tag - expected format: "ENV_NAME required:true default:value"
		parts := strings.Split(envTag, " ")
		envName := parts[0]
		
		// Initialize options
		var defaultValue string
		required := false
		
		// Process additional options in the tag
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "required:") {
				required = strings.TrimPrefix(part, "required:") == "true"
			} else if strings.HasPrefix(part, "default:") {
				defaultValue = strings.TrimPrefix(part, "default:")
			}
		}
		
		// Get environment variable value
		envValue := os.Getenv(envName)
		
		// If env value is empty, use default
		if envValue == "" {
			if required && defaultValue == "" {
				return fmt.Errorf("required environment variable %s not set", envName)
			}
			envValue = defaultValue
		}
		
		// Validate and set the field based on its type
		switch field.Kind() {
		case reflect.String:
			// Strings don't need validation
			field.SetString(envValue)
		
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// Validate that the value is a valid integer
			if envValue != "" {
				// Check if the string matches integer format
				matched, err := regexp.MatchString(`^-?\d+$`, envValue)
				if err != nil || !matched {
					return fmt.Errorf("environment variable %s must be an integer, got '%s'", envName, envValue)
				}
				
				// Parse the integer
				intValue, err := strconv.ParseInt(envValue, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid integer value for %s: %s", envName, err)
				}
				
				// Validate integer range for specific types
				switch field.Kind() {
				case reflect.Int8:
					if intValue < math.MinInt8 || intValue > math.MaxInt8 {
						return fmt.Errorf("value %d for %s is out of range for int8", intValue, envName)
					}
				case reflect.Int16:
					if intValue < math.MinInt16 || intValue > math.MaxInt16 {
						return fmt.Errorf("value %d for %s is out of range for int16", intValue, envName)
					}
				case reflect.Int32:
					if intValue < math.MinInt32 || intValue > math.MaxInt32 {
						return fmt.Errorf("value %d for %s is out of range for int32", intValue, envName)
					}
				}
				
				field.SetInt(intValue)
			}
			
		case reflect.Bool:
			// Validate boolean values
			validBools := map[string]bool{
				"true": true, "false": false,
				"yes": true, "no": false,
				"1": true, "0": false,
				"t": true, "f": false,
				"y": true, "n": false,
			}
			
			lowerValue := strings.ToLower(envValue)
			boolValue, valid := validBools[lowerValue]
			
			if !valid {
				return fmt.Errorf("environment variable %s must be a boolean value (true/false, yes/no, 1/0), got '%s'", envName, envValue)
			}
			
			field.SetBool(boolValue)
			
		case reflect.Float32, reflect.Float64:
			// Validate float values
			if envValue != "" {
				// Check if the string matches float format
				matched, err := regexp.MatchString(`^-?\d+(\.\d+)?$`, envValue)
				if err != nil || !matched {
					return fmt.Errorf("environment variable %s must be a float, got '%s'", envName, envValue)
				}
				
				// Parse the float
				floatValue, err := strconv.ParseFloat(envValue, 64)
				if err != nil {
					return fmt.Errorf("invalid float value for %s: %s", envName, err)
				}
				
				// Additional validation for float32
				if field.Kind() == reflect.Float32 {
					if floatValue < -math.MaxFloat32 || floatValue > math.MaxFloat32 {
						return fmt.Errorf("value %f for %s is out of range for float32", floatValue, envName)
					}
				}
				
				field.SetFloat(floatValue)
			}
			
		default:
			return fmt.Errorf("unsupported type for field %s (%s)", fieldType.Name, field.Kind())
		}
	}
	
	return nil
}
