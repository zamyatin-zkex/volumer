package app

import (
	"context"
)

type Service interface {
	Run(ctx context.Context) error
}

func actor(ctx context.Context, service Service) (func() error, func(err error)) {
	ctx, cancel := context.WithCancelCause(ctx)

	return func() error {
			return service.Run(ctx)
		}, func(err error) {
			cancel(err)
		}
}
