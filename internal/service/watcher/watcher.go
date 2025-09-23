package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"
	"log"
	"reflect"
	"sync"
	"time"
)

type watch struct {
	frame   time.Duration
	getter  func(ctx context.Context) (any, error)
	emitter any
}

type Watcher struct {
	eBus *ebus.EBus
	subs []watch
	mx   sync.Mutex
}

func (w *Watcher) EmitEvery(frame time.Duration, emitter any, getter func(ctx context.Context) (any, error)) *Watcher {
	w.mx.Lock()
	defer w.mx.Unlock()

	w.subs = append(w.subs, watch{frame: frame, getter: getter, emitter: emitter})
	return w
}

func NewWatcher(eBus *ebus.EBus) *Watcher {
	return &Watcher{
		eBus: eBus,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	w.mx.Lock()
	defer w.mx.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error)

	for i := range w.subs {
		go func(i int) {
			sub := w.subs[i]

			ticker := time.NewTicker(sub.frame)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					ins, err := sub.getter(ctx)
					if err != nil {
						select {
						case errs <- err:
						case <-ctx.Done():
						}
						return
					}
					_ = w.eBus.Emit(ctx, ins)
				}
			}
		}(i)
	}

	select {
	case err := <-errs:
		return fmt.Errorf("watcher: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func LogAny[T any](ctx context.Context, event T) error {
	js, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}
	log.Println(reflect.TypeOf(event).Name(), string(js))

	return nil
}
