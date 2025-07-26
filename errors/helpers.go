package errors

import errors2 "errors"

// Error variables for consistent usage.
var (
	ErrNotInt           = errors2.New("not an int")
	ErrNotInt32         = errors2.New("not an int32")
	ErrNotInt64         = errors2.New("not an int64")
	ErrNotUint          = errors2.New("not a uint")
	ErrNotUint32        = errors2.New("not a uint32")
	ErrNotUint64        = errors2.New("not a uint64")
	ErrNotFloat32       = errors2.New("not a float32")
	ErrNotFloat64       = errors2.New("not a float64")
	ErrNotString        = errors2.New("not a string")
	ErrNotBool          = errors2.New("not a bool")
	ErrNotStringInSlice = errors2.New("not a string in slice")
	ErrNotStringSlice   = errors2.New("not a string slice")
	ErrNotMap           = errors2.New("not a map")
	ErrNotTime          = errors2.New("not a time.Time")
	ErrNotDuration      = errors2.New("not a duration")
	ErrNotBytes         = errors2.New("not bytes")
	ErrNotUUID          = errors2.New("not a uuid")
	ErrNotURL           = errors2.New("not a URL")
)
