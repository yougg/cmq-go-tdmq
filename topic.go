package tdmq

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// PublishMessage
//  API: https://cloud.tencent.com/document/product/406/7411
//  input: topic string
//  input: message string
//  input: routingKey string
//  input: tags []string
//  return: *ResponseSM
//  return: error
func (c *Client) PublishMessage(topic, message, routingKey string, tags []string) (*ResponseSM, error) {
	switch {
	case topic == `` || len(topic) > 64:
		return nil, fmt.Errorf("%w topic name(0<len<65): %s", ErrInvalidParameter, topic)
	case message == `` || len(message) > 1048576:
		return nil, fmt.Errorf("%w message length(0<len<1048576): %d", ErrInvalidParameter, len(message))
	case len(routingKey) > 64:
		return nil, fmt.Errorf("%w routing key(0<=len<65): %s", ErrInvalidParameter, routingKey)
	case len(tags) > 5:
		return nil, fmt.Errorf("%w message tags count(0~64): %v", ErrInvalidParameter, tags)
	default:
		if strings.Count(routingKey, `.`) > 15 {
			return nil, fmt.Errorf("%w more than 15 dot(.) in routing key: %s", ErrInvalidParameter, routingKey)
		}
		for _, v := range tags {
			if v == `` || len(v) > 16 {
				return nil, fmt.Errorf("%w message tag(0<len<65): %s", ErrInvalidParameter, v)
			}
		}
	}

	values := url.Values{}
	values.Set(`Action`, actionPubMsg)
	values.Set(`topicName`, topic)
	values.Set(`msgBody`, message)
	values.Set(`routingKey`, routingKey)
	for i, t := range tags {
		values.Set(`msgTag.`+strconv.Itoa(i), t)
	}
	msg, err := c.call(values)
	if err != nil {
		return nil, fmt.Errorf("client call: %w", err)
	}
	resp := &ResponseSM{
		Code:      msg.Code,
		Message:   msg.Message,
		RequestId: msg.RequestId,
		ClientId:  msg.ClientId,
		MsgId:     msg.MsgId,
	}
	return resp, nil
}

// BatchPublishMessage
//  API: https://cloud.tencent.com/document/product/406/7412
//  input: topic string
//  input: routingKey string
//  input: messages []string
//  input: tags []string
//  return: *ResponseSMs
//  return: error
func (c *Client) BatchPublishMessage(topic, routingKey string, messages, tags []string) (*ResponseSMs, error) {
	switch {
	case topic == `` || len(topic) > 64:
		return nil, fmt.Errorf("%w topic name(0<len<65): %s", ErrInvalidParameter, topic)
	case len(messages) == 0 || len(messages) > 16:
		return nil, fmt.Errorf("%w message count(0~16): %d", ErrInvalidParameter, len(messages))
	case len(routingKey) > 64:
		return nil, fmt.Errorf("%w routing key(0<=len<65): %s", ErrInvalidParameter, routingKey)
	case len(tags) > 5:
		return nil, fmt.Errorf("%w message tags count(0~5): %v", ErrInvalidParameter, tags)
	default:
		if strings.Count(routingKey, `.`) > 15 {
			return nil, fmt.Errorf("%w more than 15 dot(.) in routing key: %s", ErrInvalidParameter, routingKey)
		}
		for _, v := range messages {
			if v == `` || len(v) > 1048576 {
				return nil, fmt.Errorf("%w message length(0<len<1048576): %s", ErrInvalidParameter, v)
			}
		}
		for _, v := range tags {
			if v == `` || len(v) > 16 {
				return nil, fmt.Errorf("%w message tag(0<len<17): %s", ErrInvalidParameter, v)
			}
		}
	}

	values := url.Values{}
	values.Set(`Action`, actionBatchPub)
	values.Set(`topicName`, topic)
	values.Set(`routingKey`, routingKey)
	for i, m := range messages {
		values.Set(`msgBody.`+strconv.Itoa(i), m)
	}
	for i, t := range tags {
		values.Set(`msgTag.`+strconv.Itoa(i), t)
	}
	msg, err := c.call(values)
	if err != nil {
		return nil, err
	}
	resp := &ResponseSMs{
		Code:      msg.Code,
		Message:   msg.Message,
		RequestId: msg.RequestId,
		ClientId:  msg.ClientId,
		MsgIDs:    msg.MsgIDs,
	}
	return resp, nil
}
