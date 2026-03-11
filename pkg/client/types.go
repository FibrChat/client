package client

import (
	"github.com/fibrchat/worker/pkg/message"
	"github.com/nats-io/nats.go"
)

type Client struct {
	nc       *nats.Conn
	username string
	domain   string
	onMsg    func(message.Message)
}

type Options struct {
	ServerURL string
	Username  string
	Password  string
	Domain    string
	OnMessage func(message.Message)
}
