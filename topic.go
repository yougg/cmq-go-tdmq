package tdmq

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Topic struct {
	Client     *Client
	Name       string
	RoutingKey string
	Tags       []string
}

func (t *Topic) Publish(message string) (ResponseSM, error) {
	return t.Client.PublishMessage(t.Name, message, t.RoutingKey, t.Tags)
}

func (t *Topic) BatchPublish(messages ...string) (ResponseSMs, error) {
	return t.Client.BatchPublishMessage(t.Name, t.RoutingKey, messages, t.Tags)
}

// PublishMessage
//  API: https://cloud.tencent.com/document/product/406/7411
//  input: topic string
//  input: message string
//  input: routingKey string
//  input: tags []string
//  return: ResponseSM
//  return: error
func (c *Client) PublishMessage(topic, message, routingKey string, tags []string) (ResponseSM, error) {
	switch {
	case !nameReg.MatchString(topic):
		return nil, fmt.Errorf("%w topic name(0<len<%d): %s", ErrInvalidParameter, MaxTopicNameSize+1, topic)
	case message == `` || len(message) > MaxMessageSize:
		return nil, fmt.Errorf("%w message length(0<len<%d): %d", ErrInvalidParameter, MaxMessageSize+1, len(message))
	case len(routingKey) > MaxRouteKeyLength:
		return nil, fmt.Errorf("%w routing key(0<=len<%d): %s", ErrInvalidParameter, MaxRouteKeyLength+1, routingKey)
	case len(tags) > MaxTagCount:
		return nil, fmt.Errorf("%w message tags count[0~%d]: %v", ErrInvalidParameter, MaxTagCount, tags)
	default:
		if strings.Count(routingKey, `.`) > MaxRouteKeyDots {
			return nil, fmt.Errorf("%w more than %d dot(.) in routing key: %s", ErrInvalidParameter, MaxRouteKeyDots, routingKey)
		}
		for _, v := range tags {
			if v == `` || len(v) > MaxTagLength {
				return nil, fmt.Errorf("%w message tag(0<len<%d): %s", ErrInvalidParameter, MaxTagLength+1, v)
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
	return c.call(values)
}

// BatchPublishMessage
//  API: https://cloud.tencent.com/document/product/406/7412
//  input: topic string
//  input: routingKey string
//  input: messages []string
//  input: tags []string
//  return: ResponseSMs
//  return: error
func (c *Client) BatchPublishMessage(topic, routingKey string, messages, tags []string) (ResponseSMs, error) {
	switch {
	case !nameReg.MatchString(topic):
		return nil, fmt.Errorf("%w topic name(0<len<%d): %s", ErrInvalidParameter, MaxTopicNameSize+1, topic)
	case len(messages) == 0 || len(messages) > MaxMessageCount:
		return nil, fmt.Errorf("%w messages count(0~%d]: %d", ErrInvalidParameter, MaxMessageCount, len(messages))
	case len(routingKey) > MaxRouteKeyLength:
		return nil, fmt.Errorf("%w routing key(0<=len<%d): %s", ErrInvalidParameter, MaxRouteKeyLength+1, routingKey)
	case len(tags) > MaxTagCount:
		return nil, fmt.Errorf("%w message tags count[0~%d]: %v", ErrInvalidParameter, MaxTagCount, tags)
	default:
		if strings.Count(routingKey, `.`) > MaxRouteKeyDots {
			return nil, fmt.Errorf("%w more than %d dot(.) in routing key: %s", ErrInvalidParameter, MaxRouteKeyDots, routingKey)
		}
		for _, v := range messages {
			if v == `` || len(v) > MaxMessageSize {
				return nil, fmt.Errorf("%w message length(0<len<%d): %s", ErrInvalidParameter, MaxMessageSize+1, v)
			}
		}
		for _, v := range tags {
			if v == `` || len(v) > MaxTagLength {
				return nil, fmt.Errorf("%w message tag(0<len<%d): %s", ErrInvalidParameter, MaxTagLength+1, v)
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
	return c.call(values)
}
