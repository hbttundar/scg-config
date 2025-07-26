package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hbttundar/scg-config/config"
	"github.com/hbttundar/scg-config/contract"
)

// This example demonstrates how to use SCG Config to load configuration from
// multiple sources (YAML, JSON, and environment variables), override values
// programmatically at runtime, query configuration values safely using the
// Get/Has APIs, and react to configuration changes via the built‑in
// watcher.  See the accompanying files under examples/config and .env for
// the data loaded in this example.
func main() {
	// 1. Initialise the configuration service with default provider, loaders and watcher.
	cfg := config.New()

	// 2. Load all supported configuration files from the examples/config directory.
	// Each file's base name becomes a top‑level namespace.  Supported file
	// extensions include .yaml, .yml and .json.  In this repository the
	// app.yaml and database.json files define values under "app", "server",
	// "auth" and "database".
	if err := cfg.FileLoader().LoadFromDirectory("./examples/config"); err != nil {
		log.Fatalf("failed to load config directory: %v", err)
	}

	// 3. Load environment variables with the prefix APP_.  The prefix is stripped
	// and the remaining part is normalised to dot notation.  For example,
	// APP_APP_NAME becomes "app.name" and APP_AUTH_ENABLED becomes "auth.enabled".
	// In this example we source environment values from a .env file; in a real
	// application these would come from the process environment.  To keep this
	// example self‑contained we set the environment variables explicitly.
	// You can also use a package like github.com/joho/godotenv to load a
	// .env file before calling EnvLoader().
	_ = os.Setenv("APP_APP_NAME", "SuperApp")
	_ = os.Setenv("APP_AUTH_ENABLED", "true")
	_ = os.Setenv("APP_LOGLEVEL", "debug")
	if err := cfg.EnvLoader().LoadFromEnv("APP"); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	// 4. Refresh the getter after loading from files and environment.  Reload() must
	// be called whenever the underlying provider's data changes to ensure the
	// Getter sees the latest configuration.
	if err := cfg.Reload(); err != nil {
		log.Fatalf("failed to reload configuration: %v", err)
	}

	// 5. Access scalar values.  The Get method returns a value of type any
	// and an error.  Cast the result to the appropriate Go type.  If the key
	// does not exist or cannot be converted, Get returns an error.
	nameAny, err := cfg.Get("app.name", contract.String)
	if err != nil {
		log.Fatalf("failed to get app.name: %v", err)
	}
	name := nameAny.(string)

	portAny, err := cfg.Get("server.port", contract.Int)
	if err != nil {
		log.Fatalf("failed to get server.port: %v", err)
	}
	port := portAny.(int)

	timeoutAny, err := cfg.Get("server.timeout", contract.Duration)
	var timeout time.Duration
	if err == nil {
		timeout = timeoutAny.(time.Duration)
	}

	fmt.Printf("App: %s\nPort: %d\nTimeout: %s\n", name, port, timeout)

	// 6. Access nested values and slices.  Values under nested maps and lists
	// are available via dot notation.  For example, auth.roles is read from
	// the YAML config.
	authEnabledAny, err := cfg.Get("auth.enabled", contract.Bool)
	if err != nil {
		log.Fatalf("failed to get auth.enabled: %v", err)
	}
	authEnabled := authEnabledAny.(bool)

	rolesAny, err := cfg.Get("auth.roles", contract.StringSlice)
	if err != nil {
		log.Fatalf("failed to get auth.roles: %v", err)
	}
	roles := rolesAny.([]string)

	dbHostAny, err := cfg.Get("database.host", contract.String)
	if err != nil {
		log.Fatalf("failed to get database.host: %v", err)
	}
	dbHost := dbHostAny.(string)

	dbPortAny, err := cfg.Get("database.port", contract.Int)
	if err != nil {
		log.Fatalf("failed to get database.port: %v", err)
	}
	dbPort := dbPortAny.(int)

	fmt.Printf("Auth enabled: %v\nRoles: %v\nDB: %s:%d\n", authEnabled, roles, dbHost, dbPort)

	// 7. Check for optional keys using Has().  This returns true if the key
	// exists in the configuration (including nested map or slice keys).
	if cfg.Has("cache.redis.url") {
		fmt.Println("Redis cache is configured!")
	} else {
		fmt.Println("Redis cache not configured.")
	}

	// 8. Programmatically override a value at runtime by writing to the
	// underlying provider.  After setting a value, call Reload() to update
	// the getter.  Here we change the server port to 9090.
	cfg.Provider().Set("server.port", 9090)
	if err := cfg.Reload(); err != nil {
		log.Fatalf("failed to reload after override: %v", err)
	}
	newPortAny, err := cfg.Get("server.port", contract.Int)
	if err == nil {
		fmt.Printf("Server port overridden to: %d\n", newPortAny.(int))
	}

	// 9. Demonstrate hot reloading by watching the YAML file.  When the file
	// changes, the watcher executes the provided callback.  In the callback we
	// call Reload() to refresh the getter and print the new app.name.
	configFile := "./examples/config/app.yaml"
	if err := cfg.StartWatching(configFile); err != nil {
		log.Fatalf("failed to start watcher: %v", err)
	}
	cfg.Watcher().Watch(func() {
		fmt.Println("app.yaml changed; reloading configuration…")
		if err := cfg.Reload(); err != nil {
			log.Printf("reload error: %v", err)
			return
		}
		if v, err := cfg.Get("app.name", contract.String); err == nil {
			fmt.Println("Updated app.name:", v.(string))
		}
	})

	// 10. Block forever to allow the watcher to run.  Use Ctrl+C to exit.
	select {}
}
