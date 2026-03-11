package client

import (
	"encoding/json"
	"fmt"

	"time"

	"github.com/fibrchat/server/pkg/subject"

	"github.com/fibrchat/worker/pkg/message"

	"github.com/nats-io/nats.go"
)

// HandleSend handles sending a message to the specified recipient.
func (c *Client) SendMessage(to, body string) (*message.Response, error) {
	msg := message.Message{
		From:      c.Address(),
		To:        to,
		Body:      body,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	reply, err := c.nc.Request(subject.Send, data, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	var resp message.Response
	if err := json.Unmarshal(reply.Data, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// handleIncoming handles incoming messages....
func (c *Client) HandleIncoming(msg *nats.Msg) {
	if c.onMsg == nil {
		return
	}

	var cm message.Message
	if err := json.Unmarshal(msg.Data, &cm); err != nil {
		return
	}

	c.onMsg(cm)
}
