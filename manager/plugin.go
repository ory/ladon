package manager

import "fmt"

var DefaultManagers = make(map[string]func(...Option) (Manager, error))

func New(kind string, opts ...Option) (Manager, error) {
	newManager, ok := DefaultManagers[kind]
	if !ok {
		return nil, fmt.Errorf("No registered manager plugin %s", kind)
	}
	return newManager(opts...)
}
