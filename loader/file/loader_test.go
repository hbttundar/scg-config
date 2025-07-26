package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hbttundar/scg-config/config"
	"github.com/hbttundar/scg-config/contract"
	"github.com/hbttundar/scg-config/loader/file"
	"github.com/hbttundar/scg-config/provider/viper"
)

func TestFileLoader_LoadFromFile_AllSupportedExtensions(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		ext     string
		content string
		key     string
		want    string
	}
	// Supported config file extensions and their syntax
	cases := []testCase{
		{
			name:    "yaml",
			ext:     ".yaml",
			content: "app:\n  name: scg",
			key:     "app.name",
			want:    "scg",
		},
		{
			name:    "yml",
			ext:     ".yml",
			content: "app:\n  name: scg",
			key:     "app.name",
			want:    "scg",
		},
		{
			name:    "json",
			ext:     ".json",
			content: `{"app": {"name": "scg"}}`,
			key:     "app.name",
			want:    "scg",
		},
	}

	for _, testCase := range cases {
		// capture
		t.Run("LoadFromFile_"+testCase.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			tmpFile := filepath.Join(tmpDir, "config"+testCase.ext)
			if err := os.WriteFile(tmpFile, []byte(testCase.content), 0o600); err != nil {
				t.Fatalf("failed to write temp config: %v", err)
			}

			provider := viper.NewConfigProvider()
			loader := file.NewFileLoader(provider)

			err := loader.LoadFromFile(tmpFile)
			if err != nil {
				t.Fatalf("LoadFromFile error: %v", err)
			}

			cfg := config.New(config.WithFileLoader(loader), config.WithProvider(provider))

			val, err := cfg.Get(testCase.key, contract.String)
			if err != nil {
				t.Fatalf("Get(%q, String) error: %v", testCase.key, err)
			}

			if val != testCase.want {
				t.Errorf("Get(%q, String) = %q, want %q", testCase.key, val, testCase.want)
			}
		})
	}
}
