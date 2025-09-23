package entity

import (
	"github.com/shopspring/decimal"
	"github.com/zamyatin-zkex/volumer/pkg/ringbuf"
	"time"
)

type State struct {
	Tokens map[string]Token
	Offset int64
}

type Token struct {
	Name     string
	Periods  map[string]time.Duration
	Buckets  *ringbuf.Ring[Bucket]
	RollSums map[string]decimal.Decimal
	Offset   int64
}

type Bucket struct {
	StartedAt time.Time
	Volume    decimal.Decimal
}
