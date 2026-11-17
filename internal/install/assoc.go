package install

type Associator interface {
	Register(exePath string) error
	Unregister() error
}
