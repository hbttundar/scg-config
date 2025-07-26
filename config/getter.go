package config

import (
	"time"

	"github.com/hbttundar/scg-config/contract"
	"github.com/hbttundar/scg-config/dotmap"
	"github.com/hbttundar/scg-config/errors"
	"github.com/hbttundar/scg-config/utils"
)

type Getter struct {
	config map[string]any
}

func NewGetter(config map[string]any) *Getter {
	return &Getter{config: config}
}

// Get Core logic: flat key lookup first, dot-notation fallback.
func (g *Getter) Get(key string, typ contract.KeyType) (any, error) {
	if key == "" || g.config == nil {
		return nil, errors.ErrKeyNotFound
	}

	if val, ok := g.config[key]; ok {
		result, err := tryTypeCast(val, typ)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	val := dotmap.Resolve(g.config, key)
	if val == nil {
		return nil, errors.ErrKeyNotFound
	}

	val, err := tryTypeCast(val, typ)
	if err != nil {
		return nil, errors.ErrWrongType
	}

	return val, nil
}

func (g *Getter) GetKey(key string) any {
	val, _ := g.Get(key, contract.String)

	return val
}

func (g *Getter) GetString(key string) string {
	v, _ := g.Get(key, contract.String)
	if s, ok := v.(string); ok {
		return s
	}

	return ""
}

func (g *Getter) GetInt(key string) int {
	v, _ := g.Get(key, contract.Int)
	if i, ok := v.(int); ok {
		return i
	}

	return 0
}

func (g *Getter) GetBool(key string) bool {
	v, _ := g.Get(key, contract.Bool)
	if b, ok := v.(bool); ok {
		return b
	}

	return false
}

func (g *Getter) GetFloat64(key string) float64 {
	v, _ := g.Get(key, contract.Float64)
	if f, ok := v.(float64); ok {
		return f
	}

	return 0
}

func (g *Getter) GetDuration(key string) time.Duration {
	v, _ := g.Get(key, contract.Duration)
	if d, ok := v.(time.Duration); ok {
		return d
	}

	return 0
}

func (g *Getter) GetInt64(key string) int64 {
	v, _ := g.Get(key, contract.Int64)
	if i, ok := v.(int64); ok {
		return i
	}

	return 0
}

func (g *Getter) GetStringSlice(key string) []string {
	v, _ := g.Get(key, contract.StringSlice)
	if s, ok := v.([]string); ok {
		return s
	}

	return nil
}

func (g *Getter) GetStringMap(key string) map[string]interface{} {
	v, _ := g.Get(key, contract.Map)
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	return nil
}

func (g *Getter) GetStringMapString(key string) map[string]string {
	v, _ := g.Get(key, contract.Map)
	if m, ok := v.(map[string]string); ok {
		return m
	}

	return nil
}

func (g *Getter) GetTime(key string) time.Time {
	v, _ := g.Get(key, contract.Time)
	if t, ok := v.(time.Time); ok {
		return t
	}

	return time.Time{}
}

// HasKey Improved HasKey: flat first, then dot notation.
func (g *Getter) HasKey(key string) bool {
	if key == "" || g.config == nil {
		return false
	}

	if _, ok := g.config[key]; ok {
		return true
	}

	return dotmap.Resolve(g.config, key) != nil
}

// TypeConverter defines a function that converts a value to a specific type.
type TypeConverter func(val any) (any, error)

// typeConverters maps a KeyType to its converter function and associated error type.
//
//nolint:gochecknoglobals // a global converter map simplifies lookups and avoids an overly long helper function
var typeConverters = map[contract.KeyType]struct {
	converter TypeConverter
	errorType error
}{
	contract.Int: {
		converter: func(val any) (any, error) {
			return utils.ToInt(val)
		},
		errorType: errors.ErrNotInt,
	},
	contract.Int32: {
		converter: func(val any) (any, error) {
			return utils.ToInt32(val)
		},
		errorType: errors.ErrNotInt32,
	},
	contract.Int64: {
		converter: func(val any) (any, error) {
			return utils.ToInt64(val)
		},
		errorType: errors.ErrNotInt64,
	},
	contract.Uint: {
		converter: func(val any) (any, error) {
			return utils.ToUint(val)
		},
		errorType: errors.ErrNotUint,
	},
	contract.Uint32: {
		converter: func(val any) (any, error) {
			return utils.ToUint32(val)
		},
		errorType: errors.ErrNotUint32,
	},
	contract.Uint64: {
		converter: func(val any) (any, error) {
			return utils.ToUint64(val)
		},
		errorType: errors.ErrNotUint64,
	},
	contract.Float32: {
		converter: func(val any) (any, error) {
			return utils.ToFloat32(val)
		},
		errorType: errors.ErrNotFloat32,
	},
	contract.Float64: {
		converter: func(val any) (any, error) {
			return utils.ToFloat64(val)
		},
		errorType: errors.ErrNotFloat64,
	},
	contract.String: {
		converter: func(val any) (any, error) {
			return utils.ToString(val)
		},
		errorType: errors.ErrNotString,
	},
	contract.Bool: {
		converter: func(val any) (any, error) {
			return utils.ToBool(val)
		},
		errorType: errors.ErrNotBool,
	},
	contract.StringSlice: {
		converter: func(val any) (any, error) {
			return utils.ToStringSlice(val)
		},
		errorType: errors.ErrNotStringSlice,
	},
	contract.Map: {
		converter: func(val any) (any, error) {
			return utils.ToMap(val)
		},
		errorType: errors.ErrNotMap,
	},
	contract.Time: {
		converter: func(val any) (any, error) {
			return utils.ToTime(val)
		},
		errorType: errors.ErrNotTime,
	},
	contract.Duration: {
		converter: func(val any) (any, error) {
			return utils.ToDuration(val)
		},
		errorType: errors.ErrNotDuration,
	},
	contract.Bytes: {
		converter: func(val any) (any, error) {
			return utils.ToBytes(val)
		},
		errorType: errors.ErrNotBytes,
	},
	contract.UUID: {
		converter: func(val any) (any, error) {
			return utils.ToUUID(val)
		},
		errorType: errors.ErrNotUUID,
	},
	contract.URL: {
		converter: func(val any) (any, error) {
			return utils.ToURL(val)
		},
		errorType: errors.ErrNotURL,
	},
}

// tryTypeCast converts a value to the specified type using a function map approach.
// This reduces cognitive complexity by eliminating the large switch statement.
func tryTypeCast(val any, typ contract.KeyType) (any, error) {
	converterInfo, exists := typeConverters[typ]
	if !exists {
		return nil, errors.ErrUnknownType
	}

	value, err := converterInfo.converter(val)
	if err != nil {
		return nil, converterInfo.errorType
	}

	return value, nil
}
