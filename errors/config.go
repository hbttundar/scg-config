package errors

import "errors"

var (
	ErrKeyNotFound = errors.New("config: key not found")
	ErrWrongType   = errors.New("config: wrong type for key")
	ErrUnknownType = errors.New("config: unknown type for key")
)
