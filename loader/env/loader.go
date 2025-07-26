// Package env provides environment variable loading and normalization utilities for scg-config.
package env

import (
	"os"

	"github.com/hbttundar/scg-config/contract"
	loaderErrors "github.com/hbttundar/scg-config/errors"
	"github.com/hbttundar/scg-config/utils"
)

// Loader loads configuration from environment variables into the provider provider.
type Loader struct {
	provider contract.Provider
}

// NewEnvLoader creates a new Loader for the given provider provider.
func NewEnvLoader(p contract.Provider) *Loader {
	return &Loader{provider: p}
}

// LoadFromEnv loads environment variables with the given prefix into the provider.
// Prefix is stripped and keys are normalized to dot notation (e.g. APP_NAME -> app.name).
func (l *Loader) LoadFromEnv(prefix string) error {
	provider := l.provider
	if provider == nil {
		return loaderErrors.ErrBackendProviderNotSet
	}

	prefix = utils.NormalizePrefix(prefix)

	for _, envStr := range os.Environ() {
		if !utils.ShouldProcessEnv(envStr, prefix) {
			continue
		}

		key, value := utils.SplitEnv(envStr)
		key = utils.StripPrefix(key, prefix)
		key = utils.NormalizeEnvKey(key)

		provider.Set(key, value)
	}

	return nil
}

// GetProvider returns the Provider associated with the Loader.
//
//nolint:ireturn // returning an interface is required by the contract API
func (l *Loader) GetProvider() contract.Provider {
	return l.provider
}
