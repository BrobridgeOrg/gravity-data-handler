package app

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type EventBusImpl interface {
	Emit(string, []byte) error
	On(string, func(*stan.Msg)) error
	GetConnection() *nats.Conn
}

type AppImpl interface {
	GetEventBus() EventBusImpl
}
