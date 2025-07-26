# SCG Config

SCG Config is a configuration library for Go that wraps spf13/viper and exposes a Laravel-like dot notation API.  Its goal is to keep configuration simple, predictable and idiomatic while embracing Go’s conventions.

[![CI](https://github.com/hbttundar/scg-config/actions/workflows/ci.yml/badge.svg)](https://github.com/hbttundar/scg-config/actions/workflows/ci.yml)

## Features

SCG Config offers a concise, type-safe API for working with configuration:

* Dot notation API – Access nested configuration values using a dot syntax (e.g. `app.name` or `database.host`).  Arrays can be traversed by index (e.g. `auth.roles.0`).
* Single `Get` method – Retrieve values via one method by specifying the expected type through the `contract.KeyType` (e.g. `contract.String`, `contract.Int`, `contract.Bool`).  The method returns the value as `any` and an error if the key is missing or cannot be converted.  Use `Has` to check for existence before calling `Get`.
* Multiple sources – Load configuration from YAML, JSON, TOML and other formats supported by Viper, either from a single file or from a directory of files.  Environment variables can also be loaded with an optional prefix.  Values loaded later override earlier ones.
* Case-insensitive keys and nested structures – Keys are normalised to lower-case dot notation, and you can navigate arbitrarily deep maps and arrays.
* Runtime overrides – Mutate configuration at runtime by writing to the underlying provider (`cfg.Provider().Set(key, value)`) and calling `cfg.Reload()` to refresh the getter.
* Hot reloading – Watch configuration files for changes and execute a callback when a file is modified.  In the callback, call `ReadInConfig()` on the provider (if necessary) and `Reload()` on the config to pick up the changes.
* Viper integration – Use the built-in Viper provider or wrap an existing Viper instance to add dot notation and reloading capabilities.

## Installation

go get github.com/hbttundar/scg-config

## Usage

The central type in SCG Config is `*config.Config`, created via `config.New()`.  It embeds a Viper provider and a getter for reading values.  After loading configuration, call `Reload()` to refresh the internal getter with the latest data.

### Loading from files and environment

package main

import (
"fmt"
"log"
"os"

      "github.com/hbttundar/scg-config/config"
      "github.com/hbttundar/scg-config/contract"
)

func main() {
cfg := config.New()

      // Load all .yaml/.yml/.json files from a directory.  Each file’s basename becomes
      // the top-level namespace.
      if err := cfg.FileLoader().LoadFromDirectory("./config"); err != nil {
          log.Fatalf("failed to load directory: %v", err)
      }

      // Load environment variables with prefix APP_.  The prefix is stripped
      // and the rest is normalised to dot notation (e.g. APP_APP_NAME → app.name).
      // Environment values override those from files.
      _ = os.Setenv("APP_APP_NAME", "EnvName")
      if err := cfg.EnvLoader().LoadFromEnv("APP"); err != nil {
          log.Fatalf("failed to load env: %v", err)
      }

      // Refresh the getter after loading.
      if err := cfg.Reload(); err != nil {
          log.Fatalf("failed to reload config: %v", err)
      }

      // Retrieve values.  Specify the type using a contract.KeyType and cast
      // the result to the appropriate Go type.
      nameAny, err := cfg.Get("app.name", contract.String)
      if err != nil {
          log.Fatalf("app.name error: %v", err)
      }
      fmt.Println("Application Name:", nameAny.(string))

      portAny, err := cfg.Get("server.port", contract.Int)
      if err != nil {
          log.Fatalf("server.port error: %v", err)
      }
      fmt.Println("Server Port:", portAny.(int))
}

### Programmatic overrides

To set or override configuration values at runtime, write to the underlying provider and then call `Reload()`:

// Override the log level
cfg.Provider().Set("app.loglevel", "debug")
_ = cfg.Reload()
val, _ := cfg.Get("app.loglevel", contract.String)
fmt.Println("New log level:", val.(string))

### Checking for a key

Use `Has` to check whether a key exists before attempting to read it:

if cfg.Has("feature.newFlag") {
// enable the new feature
}

### Watching for changes

To react to configuration changes at runtime, watch one or more files and reload when they are modified:

// Watch a specific config file
if err := cfg.StartWatching("config/app.yaml"); err != nil {
log.Fatal(err)
}
cfg.Watcher().Watch(func() {
// Refresh the getter when the file changes
if err := cfg.Provider().ReadInConfig(); err != nil {
log.Println("read error:", err)
}
if err := cfg.Reload(); err != nil {
log.Println("reload error:", err)
return
}
if val, err := cfg.Get("app.name", contract.String); err == nil {
fmt.Println("Updated app.name:", val.(string))
}
})

### Using an existing Viper instance

You can use SCG Config with an already configured Viper instance.  This allows you to leverage SCG’s dot notation and getter logic on top of your custom Viper setup.

import (
"github.com/spf13/viper"
"github.com/hbttundar/scg-config/config"
"github.com/hbttundar/scg-config/contract"
)

v := viper.New()
v.SetConfigName("config")
v.SetConfigType("yaml")
v.AddConfigPath("./config")
v.ReadInConfig()

cfg := config.New(config.WithProvider(v))
_ = cfg.Reload()

appName, _ := cfg.Get("app.name", contract.String)
fmt.Println("App Name:", appName.(string))

## License

MIT
