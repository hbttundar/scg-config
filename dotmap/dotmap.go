// Package dotmap provides utilities for traversing nested maps with dot notation.
package dotmap

import (
	"strconv"
	"strings"
)

// Resolve navigates a nested map using dot notation (e.g., "foo.bar.0.baz").
// Returns nil if any part of the path does not exist (case-insensitive fallback).
func Resolve(settings map[string]interface{}, path string) interface{} {
	if settings == nil || path == "" {
		return nil
	}

	parts := strings.Split(path, ".")

	// First try case-sensitive resolution
	result := resolvePath(settings, parts, false)
	if result != nil {
		return result
	}

	// If case-sensitive failed, try case-insensitive from the root
	return resolvePath(settings, parts, true)
}

// resolvePath traverses the path with the specified case sensitivity.
func resolvePath(current interface{}, parts []string, caseInsensitive bool) interface{} {
	for _, part := range parts {
		next, found := resolveStep(current, part, caseInsensitive)
		if !found {
			return nil
		}

		current = next
	}

	return current
}

// resolveStep attempts to resolve one step in the path.
func resolveStep(current interface{}, part string, caseInsensitive bool) (interface{}, bool) {
	// Handle array indexing first (same for both case modes)
	if idx, err := strconv.Atoi(part); err == nil {
		return resolveArrayIndex(current, idx)
	}

	// Handle map types
	switch curr := current.(type) {
	case map[string]interface{}:
		return resolveStringMap(curr, part, caseInsensitive)
	case map[string]string:
		return resolveStringStringMap(curr, part, caseInsensitive)
	case map[interface{}]interface{}:
		return resolveInterfaceMap(curr, part, caseInsensitive)
	}

	return nil, false
}

// resolveArrayIndex handles array/slice indexing for both []interface{} and []string.
func resolveArrayIndex(current interface{}, idx int) (interface{}, bool) {
	switch curr := current.(type) {
	case []interface{}:
		if idx >= 0 && idx < len(curr) {
			return curr[idx], true
		}
	case []string:
		if idx >= 0 && idx < len(curr) {
			return curr[idx], true
		}
	}

	return nil, false
}

// resolveStringMap handles map[string]interface{} with optional case-insensitive lookup.
func resolveStringMap(configData map[string]interface{}, key string, caseInsensitive bool) (interface{}, bool) {
	if !caseInsensitive {
		val, exists := configData[key]

		return val, exists
	}

	for k, v := range configData {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}

	return nil, false
}

// resolveStringStringMap handles map[string]string with optional case-insensitive lookup.
func resolveStringStringMap(configData map[string]string, key string, caseInsensitive bool) (interface{}, bool) {
	if !caseInsensitive {
		val, exists := configData[key]

		return val, exists
	}

	for k, v := range configData {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}

	return nil, false
}

// resolveInterfaceMap handles map[interface{}]interface{} with optional case-insensitive lookup.
func resolveInterfaceMap(configData map[interface{}]interface{}, key string, caseInsensitive bool) (interface{}, bool) {
	if !caseInsensitive {
		val, exists := configData[key]

		return val, exists
	}

	for k, v := range configData {
		if ks, ok := k.(string); ok && strings.EqualFold(ks, key) {
			return v, true
		}
	}

	return nil, false
}
