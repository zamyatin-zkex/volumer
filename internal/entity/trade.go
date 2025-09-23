package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Trade struct {
	ID        uuid.UUID
	Pair      string
	Maker     string
	Taker     string
	Price     decimal.Decimal
	Volume    decimal.Decimal
	AmoundUSD decimal.Decimal
	Time      time.Time
	// ...

	// todo
	//Offset int64
	// Let’s assume for simplicity that we have one partition
	//Partition int64
}
