package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbttundar/scg-config/config"
	"github.com/hbttundar/scg-config/contract"
	"github.com/hbttundar/scg-config/provider/viper"
)

func TestConfig_Get(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("str.int", "420")
	prov.Set("my.int", 123)
	prov.Set("my.str", "abc")
	prov.Set("my.bool", true)

	cfg := config.New(config.WithProvider(prov))

	tests := []struct {
		name     string
		key      string
		keyType  contract.KeyType
		expected interface{}
		hasError bool
	}{
		{"str int can parsed to int", "str.int", contract.Int, 420, false},
		{"existing int", "my.int", contract.Int, 123, false},
		{"existing string", "my.str", contract.String, "abc", false},
		{"existing bool", "my.bool", contract.Bool, true, false},
		{"nonexistent returns error", "missing", contract.Int, nil, true},
		{"type mismatch returns error", "my.str", contract.Int, nil, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := cfg.Get(testCase.key, testCase.keyType)
			if testCase.hasError {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, got)
			}
		})
	}
}

func TestConfig_Has(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("foo", "bar")
	prov.Set("baz", 42)
	cfg := config.New(config.WithProvider(prov))

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"existing key", "foo", true},
		{"existing key 2", "baz", true},
		{"nonexistent key", "nope", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, cfg.Has(tc.key))
		})
	}
}
