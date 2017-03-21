package manager

import (
	"context"
	"fmt"
)

var DefaultManagers = make(map[string]func(context.Context, ...Option) (Manager, error))

func New(kind string, ctx context.Context, opts ...Option) (Manager, error) {
	newManager, ok := DefaultManagers[kind]
	if !ok {
		return nil, fmt.Errorf("No registered manager plugin %s", kind)
	}
	return newManager(ctx, opts...)
}
