package config_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbttundar/scg-config/config"
	"github.com/hbttundar/scg-config/contract"
)

func baseConfigMap() map[string]any {
	const uuidStr = "1336301d-4e85-4b76-a2f7-a2fc8ec10888"
	uuidVal, _ := uuid.Parse(uuidStr)

	date := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)
	urlValue, _ := url.Parse("https://example.com")

	return map[string]any{
		"foo":         123,
		"bar":         "abc",
		"baz":         true,
		"pi":          3.14,
		"now":         date,
		"big":         int64(9876543210),
		"strslice":    []string{"a", "b"},
		"anyslice":    []any{"c", "d"},
		"smap":        map[string]any{"k": "v"},
		"bytes":       []byte{1, 2, 3},
		"duration":    2 * time.Second,
		"uuidstr":     uuidStr,
		"uuid":        uuidVal,
		"urlstr":      "https://example.com",
		"url":         urlValue,
		"nested":      map[string]any{"deep": map[string]any{"val": 42}},
		"nestedslice": map[string]any{"arr": []any{"x", "y"}},
		"nestedint64": map[string]any{"v": int64(777)},
	}
}

func TestGetter_Get(t *testing.T) {
	t.Parallel()

	conf := config.NewGetter(baseConfigMap())
	timeStamp, ok := baseConfigMap()["now"].(time.Time)
	require.True(t, ok, "failed to cast 'now' to time.Time")
	uuidValue, ok := baseConfigMap()["uuid"].(uuid.UUID)
	require.True(t, ok, "failed to cast 'uuid' to uuid.UUID")

	urlValue, _ := url.Parse("https://example.com")

	tests := []struct {
		name    string
		key     string
		typ     contract.KeyType
		want    interface{}
		wantErr bool
	}{
		{"Int flat", "foo", contract.Int, 123, false},
		{"Int type-cast from int64", "big", contract.Int, 9876543210, false},
		{"String flat", "bar", contract.String, "abc", false},
		{"Bool flat", "baz", contract.Bool, true, false},
		{"Float64 flat", "pi", contract.Float64, 3.14, false},
		{"Time flat", "now", contract.Time, timeStamp, false},
		{"Duration flat", "duration", contract.Duration, 2 * time.Second, false},
		{"Int64 flat", "big", contract.Int64, int64(9876543210), false},
		{"StringSlice flat", "strslice", contract.StringSlice, []string{"a", "b"}, false},
		{"StringSlice from []any", "anyslice", contract.StringSlice, []string{"c", "d"}, false},
		{"Map flat", "smap", contract.Map, map[string]any{"k": "v"}, false},
		{"Bytes flat", "bytes", contract.Bytes, []byte{1, 2, 3}, false},
		{"Uuid from uuid.UUID", "uuid", contract.UUID, uuidValue, false},
		{"Uuid from string", "uuidstr", contract.UUID, uuidValue, false},
		{"Url from *url.URL", "url", contract.URL, urlValue, false},
		{"Url from string", "urlstr", contract.URL, urlValue, false},
		{"Dot notation (map)", "nested.deep.val", contract.Int, 42, false},
		{"Dot notation (slice in map)", "nestedslice.arr.1", contract.String, "y", false},
		{"Dot notation (int64 in map)", "nestedint64.v", contract.Int64, int64(777), false},

		// Error cases
		{"Missing key", "nope", contract.String, nil, true},
		{"Wrong type (string as int)", "bar", contract.Int, nil, true},
		{"Wrong type (int as string)", "foo", contract.String, nil, true},
	}

	for _, testCase := range tests {
		// pin for parallel
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := conf.Get(testCase.key, testCase.typ)
			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

// TestGetter_GetWithType covers the legacy helpers as a group.
func TestGetter_GetWithTypeHelpers(t *testing.T) {
	t.Parallel()

	data := baseConfigMap()
	conf := config.NewGetter(data)
	timeStamp, ok := data["now"].(time.Time)
	require.True(t, ok, "failed to cast 'now' to time.Time")
	uuidValue, ok := data["uuid"].(uuid.UUID)
	require.True(t, ok, "failed to cast 'uuid' to uuid.UUID")

	assert.Equal(t, 123, conf.GetInt("foo"))
	assert.Equal(t, "abc", conf.GetString("bar"))
	assert.True(t, conf.GetBool("baz"))
	assert.InEpsilon(t, 3.14, conf.GetFloat64("pi"), 0.00001)
	assert.Equal(t, timeStamp, conf.GetTime("now"))
	assert.Equal(t, int64(9876543210), conf.GetInt64("big"))
	assert.Equal(t, []string{"a", "b"}, conf.GetStringSlice("strslice"))
	assert.Equal(t, map[string]any{"k": "v"}, conf.GetStringMap("smap"))
	assert.Equal(t, uuidValue.String(), conf.GetString("uuidstr"))
	// "Zero" value for missing key (helpers always return default, not error)
	assert.Equal(t, 0, conf.GetInt("doesnotexist"))
	assert.Equal(t, "", conf.GetString("doesnotexist"))
}

func TestGetter_GetBytes(t *testing.T) {
	t.Parallel()

	conf := config.NewGetter(map[string]any{
		"bytesRaw": []byte{1, 2, 3},
		"bytesStr": "abc",
	})
	b1, err := conf.Get("bytesRaw", contract.Bytes)
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, b1)

	b2, err := conf.Get("bytesStr", contract.Bytes)
	require.NoError(t, err)
	assert.Equal(t, []byte("abc"), b2)
}

func TestGetter_HasKey(t *testing.T) {
	t.Parallel()

	conf := config.NewGetter(baseConfigMap())

	assert.True(t, conf.HasKey("foo"))
	assert.True(t, conf.HasKey("nested.deep.val"))
	assert.False(t, conf.HasKey("nope"))
	assert.True(t, conf.HasKey("strslice.1"))
	assert.False(t, conf.HasKey("missing.key.123"))
}
