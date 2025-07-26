package utils

import (
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hbttundar/scg-config/contract"
	"github.com/hbttundar/scg-config/errors"
)

const (
	splitEnvParts = 2
)

// NormalizeEnvKey converts an environment variable key (e.g. APP_NAME) to dot notation (e.g. app.name).
func NormalizeEnvKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "_", "."))
}

// NormalizePrefix prepares the prefix for env matching.
func NormalizePrefix(prefix string) string {
	prefix = strings.ToUpper(prefix)
	if prefix != "" {
		return prefix + "_"
	}

	return ""
}

// ShouldProcessEnv checks if env matches the prefix (or no prefix).
func ShouldProcessEnv(env, prefix string) bool {
	return prefix == "" || strings.HasPrefix(env, prefix)
}

// SplitEnv splits "KEY=VALUE" into key, value.
func SplitEnv(env string) (string, string) {
	kv := strings.SplitN(env, "=", splitEnvParts)
	if len(kv) != splitEnvParts {
		return env, ""
	}

	return kv[0], kv[1]
}

// StripPrefix strips prefix from key, if present.
func StripPrefix(key, prefix string) string {
	if prefix != "" && strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}

	return key
}

// IsSupportedConfigFile returns true if the file has a supported config extension.
func IsSupportedConfigFile(filename string) bool {
	switch filepath.Ext(filename) {
	case contract.ExtYAML, contract.ExtYML, contract.ExtJSON:
		return true
	default:
		return false
	}
}

// --- Type conversion helpers with overflow checks and static errors ---

func ToInt(val any) (int, error) {
	switch value := val.(type) {
	case int:
		return value, nil
	case int64:
		if value > int64(math.MaxInt) || value < int64(math.MinInt) {
			return 0, fmt.Errorf("%w: int64 out of int range", errors.ErrNotInt)
		}

		converted := int(value)

		return converted, nil
	case float64:
		if value > float64(math.MaxInt) || value < float64(math.MinInt) {
			return 0, fmt.Errorf("%w: float64 out of int range", errors.ErrNotInt)
		}

		converted := int(value)

		return converted, nil
	case string:
		i, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotInt, err)
		}

		converted := i

		return converted, nil
	default:
		return 0, errors.ErrNotInt
	}
}

func ToInt32(val any) (int32, error) {
	switch value := val.(type) {
	case int32:
		return value, nil
	case int:
		if value > math.MaxInt32 || value < math.MinInt32 {
			return 0, fmt.Errorf("%w: int out of int32 range", errors.ErrNotInt32)
		}

		return int32(value), nil
	case string:
		// Use ParseInt with explicit bit size to avoid potential overflow converting from int
		int64Value, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotInt32, err)
		}
		// At this point the parsed value is guaranteed to fit into 32‑bits.
		return int32(int64Value), nil
	default:
		return 0, errors.ErrNotInt32
	}
}

func ToInt64(val any) (int64, error) {
	switch value := val.(type) {
	case int64:
		return value, nil
	case int:
		if value > math.MaxInt {
			return 0, fmt.Errorf("%w: int out of int64 range", errors.ErrNotInt64)
		}

		return int64(value), nil
	case float64:
		return int64(value), nil
	case string:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotInt64, err)
		}

		return intValue, nil
	default:
		return 0, errors.ErrNotInt64
	}
}

func ToUint(val any) (uint, error) {
	switch value := val.(type) {
	case uint:
		return value, nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("%w: int is negative", errors.ErrNotUint)
		}

		return uint(value), nil
	case string:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotUint, err)
		}

		if intVal < 0 {
			return 0, fmt.Errorf("%w: string is negative", errors.ErrNotUint)
		}

		return uint(intVal), nil
	default:
		return 0, errors.ErrNotUint
	}
}

func ToUint32(val any) (uint32, error) {
	switch value := val.(type) {
	case uint32:
		return value, nil
	case int:
		// int may be negative or exceed the maximum for uint32 on this platform
		if value < 0 || value > math.MaxUint32 {
			return 0, fmt.Errorf("%w: int out of uint32 range", errors.ErrNotUint32)
		}

		return uint32(value), nil
	case string:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotUint32, err)
		}

		return uint32(i), nil
	default:
		return 0, errors.ErrNotUint32
	}
}

func ToUint64(val any) (uint64, error) {
	switch value := val.(type) {
	case uint64:
		// A uint64 value is always within [0, math.MaxUint64], so no additional checks are needed.
		return value, nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("%w: int is negative", errors.ErrNotUint64)
		}

		return uint64(value), nil
	case string:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotUint64, err)
		}

		return i, nil
	default:
		return 0, errors.ErrNotUint64
	}
}

func ToFloat32(val any) (float32, error) {
	switch value := val.(type) {
	case float32:
		return value, nil
	case float64:
		if value > math.MaxFloat32 || value < -math.MaxFloat32 {
			return 0, fmt.Errorf("%w: float64 out of float32 range", errors.ErrNotFloat32)
		}

		return float32(value), nil
	case string:
		floatVal, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotFloat32, err)
		}

		if floatVal > math.MaxFloat32 || floatVal < -math.MaxFloat32 {
			return 0, fmt.Errorf("%w: parsed float out of float32 range", errors.ErrNotFloat32)
		}

		return float32(floatVal), nil
	default:
		return 0, errors.ErrNotFloat32
	}
}

func ToFloat64(val any) (float64, error) {
	switch value := val.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case string:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", errors.ErrNotFloat64, err)
		}

		return f, nil
	default:
		return 0, errors.ErrNotFloat64
	}
}

func ToString(val any) (string, error) {
	if s, ok := val.(string); ok {
		return s, nil
	}

	return "", errors.ErrNotString
}

func ToBool(val any) (bool, error) {
	switch value := val.(type) {
	case bool:
		return value, nil
	case string:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return false, fmt.Errorf("%w: %w", errors.ErrNotBool, err)
		}

		return b, nil
	default:
		return false, errors.ErrNotBool
	}
}

func ToStringSlice(val any) ([]string, error) {
	switch value := val.(type) {
	case []string:
		return value, nil
	case []any:
		result := make([]string, len(value))

		for idx, elem := range value {
			s, ok := elem.(string)
			if !ok {
				return nil, errors.ErrNotStringInSlice
			}

			result[idx] = s
		}

		return result, nil
	default:
		return nil, errors.ErrNotStringSlice
	}
}

func ToMap(val any) (map[string]any, error) {
	if m, ok := val.(map[string]any); ok {
		return m, nil
	}

	return nil, errors.ErrNotMap
}

func ToTime(val any) (time.Time, error) {
	if t, ok := val.(time.Time); ok {
		return t, nil
	}

	return time.Time{}, errors.ErrNotTime
}

func ToDuration(val any) (time.Duration, error) {
	if d, ok := val.(time.Duration); ok {
		return d, nil
	}

	return 0, errors.ErrNotDuration
}

func ToBytes(val any) ([]byte, error) {
	switch value := val.(type) {
	case []byte:
		return value, nil
	case string:
		return []byte(value), nil
	default:
		return nil, errors.ErrNotBytes
	}
}

func ToUUID(val any) (uuid.UUID, error) {
	switch value := val.(type) {
	case uuid.UUID:
		return value, nil
	case string:
		u, err := uuid.Parse(value)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%w: %w", errors.ErrNotUUID, err)
		}

		return u, nil
	default:
		return uuid.Nil, errors.ErrNotUUID
	}
}

func ToURL(val any) (*url.URL, error) {
	switch value := val.(type) {
	case *url.URL:
		return value, nil
	case string:
		u, err := url.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errors.ErrNotURL, err)
		}

		return u, nil
	default:
		return nil, errors.ErrNotURL
	}
}
