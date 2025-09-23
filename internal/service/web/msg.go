package web

import (
	"github.com/shopspring/decimal"
	"reflect"
)

type msg struct {
	mType int
	data  []byte
	err   error
}

type BaseMessage struct {
	Name    string
	Payload interface{}
}

func NewMessage(payload interface{}) BaseMessage {
	msg := BaseMessage{
		Name:    reflect.TypeOf(payload).Name(),
		Payload: payload,
	}

	return msg
}

type TokenStats struct {
	Token  string
	Frames []Frame
}

type Frame struct {
	Interval string
	Volume   decimal.Decimal
}
