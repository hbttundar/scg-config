# SCG Config

A configuration utility for Go applications that wraps [spf13/viper](https://github.com/spf13/viper) and provides a Laravel-like dot notation API.

[![CI](https://github.com/hbttundar/scg-config/actions/workflows/ci.yml/badge.svg)](https://github.com/hbttundar/scg-config/actions/workflows/ci.yml)

## Features

- **Dot Notation API**: Access deeply nested configuration values using simple dot notation
- **Type-Safe Methods**: Get configuration values with the correct type (string, int, bool, etc.)
- **Environment Variables**: Load configuration from environment variables
- **Configuration Files**: Load configuration from YAML, JSON, TOML, and other formats supported by Viper
- **Case-Insensitive Keys**: Access configuration values regardless of key case
- **Enhanced Nested Structure Support**: Access array elements by index and navigate complex nested structures
- **Flexible Error Handling**: Options for continuing configuration loading even when some files fail

## Installation

```bash
go get github.com/hbttundar/scg-config
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    config "github.com/hbttundar/scg-config"
)

func main() {
    // Create a new configuration instance
    cfg := config.New()

    // Load configuration from a file
    err := cfg.LoadFromFile("config.yaml")
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Access configuration values using dot notation
    appName := cfg.GetString("app.name")
    debugMode := cfg.GetBool("app.debug")
    dbHost := cfg.GetString("database.connections.pgsql.host")
    dbPort := cfg.GetInt("database.connections.pgsql.port")

    fmt.Printf("App: %s (Debug: %v)\n", appName, debugMode)
    fmt.Printf("Database: %s:%d\n", dbHost, dbPort)
}
```

### Loading from Environment Variables

```go
package main

import (
    "fmt"

    config "github.com/hbttundar/scg-config"
)

func main() {
    // Create a new configuration instance
    cfg := config.New()

    // Load configuration from environment variables with prefix "APP"
    // Environment variables like APP_DATABASE_HOST will be accessible as "database.host"
    cfg.LoadFromEnv("APP")

    // Access configuration values
    dbHost := cfg.GetString("database.host")
    dbPort := cfg.GetInt("database.port")

    fmt.Printf("Database: %s:%d\n", dbHost, dbPort)
}
```

### Loading from Directory

```go
package main

import (
    "fmt"
    "log"

    config "github.com/hbttundar/scg-config"
)

func main() {
    // Create a new configuration instance
    cfg := config.New()

    // Load all .yaml, .yml, and .json files from a directory
    // Each file's name (without extension) becomes the top-level namespace
    // For example, "config/app.yaml" becomes accessible as "app.name"
    err := cfg.LoadFromDirectory("config")
    if err != nil {
        log.Fatalf("Error loading config directory: %v", err)
    }

    // Access configuration values from different files
    appName := cfg.GetString("app.name")
    dbHost := cfg.GetString("database.host")
    authEnabled := cfg.GetBool("auth.enabled")

    fmt.Printf("App: %s\n", appName)
    fmt.Printf("Database: %s\n", dbHost)
    fmt.Printf("Auth Enabled: %v\n", authEnabled)
}
```

### Available Methods

```go
// Get a value with type inference
value := cfg.Get("some.key")

// Get values with specific types
str := cfg.GetString("string.value")
num := cfg.GetInt("int.value")
enabled := cfg.GetBool("feature.enabled")
floatVal := cfg.GetFloat64("float.value")
duration := cfg.GetDuration("timeout.duration")
timestamp := cfg.GetTime("event.timestamp")
strSlice := cfg.GetStringSlice("tags")
strMap := cfg.GetStringMap("metadata")
strMapStr := cfg.GetStringMapString("attributes")

// Get values with panic on missing keys
requiredStr := cfg.MustGetString("required.string")
requiredNum := cfg.MustGetInt("required.int")
requiredBool := cfg.MustGetBool("required.feature")
requiredFloat := cfg.MustGetFloat64("required.float")
requiredDuration := cfg.MustGetDuration("required.timeout")
requiredTime := cfg.MustGetTime("required.timestamp")
requiredSlice := cfg.MustGetStringSlice("required.tags")
requiredMap := cfg.MustGetStringMap("required.metadata")
requiredMapStr := cfg.MustGetStringMapString("required.attributes")

// Check if a key exists
if cfg.Has("optional.feature") {
    // Use the feature
}

// Set a value programmatically
cfg.Set("dynamic.setting", "value")
```

### Advanced: Using with Existing Viper Instance

If you already have a configured Viper instance, you can wrap it with SCG Config:

```go
package main

import (
    "github.com/spf13/viper"
    config "github.com/hbttundar/scg-config"
)

func main() {
    // Create and configure a Viper instance
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")
    v.ReadInConfig()

    // Wrap it with SCG Config
    cfg := config.NewWithViper(v)

    // Now use the dot notation API
    appName := cfg.GetString("app.name")
    // ...
}
```

## License

MIT
