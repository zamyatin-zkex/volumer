package event

import "github.com/zamyatin-zkex/volumer/internal/entity"

type TradeReceived struct {
	entity.Trade

	Offset int64
	// Let’s assume for simplicity that we have one partition
	//Partition int64
}

type TradeSkipped struct {
	//entity.Trade

	Offset int64
}
