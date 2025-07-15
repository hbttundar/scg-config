// Package scg_config provides a Laravel-like configuration system wrapping spf13/viper
package scg_config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config is a wrapper around viper that provides a Laravel-like configuration API
type Config struct {
	viper        *viper.Viper
	watchedFiles map[string]bool
	watchMutex   sync.RWMutex
	watcher      *fsnotify.Watcher
	validators   map[string]ValidatorFunc
}

// ValidatorFunc is a function that validates a configuration value
// It returns an error if the value is invalid, or nil if it's valid
type ValidatorFunc func(value interface{}) error

// New creates a new Config instance with default settings
func New() *Config {
	v := viper.New()
	return &Config{
		viper:        v,
		watchedFiles: make(map[string]bool),
		validators:   make(map[string]ValidatorFunc),
	}
}

// NewWithViper creates a new Config instance with a pre-configured viper instance
func NewWithViper(v *viper.Viper) *Config {
	return &Config{
		viper:        v,
		watchedFiles: make(map[string]bool),
		validators:   make(map[string]ValidatorFunc),
	}
}

// LoadFromFile loads configuration from a file
// Returns an error if the file cannot be read or parsed
func (c *Config) LoadFromFile(configFile string) error {
	c.viper.SetConfigFile(configFile)
	return c.viper.ReadInConfig()
}

// LoadFromEnv loads configuration from environment variables with a specific prefix
// Environment variables are automatically converted from dot notation to underscore notation
// For example, "app.debug" becomes "PREFIX_APP_DEBUG"
func (c *Config) LoadFromEnv(prefix string) {
	c.viper.SetEnvPrefix(prefix)
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// LoadFromDirectory loads all .yaml, .yml, and .json files from the given directory
// and merges them into the configuration. The filename (without extension) is used
// as the top-level namespace.
// Nested directories are supported and their names are used as part of the namespace.
// Returns an error if a file fails to load.
func (c *Config) LoadFromDirectory(dir string) error {
	return c.loadFromDirectoryWithPrefix(dir, "")
}

// LoadFromDirectoryWithOptions loads all .yaml, .yml, and .json files from the given directory
// with additional options for error handling.
// Parameters:
//   - dir: the directory to load configuration files from
//   - continueOnError: if true, will continue loading files even if some fail
//
// Returns a slice of errors encountered during loading, or nil if no errors occurred.
// This method allows for more flexible error handling compared to LoadFromDirectory.
func (c *Config) LoadFromDirectoryWithOptions(dir string, continueOnError bool) []error {
	return c.loadFromDirectoryWithPrefixAndOptions(dir, "", continueOnError)
}

// loadFromDirectoryWithPrefix is a helper method that loads configuration files from a directory
// with an optional namespace prefix for nested directories.
func (c *Config) loadFromDirectoryWithPrefix(dir string, prefix string) error {
	errors := c.loadFromDirectoryWithPrefixAndOptions(dir, prefix, false)
	if len(errors) > 0 {
		return errors[0] // Return the first error encountered
	}
	return nil
}

// loadFromDirectoryWithPrefixAndOptions is a helper method that loads configuration files from a directory
// with an optional namespace prefix for nested directories and additional error handling options.
// continueOnError: if true, will continue loading files even if some fail
// Returns a slice of errors encountered during loading, or nil if no errors occurred
func (c *Config) loadFromDirectoryWithPrefixAndOptions(dir string, prefix string, continueOnError bool) []error {
	var errors []error

	// Read all files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []error{fmt.Errorf("failed to read directory: %w", err)}
	}

	// Process each file and subdirectory
	for _, entry := range entries {
		// Handle subdirectories recursively
		if entry.IsDir() {
			subDirName := entry.Name()
			var subDirPrefix string
			if prefix == "" {
				subDirPrefix = subDirName
			} else {
				subDirPrefix = prefix + "." + subDirName
			}

			subErrors := c.loadFromDirectoryWithPrefixAndOptions(filepath.Join(dir, subDirName), subDirPrefix, continueOnError)
			if len(subErrors) > 0 {
				if continueOnError {
					errors = append(errors, subErrors...)
				} else {
					return subErrors
				}
			}
			continue
		}

		// Get the file name and extension
		filename := entry.Name()
		ext := filepath.Ext(filename)

		// Check if the file has a supported extension
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		// Get the file name without extension to use as namespace
		namespace := filename[:len(filename)-len(ext)]

		// Apply prefix if it exists
		if prefix != "" {
			namespace = prefix + "." + namespace
		}

		// Create a new viper instance for this file
		fileViper := viper.New()
		fileViper.SetConfigFile(filepath.Join(dir, filename))

		// Read the file
		if err := fileViper.ReadInConfig(); err != nil {
			fileErr := fmt.Errorf("failed to read config file %s: %w", filename, err)
			if continueOnError {
				errors = append(errors, fileErr)
				continue
			}
			return []error{fileErr}
		}

		// Get all settings from the file
		settings := fileViper.AllSettings()

		// Create a new map with the namespace as the top-level key
		namespacedSettings := map[string]interface{}{
			namespace: settings,
		}

		// Merge the settings into the main viper instance
		if err := c.viper.MergeConfigMap(namespacedSettings); err != nil {
			mergeErr := fmt.Errorf("failed to merge config from file %s: %w", filename, err)
			if continueOnError {
				errors = append(errors, mergeErr)
				continue
			}
			return []error{mergeErr}
		}
	}

	return errors
}

// traverseNestedMap retrieves a value from a nested map using dot notation
// It splits the path by dots and traverses the nested map structure
// Enhanced to handle arrays/slices with numeric indices and various map types
func (c *Config) traverseNestedMap(settings map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = settings

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			// Standard map[string]interface{} case
			current = v[part]

		case map[string]string:
			// Handle map[string]string case
			if val, ok := v[part]; ok {
				current = val
			} else {
				return nil
			}

		case []interface{}:
			// Handle array/slice with numeric index
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return nil
			}
			current = v[index]

		case []string:
			// Handle string slice case
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return nil
			}
			current = v[index]

		case map[interface{}]interface{}:
			// Handle map[interface{}]interface{} case
			current = v[part]

		default:
			// If we reach here with any other type, we can't navigate further
			return nil
		}

		// If current became nil during navigation, we can't go further
		if current == nil {
			return nil
		}
	}

	return current
}

// Get retrieves a configuration value using dot notation
// If viper's native Get fails to retrieve a nested value, it falls back to manual traversal
// Returns nil if the key does not exist
func (c *Config) Get(key string) interface{} {
	// First try viper's native Get
	value := c.viper.Get(key)

	// If value is nil and the key contains dots, try our custom traversal
	if value == nil && strings.Contains(key, ".") {
		// Split the key by dots to get the top-level key
		parts := strings.SplitN(key, ".", 2)
		topLevelKey := parts[0]

		// Get the top-level value from viper
		topLevelValue := c.viper.Get(topLevelKey)

		if topLevelValue != nil && len(parts) > 1 {
			// If we have a top-level value and there's a path to traverse
			if m, ok := topLevelValue.(map[string]interface{}); ok {
				// Use our enhanced traversal for the rest of the path
				value = c.traverseNestedMap(m, parts[1])
			}
		}

		// If still nil, try with lowercase keys (Viper converts keys to lowercase)
		if value == nil {
			// Try with lowercase top-level key
			lowercaseTopLevelKey := strings.ToLower(topLevelKey)
			if lowercaseTopLevelKey != topLevelKey {
				topLevelValue = c.viper.Get(lowercaseTopLevelKey)
				if topLevelValue != nil && len(parts) > 1 {
					if m, ok := topLevelValue.(map[string]interface{}); ok {
						value = c.traverseNestedMapCaseInsensitive(m, parts[1])
					}
				}
			} else {
				// Try with case-insensitive traversal
				topLevelValue = c.viper.Get(topLevelKey)
				if topLevelValue != nil && len(parts) > 1 {
					if m, ok := topLevelValue.(map[string]interface{}); ok {
						value = c.traverseNestedMapCaseInsensitive(m, parts[1])
					}
				}
			}
		}
	}

	return value
}

// traverseNestedMapCaseInsensitive is similar to traverseNestedMap but ignores case in keys
// This is useful because Viper converts keys to lowercase
func (c *Config) traverseNestedMapCaseInsensitive(settings map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = settings

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			// Try exact match first
			if val, ok := v[part]; ok {
				current = val
			} else {
				// Try case-insensitive match
				lowerPart := strings.ToLower(part)
				found := false
				for k, val := range v {
					if strings.ToLower(k) == lowerPart {
						current = val
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}

		case map[string]string:
			// Try exact match first
			if val, ok := v[part]; ok {
				current = val
			} else {
				// Try case-insensitive match
				lowerPart := strings.ToLower(part)
				found := false
				for k, val := range v {
					if strings.ToLower(k) == lowerPart {
						current = val
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}

		case []interface{}:
			// Handle array/slice with numeric index
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return nil
			}
			current = v[index]

		case []string:
			// Handle string slice case
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return nil
			}
			current = v[index]

		case map[interface{}]interface{}:
			// Try exact match first
			if val, ok := v[part]; ok {
				current = val
			} else {
				// Try case-insensitive match for string keys
				lowerPart := strings.ToLower(part)
				found := false
				for k, val := range v {
					if kStr, ok := k.(string); ok && strings.ToLower(kStr) == lowerPart {
						current = val
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}

		default:
			// If we reach here with any other type, we can't navigate further
			return nil
		}

		// If current became nil during navigation, we can't go further
		if current == nil {
			return nil
		}
	}

	return current
}

// getTypedValue is a helper method to handle type conversion with fallback to viper
// It attempts to convert the value to the desired type, and if that fails,
// it falls back to the corresponding viper method
func (c *Config) getTypedValue(key string, defaultValue interface{},
	typeCheck func(interface{}) (interface{}, bool),
	viperGet func(string) interface{}) interface{} {

	value := c.Get(key)
	if value == nil {
		return defaultValue
	}

	if converted, ok := typeCheck(value); ok {
		return converted
	}

	return viperGet(key)
}

// GetString retrieves a string configuration value using dot notation
// Returns an empty string if the key does not exist
func (c *Config) GetString(key string) string {
	result := c.getTypedValue(
		key,
		"",
		func(v interface{}) (interface{}, bool) {
			str, ok := v.(string)
			return str, ok
		},
		func(k string) interface{} {
			return c.viper.GetString(k)
		},
	)
	return result.(string)
}

// GetBool retrieves a boolean configuration value using dot notation
// Returns false if the key does not exist
func (c *Config) GetBool(key string) bool {
	result := c.getTypedValue(
		key,
		false,
		func(v interface{}) (interface{}, bool) {
			b, ok := v.(bool)
			return b, ok
		},
		func(k string) interface{} {
			return c.viper.GetBool(k)
		},
	)
	return result.(bool)
}

// GetInt retrieves an integer configuration value using dot notation
// Returns 0 if the key does not exist
func (c *Config) GetInt(key string) int {
	result := c.getTypedValue(
		key,
		0,
		func(v interface{}) (interface{}, bool) {
			i, ok := v.(int)
			return i, ok
		},
		func(k string) interface{} {
			return c.viper.GetInt(k)
		},
	)
	return result.(int)
}

// GetInt64 retrieves a 64-bit integer configuration value using dot notation
// Returns 0 if the key does not exist
func (c *Config) GetInt64(key string) int64 {
	result := c.getTypedValue(
		key,
		int64(0),
		func(v interface{}) (interface{}, bool) {
			i, ok := v.(int64)
			return i, ok
		},
		func(k string) interface{} {
			return c.viper.GetInt64(k)
		},
	)
	return result.(int64)
}

// GetFloat64 retrieves a 64-bit float configuration value using dot notation
// Returns 0.0 if the key does not exist
func (c *Config) GetFloat64(key string) float64 {
	result := c.getTypedValue(
		key,
		float64(0),
		func(v interface{}) (interface{}, bool) {
			f, ok := v.(float64)
			return f, ok
		},
		func(k string) interface{} {
			return c.viper.GetFloat64(k)
		},
	)
	return result.(float64)
}

// GetTime retrieves a time.Time configuration value using dot notation
// Returns zero time if the key does not exist
func (c *Config) GetTime(key string) time.Time {
	result := c.getTypedValue(
		key,
		time.Time{},
		func(v interface{}) (interface{}, bool) {
			t, ok := v.(time.Time)
			return t, ok
		},
		func(k string) interface{} {
			return c.viper.GetTime(k)
		},
	)
	return result.(time.Time)
}

// GetDuration retrieves a time.Duration configuration value using dot notation
// Returns 0 duration if the key does not exist
func (c *Config) GetDuration(key string) time.Duration {
	result := c.getTypedValue(
		key,
		time.Duration(0),
		func(v interface{}) (interface{}, bool) {
			d, ok := v.(time.Duration)
			return d, ok
		},
		func(k string) interface{} {
			return c.viper.GetDuration(k)
		},
	)
	return result.(time.Duration)
}

// GetStringSlice retrieves a string slice configuration value using dot notation
// Returns nil if the key does not exist
func (c *Config) GetStringSlice(key string) []string {
	result := c.getTypedValue(
		key,
		[]string(nil),
		func(v interface{}) (interface{}, bool) {
			slice, ok := v.([]string)
			return slice, ok
		},
		func(k string) interface{} {
			return c.viper.GetStringSlice(k)
		},
	)
	return result.([]string)
}

// GetStringMap retrieves a string map configuration value using dot notation
// Returns nil if the key does not exist
func (c *Config) GetStringMap(key string) map[string]interface{} {
	result := c.getTypedValue(
		key,
		map[string]interface{}(nil),
		func(v interface{}) (interface{}, bool) {
			m, ok := v.(map[string]interface{})
			return m, ok
		},
		func(k string) interface{} {
			return c.viper.GetStringMap(k)
		},
	)
	return result.(map[string]interface{})
}

// GetStringMapString retrieves a string map of strings configuration value using dot notation
// Returns nil if the key does not exist
func (c *Config) GetStringMapString(key string) map[string]string {
	result := c.getTypedValue(
		key,
		map[string]string(nil),
		func(v interface{}) (interface{}, bool) {
			m, ok := v.(map[string]string)
			return m, ok
		},
		func(k string) interface{} {
			return c.viper.GetStringMapString(k)
		},
	)
	return result.(map[string]string)
}

// Set sets a configuration value
// The key can use dot notation to set nested values
func (c *Config) Set(key string, value interface{}) {
	c.viper.Set(key, value)
}

// IsSet checks if a configuration key exists
// The key can use dot notation to check nested values
func (c *Config) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

// Has is an alias for IsSet for backward compatibility
// Deprecated: Use IsSet instead
func (c *Config) Has(key string) bool {
	return c.IsSet(key)
}

// mustGet is a helper method that panics if the key doesn't exist
func (c *Config) mustGet(key string) {
	if !c.IsSet(key) {
		panic("Configuration key not found: " + key)
	}
}

// MustGetString retrieves a string configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetString(key string) string {
	c.mustGet(key)
	return c.GetString(key)
}

// MustGetBool retrieves a boolean configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetBool(key string) bool {
	c.mustGet(key)
	return c.GetBool(key)
}

// MustGetInt retrieves an integer configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetInt(key string) int {
	c.mustGet(key)
	return c.GetInt(key)
}

// MustGetInt64 retrieves a 64-bit integer configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetInt64(key string) int64 {
	c.mustGet(key)
	return c.GetInt64(key)
}

// MustGetFloat64 retrieves a 64-bit float configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetFloat64(key string) float64 {
	c.mustGet(key)
	return c.GetFloat64(key)
}

// MustGetTime retrieves a time.Time configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetTime(key string) time.Time {
	c.mustGet(key)
	return c.GetTime(key)
}

// MustGetDuration retrieves a time.Duration configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetDuration(key string) time.Duration {
	c.mustGet(key)
	return c.GetDuration(key)
}

// MustGetStringSlice retrieves a string slice configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetStringSlice(key string) []string {
	c.mustGet(key)
	return c.GetStringSlice(key)
}

// MustGetStringMap retrieves a string map configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetStringMap(key string) map[string]interface{} {
	c.mustGet(key)
	return c.GetStringMap(key)
}

// MustGetStringMapString retrieves a string map of strings configuration value using dot notation
// Panics if the key does not exist
func (c *Config) MustGetStringMapString(key string) map[string]string {
	c.mustGet(key)
	return c.GetStringMapString(key)
}

// GetViper returns the underlying viper instance
// This can be used to access viper functionality not exposed by this wrapper
func (c *Config) GetViper() *viper.Viper {
	return c.viper
}

// WatchConfig starts watching a configuration file for changes
// When the file changes, the configuration is automatically reloaded
// Returns an error if the file cannot be watched
func (c *Config) WatchConfig(configFile string) error {
	return c.WatchConfigWithCallback(configFile, nil)
}

// WatchConfigWithCallback starts watching a configuration file for changes
// When the file changes, the configuration is automatically reloaded and the callback is called
// If callback is nil, only the configuration is reloaded
// Returns an error if the file cannot be watched
func (c *Config) WatchConfigWithCallback(configFile string, callback func()) error {
	c.watchMutex.Lock()
	defer c.watchMutex.Unlock()

	// Initialize the watcher if it doesn't exist
	if c.watcher == nil {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create file watcher: %w", err)
		}
		c.watcher = watcher

		// Start a goroutine to handle events
		go func() {
			for {
				select {
				case event, ok := <-c.watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						// Reload the configuration
						if err := c.ReloadConfig(event.Name); err != nil {
							fmt.Printf("Error reloading config: %v\n", err)
						}
						// Call the callback if provided
						if callback != nil {
							callback()
						}
					}
				case err, ok := <-c.watcher.Errors:
					if !ok {
						return
					}
					fmt.Printf("Error watching config file: %v\n", err)
				}
			}
		}()
	}

	// Check if the file is already being watched
	if c.watchedFiles[configFile] {
		return nil
	}

	// Add the file to the watcher
	if err := c.watcher.Add(configFile); err != nil {
		return fmt.Errorf("failed to watch config file %s: %w", configFile, err)
	}

	// Mark the file as watched
	c.watchedFiles[configFile] = true

	return nil
}

// StopWatching stops watching all configuration files
// Returns an error if the watcher cannot be closed
func (c *Config) StopWatching() error {
	c.watchMutex.Lock()
	defer c.watchMutex.Unlock()

	if c.watcher == nil {
		return nil
	}

	// Close the watcher
	if err := c.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close file watcher: %w", err)
	}

	// Clear the watcher and watched files
	c.watcher = nil
	c.watchedFiles = make(map[string]bool)

	return nil
}

// StopWatchingFile stops watching a specific configuration file
// Returns an error if the file cannot be removed from the watcher
func (c *Config) StopWatchingFile(configFile string) error {
	c.watchMutex.Lock()
	defer c.watchMutex.Unlock()

	if c.watcher == nil || !c.watchedFiles[configFile] {
		return nil
	}

	// Remove the file from the watcher
	if err := c.watcher.Remove(configFile); err != nil {
		return fmt.Errorf("failed to stop watching config file %s: %w", configFile, err)
	}

	// Mark the file as not watched
	delete(c.watchedFiles, configFile)

	return nil
}

// ReloadConfig reloads configuration from a file
// Returns an error if the file cannot be read or parsed
func (c *Config) ReloadConfig(configFile string) error {
	// Create a new viper instance
	v := viper.New()
	v.SetConfigFile(configFile)

	// Read the configuration
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to reload config file %s: %w", configFile, err)
	}

	// Merge the new configuration into the existing one
	if err := c.viper.MergeConfigMap(v.AllSettings()); err != nil {
		return fmt.Errorf("failed to merge reloaded config: %w", err)
	}

	return nil
}

// SaveConfig saves the current configuration to a file
// The file format is determined by the file extension
// Supported formats: JSON, YAML, TOML, INI, HCL
// Returns an error if the file cannot be written
func (c *Config) SaveConfig(configFile string) error {
	// Set the config file in viper
	c.viper.SetConfigFile(configFile)

	// Get the file extension to determine the format
	ext := filepath.Ext(configFile)
	if ext == "" {
		return fmt.Errorf("config file must have an extension to determine format")
	}

	// Remove the dot from the extension
	format := ext[1:]

	// Set the config type
	c.viper.SetConfigType(format)

	// Write the config file
	if err := c.viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to save config to file %s: %w", configFile, err)
	}

	return nil
}

// SaveConfigAs saves the current configuration to a file with a specific format
// format should be one of: "json", "yaml", "toml", "ini", "hcl"
// Returns an error if the file cannot be written or the format is not supported
func (c *Config) SaveConfigAs(configFile string, format string) error {
	// Validate the format
	format = strings.ToLower(format)
	switch format {
	case "json", "yaml", "yml", "toml", "ini", "hcl":
		// Valid format
	default:
		return fmt.Errorf("unsupported config format: %s", format)
	}

	// Set the config file and type in viper
	c.viper.SetConfigFile(configFile)
	c.viper.SetConfigType(format)

	// Write the config file
	if err := c.viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to save config to file %s with format %s: %w", configFile, format, err)
	}

	return nil
}

// AddValidator adds a validator function for a configuration key
// The validator will be called when ValidateKey or ValidateAll is called
// If a validator already exists for the key, it will be replaced
func (c *Config) AddValidator(key string, validator ValidatorFunc) {
	c.validators[key] = validator
}

// RemoveValidator removes a validator function for a configuration key
// If no validator exists for the key, this is a no-op
func (c *Config) RemoveValidator(key string) {
	delete(c.validators, key)
}

// ValidateKey validates a specific configuration key
// Returns an error if the key has a validator and the validation fails
// Returns nil if the key has no validator or the validation passes
func (c *Config) ValidateKey(key string) error {
	validator, exists := c.validators[key]
	if !exists {
		return nil
	}

	value := c.Get(key)
	return validator(value)
}

// ValidateAll validates all configuration keys that have validators
// Returns a map of keys to validation errors for keys that failed validation
// Returns nil if all validations pass
func (c *Config) ValidateAll() map[string]error {
	errors := make(map[string]error)

	for key, validator := range c.validators {
		value := c.Get(key)
		if err := validator(value); err != nil {
			errors[key] = err
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

// ValidateAllStrict validates all configuration keys that have validators
// Returns an error if any validation fails, with details about all failed validations
// Returns nil if all validations pass
func (c *Config) ValidateAllStrict() error {
	errors := c.ValidateAll()
	if errors == nil {
		return nil
	}

	var errMsgs []string
	for key, err := range errors {
		errMsgs = append(errMsgs, fmt.Sprintf("%s: %v", key, err))
	}

	return fmt.Errorf("validation failed for %d keys: %s", len(errors), strings.Join(errMsgs, "; "))
}
