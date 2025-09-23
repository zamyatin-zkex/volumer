package web

import (
	"github.com/shopspring/decimal"
	"sync"
)

type state struct {
	tokens map[string]map[string]decimal.Decimal // token -> interval -> volume
	mx     sync.RWMutex
}

func newState() *state {
	return &state{
		tokens: make(map[string]map[string]decimal.Decimal),
	}
}

func (s *state) update(tokens map[string]map[string]decimal.Decimal) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.tokens = tokens
}

func (s *state) get(token string) map[string]decimal.Decimal {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.tokens[token]
}
