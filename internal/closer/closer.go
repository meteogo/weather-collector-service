package closer

import (
	"context"
	"errors"
	"sync"
)

var (
	mu    sync.Mutex
	stack []func(ctx context.Context) error
)

func Add(fn func(ctx context.Context) error) {
	mu.Lock()
	defer mu.Unlock()

	stack = append(stack, fn)
}

func Close(ctx context.Context) error {
	mu.Lock()
	fns := make([]func(ctx context.Context) error, len(stack))
	copy(fns, stack)
	stack = nil
	mu.Unlock()

	var errs []error
	for i := len(fns) - 1; i >= 0; i-- {
		if err := fns[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
