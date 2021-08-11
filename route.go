package tdmq

import (
	"fmt"
	"net/url"
)

// QueryQueueRoute
//  input: queue string
//  return: *ResponseRoute
//  return: error
func (c *Client) QueryQueueRoute(queue string) (*ResponseRoute, error) {
	return c.query(actionQueueRoute, queue)
}

// QueryTopicRoute
//  input: topic string
//  return: *ResponseRoute
//  return: error
func (c *Client) QueryTopicRoute(topic string) (*ResponseRoute, error) {
	return c.query(actionTopicRoute, topic)
}

// query
//  input: action string
//  input: name string
//  return: *ResponseRoute
//  return: error
func (c *Client) query(action, name string) (*ResponseRoute, error) {
	if name == `` || len(name) > 64 {
		return nil, fmt.Errorf("%w %s name(0<len<65): %s", ErrInvalidParameter, action, name)
	}

	values := url.Values{}
	values.Set(`Action`, action)
	if action == actionQueueRoute {
		values.Set(`queueName`, name)
	} else {
		values.Set(`topicName`, name)
	}
	msg, err := c.call(values)
	if err != nil {
		return nil, fmt.Errorf("client call: %w", err)
	}
	resp := &ResponseRoute{
		Code:      msg.Code,
		Message:   msg.Message,
		RequestId: msg.RequestId,
		ClientId:  msg.ClientId,
		Addr:      msg.Addr,
	}
	return resp, nil
}
