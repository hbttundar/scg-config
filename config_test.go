package scg_config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := New()
	assert.NotNil(t, config)
	assert.NotNil(t, config.viper)
}

func TestNewWithViper(t *testing.T) {
	v := viper.New()
	config := NewWithViper(v)
	assert.NotNil(t, config)
	assert.Equal(t, v, config.viper)
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	configContent := []byte(`
app:
  name: "SCG Config"
  debug: true
database:
  connections:
    pgsql:
      host: "localhost"
      port: 5432
      database: "scg_db"
features:
  analytics:
    enabled: true
pagination:
  default_limit: 25
numbers:
  - 1
  - 2
  - 3
duration: "5s"
timestamp: "2023-01-01T12:00:00Z"
`)
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(configContent); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading from file
	config := New()
	err = config.LoadFromFile(tmpfile.Name())
	assert.NoError(t, err)

	// Test retrieving values
	assert.Equal(t, "SCG Config", config.GetString("app.name"))
	assert.Equal(t, true, config.GetBool("app.debug"))
	assert.Equal(t, "localhost", config.GetString("database.connections.pgsql.host"))
	assert.Equal(t, 5432, config.GetInt("database.connections.pgsql.port"))
	assert.Equal(t, "scg_db", config.GetString("database.connections.pgsql.database"))
	assert.Equal(t, true, config.GetBool("features.analytics.enabled"))
	assert.Equal(t, 25, config.GetInt("pagination.default_limit"))
	assert.Equal(t, []string{"1", "2", "3"}, config.GetStringSlice("numbers"))
	assert.Equal(t, 5*time.Second, config.GetDuration("duration"))

	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	assert.Equal(t, expectedTime, config.GetTime("timestamp"))
}

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_APP_NAME", "SCG Config Env")
	os.Setenv("TEST_APP_DEBUG", "true")
	os.Setenv("TEST_DATABASE_CONNECTIONS_PGSQL_HOST", "localhost")
	os.Setenv("TEST_DATABASE_CONNECTIONS_PGSQL_PORT", "5432")
	defer func() {
		os.Unsetenv("TEST_APP_NAME")
		os.Unsetenv("TEST_APP_DEBUG")
		os.Unsetenv("TEST_DATABASE_CONNECTIONS_PGSQL_HOST")
		os.Unsetenv("TEST_DATABASE_CONNECTIONS_PGSQL_PORT")
	}()

	// Test loading from environment
	config := New()
	config.LoadFromEnv("TEST")

	// Test retrieving values
	assert.Equal(t, "SCG Config Env", config.GetString("app.name"))
	assert.Equal(t, true, config.GetBool("app.debug"))
	assert.Equal(t, "localhost", config.GetString("database.connections.pgsql.host"))
	assert.Equal(t, 5432, config.GetInt("database.connections.pgsql.port"))
}

func TestGetMethods(t *testing.T) {
	config := New()

	// Set some test values
	config.Set("string.value", "test")
	config.Set("int.value", 42)
	config.Set("bool.value", true)
	config.Set("float.value", 3.14)
	config.Set("stringSlice.value", []string{"a", "b", "c"})
	config.Set("stringMap.value", map[string]interface{}{"key1": "value1", "key2": "value2"})
	config.Set("stringMapString.value", map[string]string{"key1": "value1", "key2": "value2"})
	config.Set("duration.value", "10s")
	config.Set("time.value", "2023-01-01T12:00:00Z")

	// Test Get
	assert.Equal(t, "test", config.Get("string.value"))

	// Test GetString
	assert.Equal(t, "test", config.GetString("string.value"))

	// Test GetInt
	assert.Equal(t, 42, config.GetInt("int.value"))

	// Test GetBool
	assert.Equal(t, true, config.GetBool("bool.value"))

	// Test GetFloat64
	assert.Equal(t, 3.14, config.GetFloat64("float.value"))

	// Test GetStringSlice
	assert.Equal(t, []string{"a", "b", "c"}, config.GetStringSlice("stringSlice.value"))

	// Test GetStringMap
	expectedMap := map[string]interface{}{"key1": "value1", "key2": "value2"}
	assert.Equal(t, expectedMap, config.GetStringMap("stringMap.value"))

	// Test GetStringMapString
	expectedMapString := map[string]string{"key1": "value1", "key2": "value2"}
	assert.Equal(t, expectedMapString, config.GetStringMapString("stringMapString.value"))

	// Test GetDuration
	assert.Equal(t, 10*time.Second, config.GetDuration("duration.value"))

	// Test GetTime
	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	assert.Equal(t, expectedTime, config.GetTime("time.value"))
}

func TestNestedDotNotation(t *testing.T) {
	// Create a config with nested values
	config := New()

	// Set a deeply nested structure
	nestedMap := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"string_value": "nested string",
					"int_value":    42,
					"bool_value":   true,
					"float_value":  3.14,
				},
				"array": []string{"a", "b", "c"},
			},
		},
	}

	config.Set("nested", nestedMap)

	// Test accessing nested values with dot notation
	assert.Equal(t, "nested string", config.GetString("nested.level1.level2.level3.string_value"))
	assert.Equal(t, 42, config.GetInt("nested.level1.level2.level3.int_value"))
	assert.Equal(t, true, config.GetBool("nested.level1.level2.level3.bool_value"))
	assert.Equal(t, 3.14, config.GetFloat64("nested.level1.level2.level3.float_value"))

	// Test accessing array in nested structure
	assert.Equal(t, []string{"a", "b", "c"}, config.GetStringSlice("nested.level1.level2.array"))

	// Test accessing non-existent nested values
	assert.Nil(t, config.Get("nested.level1.level2.level3.nonexistent"))
	assert.Equal(t, "", config.GetString("nested.level1.level2.level3.nonexistent"))
	assert.Equal(t, 0, config.GetInt("nested.level1.level2.level3.nonexistent"))
	assert.Equal(t, false, config.GetBool("nested.level1.level2.level3.nonexistent"))
}

func TestEnhancedTraverseNestedMap(t *testing.T) {
	// Create a config with complex nested values
	config := New()

	// Set a structure with arrays and map[string]string
	complexMap := map[string]interface{}{
		"array": []interface{}{"first", "second", "third"},
		"stringMap": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		"nestedArray": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{
					"id":   1,
					"name": "Item 1",
				},
				map[string]interface{}{
					"id":   2,
					"name": "Item 2",
				},
			},
		},
	}

	// Print the complex map for debugging
	t.Logf("Original complexMap: %+v", complexMap)

	// Set the complex map in the config
	config.Set("complex", complexMap)

	// Get the complex map back from the config for debugging
	retrievedMap := config.Get("complex")
	t.Logf("Retrieved complex map: %+v", retrievedMap)

	// Get specific values for debugging
	t.Logf("complex.array: %+v", config.Get("complex.array"))
	t.Logf("complex.stringMap: %+v", config.Get("complex.stringMap"))
	t.Logf("complex.nestedArray: %+v", config.Get("complex.nestedArray"))
	t.Logf("complex.nestedArray.items: %+v", config.Get("complex.nestedArray.items"))

	// Test accessing array elements by index
	arrayValue0 := config.GetString("complex.array.0")
	t.Logf("complex.array.0: %+v", arrayValue0)
	assert.Equal(t, "first", arrayValue0)

	assert.Equal(t, "second", config.GetString("complex.array.1"))
	assert.Equal(t, "third", config.GetString("complex.array.2"))

	// Test accessing map[string]string
	stringMapValue := config.Get("complex.stringMap")
	t.Logf("complex.stringMap: %+v (type: %T)", stringMapValue, stringMapValue)

	stringMapKey1 := config.GetString("complex.stringMap.key1")
	t.Logf("complex.stringMap.key1: %+v", stringMapKey1)
	assert.Equal(t, "value1", stringMapKey1)

	assert.Equal(t, "value2", config.GetString("complex.stringMap.key2"))

	// Test accessing nested array elements
	nestedArrayItems := config.Get("complex.nestedArray.items")
	t.Logf("complex.nestedArray.items: %+v (type: %T)", nestedArrayItems, nestedArrayItems)

	item0Id := config.GetInt("complex.nestedArray.items.0.id")
	t.Logf("complex.nestedArray.items.0.id: %+v", item0Id)
	assert.Equal(t, 1, item0Id)

	assert.Equal(t, "Item 1", config.GetString("complex.nestedArray.items.0.name"))
	assert.Equal(t, 2, config.GetInt("complex.nestedArray.items.1.id"))
	assert.Equal(t, "Item 2", config.GetString("complex.nestedArray.items.1.name"))

	// Test out of bounds array access
	assert.Nil(t, config.Get("complex.array.3"))
	assert.Equal(t, "", config.GetString("complex.array.3"))

	// Test invalid array index
	assert.Nil(t, config.Get("complex.array.invalid"))
	assert.Equal(t, "", config.GetString("complex.array.invalid"))
}

func TestIsSet(t *testing.T) {
	config := New()

	// Set a test value
	config.Set("test.key", "value")

	// Test IsSet with existing key
	assert.True(t, config.IsSet("test.key"))

	// Test IsSet with non-existing key
	assert.False(t, config.IsSet("non.existing.key"))

	// Test Has (deprecated alias) with existing key
	assert.True(t, config.Has("test.key"))
}

func TestGetViper(t *testing.T) {
	config := New()
	assert.NotNil(t, config.GetViper())
	assert.Equal(t, config.viper, config.GetViper())
}

func TestLoadFromDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test YAML file
	yamlContent := []byte(`
name: "App Config"
debug: true
version: 1.0
`)
	yamlFile, err := os.Create(filepath.Join(tempDir, "app.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := yamlFile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	if err := yamlFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create test JSON file
	jsonContent := []byte(`{
  "host": "localhost",
  "port": 5432,
  "credentials": {
    "username": "admin",
    "password": "secret"
  }
}`)
	jsonFile, err := os.Create(filepath.Join(tempDir, "database.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jsonFile.Write(jsonContent); err != nil {
		t.Fatal(err)
	}
	if err := jsonFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create test YML file
	ymlContent := []byte(`
enabled: true
providers:
  - google
  - facebook
  - twitter
settings:
  timeout: 30s
  retries: 3
`)
	ymlFile, err := os.Create(filepath.Join(tempDir, "auth.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ymlFile.Write(ymlContent); err != nil {
		t.Fatal(err)
	}
	if err := ymlFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create a text file that should be ignored
	txtFile, err := os.Create(filepath.Join(tempDir, "readme.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := txtFile.WriteString("This file should be ignored"); err != nil {
		t.Fatal(err)
	}
	if err := txtFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load configuration from directory
	config := New()
	err = config.LoadFromDirectory(tempDir)
	assert.NoError(t, err)

	// Test app.yaml values
	assert.Equal(t, "App Config", config.GetString("app.name"))
	assert.Equal(t, true, config.GetBool("app.debug"))
	assert.Equal(t, 1.0, config.GetFloat64("app.version"))

	// Test database.json values
	assert.Equal(t, "localhost", config.GetString("database.host"))
	assert.Equal(t, 5432, config.GetInt("database.port"))
	assert.Equal(t, "admin", config.GetString("database.credentials.username"))
	assert.Equal(t, "secret", config.GetString("database.credentials.password"))

	// Test auth.yml values
	assert.Equal(t, true, config.GetBool("auth.enabled"))
	assert.Equal(t, []string{"google", "facebook", "twitter"}, config.GetStringSlice("auth.providers"))
	assert.Equal(t, 30*time.Second, config.GetDuration("auth.settings.timeout"))
	assert.Equal(t, 3, config.GetInt("auth.settings.retries"))

	// Test that text file was ignored
	assert.False(t, config.IsSet("readme"))
}

func TestLoadFromDirectoryError(t *testing.T) {
	// Test with non-existent directory
	config := New()
	err := config.LoadFromDirectory("/non/existent/directory")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read directory")
}

func TestLoadFromDirectoryWithOptions(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config_options_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid YAML file
	validContent := []byte(`
name: "Valid Config"
version: 1.0
`)
	validFile, err := os.Create(filepath.Join(tempDir, "valid.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := validFile.Write(validContent); err != nil {
		t.Fatal(err)
	}
	if err := validFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create an invalid YAML file with syntax error
	invalidContent := []byte(`
name: "Invalid Config
version: 1.0
this is not valid yaml
`)
	invalidFile, err := os.Create(filepath.Join(tempDir, "invalid.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := invalidFile.Write(invalidContent); err != nil {
		t.Fatal(err)
	}
	if err := invalidFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test 1: Without continueOnError, should fail on first error
	config1 := New()
	errors1 := config1.LoadFromDirectoryWithOptions(tempDir, false)
	assert.NotEmpty(t, errors1)
	assert.Len(t, errors1, 1)
	assert.Contains(t, errors1[0].Error(), "failed to read config file")

	// Test 2: With continueOnError, should load valid file and report error for invalid file
	config2 := New()
	errors2 := config2.LoadFromDirectoryWithOptions(tempDir, true)
	assert.NotEmpty(t, errors2)
	assert.Len(t, errors2, 1)
	assert.Contains(t, errors2[0].Error(), "failed to read config file")

	// Verify that the valid file was loaded
	assert.Equal(t, "Valid Config", config2.GetString("valid.name"))
	assert.Equal(t, 1.0, config2.GetFloat64("valid.version"))
}

func TestLoadFromDirectoryWithNestedDirectories(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "config_nested_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a nested subdirectory
	nestedSubDir := filepath.Join(subDir, "nested")
	if err := os.Mkdir(nestedSubDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test YAML file in root directory
	rootYamlContent := []byte(`
name: "Root Config"
version: 1.0
`)
	rootYamlFile, err := os.Create(filepath.Join(tempDir, "root.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rootYamlFile.Write(rootYamlContent); err != nil {
		t.Fatal(err)
	}
	if err := rootYamlFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create test JSON file in subdirectory
	subDirJsonContent := []byte(`{
  "name": "SubDir Config",
  "settings": {
    "timeout": 60,
    "enabled": true
  }
}`)
	subDirJsonFile, err := os.Create(filepath.Join(subDir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := subDirJsonFile.Write(subDirJsonContent); err != nil {
		t.Fatal(err)
	}
	if err := subDirJsonFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create test YML file in nested subdirectory
	nestedYmlContent := []byte(`
name: "Nested Config"
items:
  - item1
  - item2
  - item3
deep:
  nested:
    value: 42
`)
	nestedYmlFile, err := os.Create(filepath.Join(nestedSubDir, "settings.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := nestedYmlFile.Write(nestedYmlContent); err != nil {
		t.Fatal(err)
	}
	if err := nestedYmlFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load configuration from directory with nested subdirectories
	config := New()
	err = config.LoadFromDirectory(tempDir)
	assert.NoError(t, err)

	// Test root directory file values
	assert.Equal(t, "Root Config", config.GetString("root.name"))
	assert.Equal(t, 1.0, config.GetFloat64("root.version"))

	// Test subdirectory file values
	assert.Equal(t, "SubDir Config", config.GetString("subdir.config.name"))
	assert.Equal(t, 60, config.GetInt("subdir.config.settings.timeout"))
	assert.Equal(t, true, config.GetBool("subdir.config.settings.enabled"))

	// Test nested subdirectory file values
	assert.Equal(t, "Nested Config", config.GetString("subdir.nested.settings.name"))
	assert.Equal(t, []string{"item1", "item2", "item3"}, config.GetStringSlice("subdir.nested.settings.items"))
	assert.Equal(t, 42, config.GetInt("subdir.nested.settings.deep.nested.value"))
}

func TestMustGetMethods(t *testing.T) {
	config := New()

	// Set some test values
	config.Set("string.value", "test")
	config.Set("int.value", 42)
	config.Set("bool.value", true)
	config.Set("float.value", 3.14)
	config.Set("int64.value", int64(42))
	config.Set("stringSlice.value", []string{"a", "b", "c"})
	config.Set("stringMap.value", map[string]interface{}{"key1": "value1", "key2": "value2"})
	config.Set("stringMapString.value", map[string]string{"key1": "value1", "key2": "value2"})
	config.Set("duration.value", "10s")
	config.Set("time.value", "2023-01-01T12:00:00Z")

	// Test MustGetString with existing key
	assert.Equal(t, "test", config.MustGetString("string.value"))

	// Test MustGetInt with existing key
	assert.Equal(t, 42, config.MustGetInt("int.value"))

	// Test MustGetBool with existing key
	assert.Equal(t, true, config.MustGetBool("bool.value"))

	// Test MustGetFloat64 with existing key
	assert.Equal(t, 3.14, config.MustGetFloat64("float.value"))

	// Test MustGetInt64 with existing key
	assert.Equal(t, int64(42), config.MustGetInt64("int64.value"))

	// Test MustGetStringSlice with existing key
	assert.Equal(t, []string{"a", "b", "c"}, config.MustGetStringSlice("stringSlice.value"))

	// Test MustGetStringMap with existing key
	expectedMap := map[string]interface{}{"key1": "value1", "key2": "value2"}
	assert.Equal(t, expectedMap, config.MustGetStringMap("stringMap.value"))

	// Test MustGetStringMapString with existing key
	expectedMapString := map[string]string{"key1": "value1", "key2": "value2"}
	assert.Equal(t, expectedMapString, config.MustGetStringMapString("stringMapString.value"))

	// Test MustGetDuration with existing key
	assert.Equal(t, 10*time.Second, config.MustGetDuration("duration.value"))

	// Test MustGetTime with existing key
	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	assert.Equal(t, expectedTime, config.MustGetTime("time.value"))

	// Test panic behavior for non-existent keys
	testPanicCases := []struct {
		name string
		fn   func()
	}{
		{"MustGetString", func() { config.MustGetString("nonexistent") }},
		{"MustGetInt", func() { config.MustGetInt("nonexistent") }},
		{"MustGetBool", func() { config.MustGetBool("nonexistent") }},
		{"MustGetFloat64", func() { config.MustGetFloat64("nonexistent") }},
		{"MustGetInt64", func() { config.MustGetInt64("nonexistent") }},
		{"MustGetStringSlice", func() { config.MustGetStringSlice("nonexistent") }},
		{"MustGetStringMap", func() { config.MustGetStringMap("nonexistent") }},
		{"MustGetStringMapString", func() { config.MustGetStringMapString("nonexistent") }},
		{"MustGetDuration", func() { config.MustGetDuration("nonexistent") }},
		{"MustGetTime", func() { config.MustGetTime("nonexistent") }},
	}

	for _, tc := range testPanicCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				assert.NotNil(t, r, "Expected panic but got none")
				assert.Contains(t, r, "Configuration key not found: nonexistent")
			}()
			tc.fn()
			t.Errorf("Expected panic but got none")
		})
	}
}
