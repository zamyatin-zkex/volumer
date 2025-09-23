package interrupter

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var ErrInterrupted = fmt.Errorf("got interrupt signal")

type Interrupter struct {
}

func (i Interrupter) Run(ctx context.Context) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		return fmt.Errorf("%w: %s", ErrInterrupted, sig.String())
	case <-ctx.Done():
		return fmt.Errorf("interrupter: %w", ctx.Err())
	}
}
