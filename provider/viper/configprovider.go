package viper

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/hbttundar/scg-config/contract"
)

// ConfigProvider implements contract.Provider using Viper.
type ConfigProvider struct {
	v *viper.Viper
}

// NewConfigProvider returns a new ConfigProvider instance (satisfies contract.Provider).
func NewConfigProvider() *ConfigProvider {
	return &ConfigProvider{v: viper.New()}
}

// AllSettings returns the entire config as a nested map.
func (b *ConfigProvider) AllSettings() map[string]interface{} {
	return b.v.AllSettings()
}

// GetKey returns the value for a key (flat lookup only, for bootstrapping and tests).
func (b *ConfigProvider) GetKey(key string) any {
	return b.v.Get(key)
}

// IsSet checks if a config key is present (flat lookup).
func (b *ConfigProvider) IsSet(key string) bool {
	return b.v.IsSet(key)
}

// Set sets a key in the Viper store (for tests or live editing).
func (b *ConfigProvider) Set(key string, value any) {
	b.v.Set(key, value)
}

// ReadInConfig reloads from file/env if supported by Viper.
func (b *ConfigProvider) ReadInConfig() error {
	if err := b.v.ReadInConfig(); err != nil {
		return fmt.Errorf("provider: failed to read config: %w", err)
	}

	return nil
}

// SetConfigFile sets which file to read.
func (b *ConfigProvider) SetConfigFile(file string) {
	b.v.SetConfigFile(file)
}

// MergeConfigMap merges another map into config.
func (b *ConfigProvider) MergeConfigMap(cfg map[string]interface{}) error {
	if err := b.v.MergeConfigMap(cfg); err != nil {
		return fmt.Errorf("provider: failed to merge config map: %w", err)
	}

	return nil
}

// Provider returns the underlying Viper object for advanced use.
func (b *ConfigProvider) Provider() any {
	return b.v
}

// Interface assertion: this struct implements contract.Provider.
var _ contract.Provider = (*ConfigProvider)(nil)
