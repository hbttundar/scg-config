package contract

type Provider interface {
	// ReadInConfig Loads/reloads config if supported by backend.
	ReadInConfig() error

	// AllSettings Returns the config as a nested map.
	AllSettings() map[string]interface{}

	// GetKey Flat/fast lookup for a single key (optional, for Provider tests/debug).
	GetKey(key string) any

	// Set Setters for tests or live config editing.
	Set(key string, value any)

	// IsSet Returns true if this key exists (flat only, for legacy/quick lookups).
	IsSet(key string) bool

	// Provider For advanced direct backend use.
	Provider() any

	// SetConfigFile Optionally, for changing file, merging maps, etc.
	SetConfigFile(file string)
	MergeConfigMap(cfg map[string]interface{}) error
}
