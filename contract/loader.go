package contract

type EnvLoader interface {
	LoadFromEnv(prefix string) error
	GetProvider() Provider
}

type FileLoader interface {
	LoadFromFile(configFile string) error
	LoadFromDirectory(dir string) error
	GetProvider() Provider
}
