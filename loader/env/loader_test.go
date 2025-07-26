package env_test

import (
	"testing"

	"github.com/hbttundar/scg-config/config"
	"github.com/hbttundar/scg-config/contract"
	"github.com/hbttundar/scg-config/loader/env"
	"github.com/hbttundar/scg-config/provider/viper"
)

func TestEnvLoader_LoadFromEnv(t *testing.T) {
	tests := []struct {
		name   string
		envs   map[string]string
		prefix string
		expect map[string]any
		notSet []string
	}{
		{
			name:   "Loads env with prefix",
			envs:   map[string]string{"APP_NAME": "scg", "APP_SERVER_PORT": "8080"},
			prefix: "APP",
			expect: map[string]any{"name": "scg", "server.port": 8080},
			notSet: nil,
		},
		{
			name:   "Ignores envs not matching prefix",
			envs:   map[string]string{"FOO_BAR": "baz"},
			prefix: "APP",
			expect: map[string]any{},
			notSet: []string{"bar"},
		},
		{
			name:   "Loads envs with empty prefix",
			envs:   map[string]string{"FOO_BAR": "baz"},
			prefix: "",
			expect: map[string]any{"foo.bar": "baz"},
			notSet: nil,
		},
	}

	for _, test := range tests {
		// capture range var
		t.Run(test.name, func(t *testing.T) {
			// Set up environment
			for key, value := range test.envs {
				t.Setenv(key, value)
			}

			provider := viper.NewConfigProvider()
			loader := env.NewEnvLoader(provider)

			err := loader.LoadFromEnv(test.prefix)
			if err != nil {
				t.Errorf("cound not initilized EnvLoader with this error: %value", err)
			}

			cfg := config.New(config.WithEnvLoader(loader), config.WithProvider(provider))

			if err != nil {
				t.Fatalf("LoadFromEnv error: %value", err)
			}

			for key, want := range test.expect {
				switch want := want.(type) {
				case string:
					val, err := cfg.Get(key, contract.String)
					if err != nil {
						t.Errorf("Get(%q, String) error: %value", key, err)
					}

					if val != want {
						t.Errorf("Get(%q, String) = %value, want %value", key, val, want)
					}
				case int:
					val, err := cfg.Get(key, contract.Int)
					if err != nil {
						t.Errorf("Get(%q, Int) error: %value", key, err)
					}

					if val != want {
						t.Errorf("Get(%q, Int) = %value, want %value", key, val, want)
					}
				default:
					t.Errorf("Unhandled expected type for key %q: %T", key, want)
				}
			}

			for _, notKey := range test.notSet {
				if cfg.Has(notKey) {
					t.Errorf("Has(%q) = true, want false", notKey)
				}
			}
		})
	}
}
