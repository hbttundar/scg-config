package contract

// File extensions for supported config formats.
const (
	ExtYAML = ".yaml"
	ExtYML  = ".yml"
	ExtJSON = ".json"
)

// KeyType describes supported type names for config keys.
type KeyType string

const (
	Int         KeyType = "int"
	Int32       KeyType = "int32"
	Int64       KeyType = "int64"
	Uint        KeyType = "uint"
	Uint32      KeyType = "uint32"
	Uint64      KeyType = "uint64"
	Float32     KeyType = "float32"
	Float64     KeyType = "float64"
	String      KeyType = "string"
	Bool        KeyType = "bool"
	StringSlice KeyType = "[]string"
	Map         KeyType = "map"
	Time        KeyType = "time"
	Duration    KeyType = "duration"
	Bytes       KeyType = "bytes"
	UUID        KeyType = "uuid"
	URL         KeyType = "url"
)

// ValueAccessor: type-safe modern API for main config.
type ValueAccessor interface {
	Get(key string, typ KeyType) (any, error)
	Has(key string) bool
}

// Config: core interface for your config service.
type Config interface {
	ValueAccessor
	Provider() Provider
	ReadInConfig() error
	EnvLoader() EnvLoader
	FileLoader() FileLoader
	Watcher() Watcher
	Reload() error
}
