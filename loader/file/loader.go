// Package file provides file loading utilities for scg-config.
package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hbttundar/scg-config/contract"
	"github.com/hbttundar/scg-config/errors"
	"github.com/hbttundar/scg-config/provider/viper"
	"github.com/hbttundar/scg-config/utils"
)

// Loader loads configuration files into the provider provider.
type Loader struct {
	provider contract.Provider
}

// NewFileLoader creates a new Loader for the given provider provider.
func NewFileLoader(p contract.Provider) *Loader {
	return &Loader{provider: p}
}

// LoadFromFile loads a single configuration file into the provider.
func (l *Loader) LoadFromFile(configFile string) error {
	provider := l.provider
	if provider == nil {
		return errors.ErrBackendProviderHasNoConfig
	}

	provider.SetConfigFile(configFile)

	if err := provider.ReadInConfig(); err != nil {
		return fmt.Errorf("%w: %w", errors.ErrReadConfigFileFailed, err)
	}

	return nil
}

// LoadFromDirectory loads all supported config files from a directory.
// Files are processed in alphabetical order, with the first file loaded normally
// and subsequent files merged to preserve nested block structures.
func (l *Loader) LoadFromDirectory(dir string) error {
	provider := l.provider
	if provider == nil {
		return errors.ErrBackendProviderHasNoConfig
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("%w: %w", errors.ErrFailedReadDirectory, err)
	}

	// Filter and collect supported config files
	var configFiles []string

	for _, file := range files {
		if file.IsDir() || !utils.IsSupportedConfigFile(file.Name()) {
			continue
		}

		configFiles = append(configFiles, file.Name())
	}

	if len(configFiles) == 0 {
		return nil // No config files found, not an error
	}

	isFirst := true

	for _, fileName := range configFiles {
		path := filepath.Join(dir, fileName)

		if isFirst {
			// Load the first file normally to establish the base configuration
			provider.SetConfigFile(path)

			if err := provider.ReadInConfig(); err != nil {
				return fmt.Errorf("failed to load initial config file %s: %w", path, err)
			}

			isFirst = false
		} else {
			// For subsequent files, use a more robust merging approach
			if err := l.mergeConfigFile(path); err != nil {
				return fmt.Errorf("failed to merge config file %s: %w", path, err)
			}
		}
	}

	return nil
}

// mergeConfigFile merges a configuration file into the existing provider configuration.
// This method uses the contract's MergeConfigMap method for better abstraction and
// handles complex nested block structures more reliably.
func (l *Loader) mergeConfigFile(configFile string) error {
	// Create a temporary provider to load the file we want to merge
	tempProvider := viper.NewConfigProvider()
	tempProvider.SetConfigFile(configFile)

	if err := tempProvider.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file for merging: %w", err)
	}

	// Get the configuration as a map and merge it using the contract method
	configMap := tempProvider.AllSettings()
	if err := l.provider.MergeConfigMap(configMap); err != nil {
		return fmt.Errorf("failed to merge configuration map: %w", err)
	}

	return nil
}

// GetProvider returns the Provider associated with the Loader.
//
//nolint:ireturn // returning an interface is required by the contract API
func (l *Loader) GetProvider() contract.Provider {
	return l.provider
}
