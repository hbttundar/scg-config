package errors

import "errors"

var (
	ErrBackendProviderNotSet      = errors.New("no provider provider set for environment loader")
	ErrBackendProviderHasNoConfig = errors.New("provider provider has no config provider set")
	ErrReadConfigFileFailed       = errors.New("failed to read configuration file")
	ErrFailedReadDirectory        = errors.New("failed to read directory")
)
