package contract

type Watcher interface {
	AddFile(path string, callback func()) error
	Watch(callback func())
	Close() error
}
