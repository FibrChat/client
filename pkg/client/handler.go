package client

import (
	"encoding/json"
	"fmt"

	"time"

	"github.com/fibrchat/server/pkg/subject"
	"github.com/fibrchat/worker/pkg/address"
	"github.com/fibrchat/worker/pkg/event"

	"github.com/fibrchat/worker/pkg/message"
	"github.com/fibrchat/worker/pkg/request"

	"github.com/nats-io/nats.go"
)

// SendMessage handles sending a message to the specified recipient.
func (c *Client) SendMessage(dst address.Address, body string) (*request.Response, error) {
	msg := message.Message{
		Src:       c.addr,
		Dst:       dst,
		Content:   body,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	reply, err := c.nc.Request(subject.PublishSubject, data, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	var resp request.Response
	if err := json.Unmarshal(reply.Data, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// ListOnline returns the list of currently online users.
func (c *Client) ListOnline() ([]string, error) {
	reply, err := c.nc.Request(subject.UsersSubject, nil, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("request users: %w", err)
	}

	var resp request.UsersResponse
	if err := json.Unmarshal(reply.Data, &resp); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	return resp.Users, nil
}

// handleIncoming processes incoming messages and dispatches them to the event handler.
func (c *Client) handleIncoming(msg *nats.Msg) {
	var cm message.Message
	if err := json.Unmarshal(msg.Data, &cm); err != nil {
		return
	}

	c.handler.OnMessage(cm)
}

// handlePresence processes presence events (connect/disconnect) and dispatches them to the event handler.
func (c *Client) handlePresence(msg *nats.Msg) {
	var evt event.Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}

	switch evt.Type {
	case event.Connect:
		c.handler.OnConnect(evt)

	case event.Disconnect:
		c.handler.OnDisconnect(evt)
	}
}
