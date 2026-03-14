package client

import (
	"github.com/fibrchat/worker/pkg/address"
	"github.com/fibrchat/worker/pkg/event"
	"github.com/fibrchat/worker/pkg/message"
	"github.com/nats-io/nats.go"
)

type EventHandler interface {
	OnConnect(evt event.Event)
	OnDisconnect(evt event.Event)
	OnMessage(msg message.Message)
}

type Client struct {
	nc      *nats.Conn
	addr    address.Address
	handler EventHandler
}

type Options struct {
	ServerURL string
	Username  string
	Password  string
	Domain    string
	Handler   EventHandler
}
