package client

import (
	"fmt"

	"github.com/fibrchat/server/pkg/subject"
	"github.com/fibrchat/worker/pkg/address"

	"github.com/nats-io/nats.go"
)

// New establishes a connection to the chat server and subscribes to incoming messages.
func New(o Options) (*Client, error) {
	if o.ServerURL == "" {
		return nil, fmt.Errorf("ServerURL is required")
	}
	if o.Username == "" {
		return nil, fmt.Errorf("Username is required")
	}
	if o.Handler == nil {
		return nil, fmt.Errorf("Handler is required")
	}

	nc, err := nats.Connect(
		o.ServerURL,
		nats.MaxReconnects(-1),
		nats.Name("client-"+o.Username),
		nats.UserInfo(o.Username, o.Password),
		nats.CustomInboxPrefix(subject.NATSInbox(o.Username)),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	c := &Client{
		nc:      nc,
		addr:    address.Address{ID: o.Username, Domain: o.Domain},
		handler: o.Handler,
	}

	_, err = nc.Subscribe(subject.Inbox(o.Username), c.handleIncoming)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("subscribe inbox: %w", err)
	}

	_, err = nc.Subscribe(subject.PresenceSubject, c.handlePresence)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("subscribe presence: %w", err)
	}

	return c, nil
}

// Address returns the client's address.
func (c *Client) Address() address.Address {
	return c.addr
}

// Close gracefully closes the connection to the chat server.
func (c *Client) Close() {
	c.nc.Drain()
}
