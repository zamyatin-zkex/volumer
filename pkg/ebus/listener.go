package ebus

import (
	"context"
	"fmt"
)

type Listener func(ctx context.Context, event interface{}) error

func Typed[T any](fn func(ctx context.Context, typed T) error) Listener {
	return func(ctx context.Context, event interface{}) error {
		typed, ok := event.(T)
		if !ok {
			return fmt.Errorf("invalid event type %T", event)
		}
		return fn(ctx, typed)
	}
}
