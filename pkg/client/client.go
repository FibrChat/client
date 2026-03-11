package client

import (
	"fmt"
	"time"

	"github.com/fibrchat/server/pkg/subject"

	"github.com/nats-io/nats.go"
)

// Connect establishes a connection to the chat server and subscribes to incoming messages.
func New(o Options) (*Client, error) {
	if o.ServerURL == "" {
		return nil, fmt.Errorf("ServerURL is required")
	}
	if o.Username == "" {
		return nil, fmt.Errorf("Username is required")
	}

	nc, err := nats.Connect(
		o.ServerURL,
		nats.UserInfo(o.Username, o.Password),
		nats.Name(o.Username),
		nats.CustomInboxPrefix(subject.Inbox(o.Username)),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.Timeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	c := &Client{
		nc:       nc,
		username: o.Username,
		domain:   o.Domain,
		onMsg:    o.OnMessage,
	}

	_, err = nc.Subscribe(subject.DM(o.Username), c.HandleIncoming)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	return c, nil
}

// Address returns the full user@domain address.
func (c *Client) Address() string {
	return c.username + "@" + c.domain
}

// Close gracefully closes the connection to the chat server.
func (c *Client) Close() {
	c.nc.Drain()
}
