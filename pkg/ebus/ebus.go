package ebus

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type EBus struct {
	listeners map[string][]Listener
	mx        sync.RWMutex
}

func New() *EBus {
	return &EBus{
		listeners: make(map[string][]Listener),
	}
}

func (e *EBus) Subscribe(event any, handler Listener) *EBus {
	e.mx.Lock()
	defer e.mx.Unlock()

	name := reflect.TypeOf(event).Name()

	if _, ok := e.listeners[name]; !ok {
		e.listeners[name] = make([]Listener, 0)
	}
	e.listeners[name] = append(e.listeners[name], handler)

	return e
}

func (e *EBus) Emit(ctx context.Context, event any) error {
	e.mx.RLock()
	defer e.mx.RUnlock()

	name := reflect.TypeOf(event).Name()

	if _, ok := e.listeners[name]; !ok {
		return fmt.Errorf("no one listener registered: type %T", event)
	}

	for _, handler := range e.listeners[name] {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
