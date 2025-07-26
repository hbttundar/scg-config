package main

import (
	"fmt"
	"log"

	"github.com/hbttundar/scg-config/config"
	"github.com/hbttundar/scg-config/contract"
)

func main() {
	// --- 1. Initialize configuration service with all loaders, watcher, etc. ---
	cfg := config.New()

	// --- 2. Load configuration from files (directory, any format supported) ---
	if err := cfg.FileLoader().LoadFromDirectory("./examples/config"); err != nil {
		log.Fatalf("Failed to load config directory: %v", err)
	}

	// --- 3. Load/override with environment variables using prefix (e.g. APP_) ---
	if err := cfg.EnvLoader().LoadFromEnv("APP"); err != nil {
		log.Fatalf("Failed to load env config: %v", err)
	}

	// --- 4. Access values using ValueAccessor (type-safe) ---
	port, err := cfg.Get("server.port", contract.Int)
	if err != nil {
		log.Fatalf("Server port missing: %v", err)
	}

	appName, err := cfg.Get("app.name", contract.String)
	if err != nil {
		log.Fatalf("App name missing: %v", err)
	}

	// Optional: fall back to "info" if missing or wrong type
	logLevel, err := cfg.Get("app.loglevel", contract.String)
	if err != nil || logLevel == "" {
		logLevel = "info"
	}

	fmt.Printf("App: %s\nPort: %v\nLogLevel: %v\n", appName, port, logLevel)

	// --- 5. Nested access via dot notation ---
	dbHost, err := cfg.Get("database.host", contract.String)
	if err == nil {
		fmt.Printf("DB Host: %s\n", dbHost)
	}

	// --- 6. Has/Exist checks (works with dot notation) ---
	if cfg.Has("database.host") {
		fmt.Println("Database host is configured!")
	}

	// --- 7. Error handling: show user-friendly message for missing config
	if !cfg.Has("cache.redis.url") {
		fmt.Println("WARNING: Redis cache not configured.")
	}

	// --- 8. Advanced: interface{} generic usage (not recommended, but possible)
	val, err := cfg.Get("advanced.featureToggle", contract.Bool)
	if err == nil && val == true {
		fmt.Println("Feature toggle is enabled")
	}

	// --- 9. Watch for file changes (hot reload, optional) ---
	// cfg.WatchFile("./examples/config/app.yaml")
	// go func() {
	//     for {
	//         // Your logic to react on config reloads
	//     }
	// }()
}
