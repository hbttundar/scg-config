package viper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hbttundar/scg-config/provider/viper"
)

const (
	testKey   = "foo"
	testValue = "bar"
	testNum   = 42
)

func TestConfigProvider_Basic(t *testing.T) {
	t.Parallel()

	provider := viper.NewConfigProvider()
	provider.Set(testKey, testValue)

	if v := provider.GetKey(testKey); v != testValue {
		t.Errorf("Get/Set mismatch, got %v", v)
	}

	if !provider.IsSet(testKey) {
		t.Errorf("IsSet false for existing key")
	}

	all := provider.AllSettings()
	if all[testKey] != testValue {
		t.Errorf("AllSettings missing value, got %v", all)
	}
}

func TestConfigProvider_ConfigFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.yaml")
	yaml := []byte("foo: bar\nnum: 42")

	if err := os.WriteFile(path, yaml, 0o600); err != nil { // fixed permission to 0o600 for gosec
		t.Fatalf("write: %v", err)
	}

	provider := viper.NewConfigProvider()
	provider.SetConfigFile(path)

	if err := provider.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	if provider.GetKey(testKey) != testValue || provider.GetKey("num") != testNum {
		t.Errorf("unexpected config: foo=%v num=%v", provider.GetKey(testKey), provider.GetKey("num"))
	}
}
