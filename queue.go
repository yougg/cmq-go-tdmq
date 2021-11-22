package tdmq

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// SendMessage
//  API: https://cloud.tencent.com/document/product/406/5837
//  input: queue string
//  input: message string
//  input: delaySeconds int
//  return: ResponseSM
//  return: error
func (c *Client) SendMessage(queue, message string, delaySeconds int) (ResponseSM, error) {
	switch {
	case !nameReg.MatchString(queue):
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case message == `` || len(message) > 1048576:
		return nil, fmt.Errorf("%w message length(0<len<1048577): %d", ErrInvalidParameter, len(message))
	case delaySeconds < 0 || delaySeconds > 6048000:
		return nil, fmt.Errorf("%w delay seconds[0~6048000]: %d", ErrInvalidParameter, delaySeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionSendMsg)
	values.Set(`queueName`, queue)
	values.Set(`msgBody`, message)
	values.Set(`delaySeconds`, strconv.Itoa(delaySeconds))
	return c.call(values)
}

// BatchSendMessage
//  API: https://cloud.tencent.com/document/product/406/5838
//  input: queue string
//  input: messages []string
//  input: delaySeconds int
//  return: ResponseSMs
//  return: error
func (c *Client) BatchSendMessage(queue string, messages []string, delaySeconds int) (ResponseSMs, error) {
	switch {
	case !nameReg.MatchString(queue):
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case len(messages) == 0 || len(messages) > 16:
		return nil, fmt.Errorf("%w message count(0<len<17): %d", ErrInvalidParameter, len(messages))
	case delaySeconds < 0 || delaySeconds > 6048000:
		return nil, fmt.Errorf("%w delay seconds[0~6048000]: %d", ErrInvalidParameter, delaySeconds)
	default:
		for _, v := range messages {
			if v == `` || len(v) > 1048576 {
				return nil, fmt.Errorf("%w message length(0<len<1048577): %s", ErrInvalidParameter, v)
			}
		}
	}

	values := url.Values{}
	values.Set(`Action`, actionBatchSend)
	values.Set(`queueName`, queue)
	values.Set(`delaySeconds`, strconv.Itoa(delaySeconds))
	for i, m := range messages {
		values.Set(`msgBody.`+strconv.Itoa(i), m)
	}
	return c.call(values)
}

// ReceiveMessage
//  API: https://cloud.tencent.com/document/product/406/5839
//  input: queue string
//  input: pollingWaitSeconds int
//  return: ResponseRM
//  return: error
func (c *Client) ReceiveMessage(queue string, pollingWaitSeconds int) (ResponseRM, error) {
	switch {
	case !nameReg.MatchString(queue):
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case pollingWaitSeconds < 0 || pollingWaitSeconds > 30:
		return nil, fmt.Errorf("%w polling wait seconds[0~30]: %d", ErrInvalidParameter, pollingWaitSeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionRecvMsg)
	values.Set(`queueName`, queue)
	values.Set(`pollingWaitSeconds`, strconv.Itoa(pollingWaitSeconds))

	t := time.Duration(pollingWaitSeconds) * time.Second
	if t > c.client.Timeout {
		c.client.Timeout = t + time.Second
	}
	return c.call(values)
}

// BatchReceiveMessage
//  API: https://cloud.tencent.com/document/product/406/5924
//  input: queue string
//  input: pollingWaitSeconds int
//  input: numOfMsg int
//  return: *ResponseRMs
//  return: error
func (c *Client) BatchReceiveMessage(queue string, pollingWaitSeconds, numOfMsg int) (ResponseRMs, error) {
	switch {
	case !nameReg.MatchString(queue):
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case pollingWaitSeconds < 0 || pollingWaitSeconds > 30:
		return nil, fmt.Errorf("%w polling wait seconds[0~30]: %d", ErrInvalidParameter, pollingWaitSeconds)
	case numOfMsg < 1 || numOfMsg > 16:
		return nil, fmt.Errorf("%w number of message[1~16]: %d", ErrInvalidParameter, pollingWaitSeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionBatchRecv)
	values.Set(`queueName`, queue)
	values.Set(`pollingWaitSeconds`, strconv.Itoa(pollingWaitSeconds))
	values.Set(`numOfMsg`, strconv.Itoa(numOfMsg))

	t := time.Duration(pollingWaitSeconds) * time.Second
	if t > c.client.Timeout {
		c.client.Timeout = t + time.Second
	}
	return c.call(values)
}

// DeleteMessage
//  API: https://cloud.tencent.com/document/product/406/5840
//  input: queue string
//  input: receiptHandle string
//  return: ResponseDM
//  return: error
func (c *Client) DeleteMessage(queue, receiptHandle string) (ResponseDM, error) {
	switch {
	case !nameReg.MatchString(queue):
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case !handleReg.MatchString(receiptHandle):
		return nil, fmt.Errorf("%w receipt handle(0<len<81): %s", ErrInvalidParameter, receiptHandle)
	}

	values := url.Values{}
	values.Set(`Action`, actionDelMsg)
	values.Set(`queueName`, queue)
	values.Set(`receiptHandle`, receiptHandle)
	return c.call(values)
}

// BatchDeleteMessage
//  API: https://cloud.tencent.com/document/product/406/5841
//  input: queue string
//  input: receiptHandles []string
//  return: ResponseDMs
//  return: error
func (c *Client) BatchDeleteMessage(queue string, receiptHandles []string) (ResponseDMs, error) {
	switch {
	case !nameReg.MatchString(queue):
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case len(receiptHandles) == 0 || len(receiptHandles) > 16:
		return nil, fmt.Errorf("%w receipt handle count[0~16]: %v", ErrInvalidParameter, receiptHandles)
	default:
		for _, h := range receiptHandles {
			if !handleReg.MatchString(h) {
				return nil, fmt.Errorf("%w receipt handle(0<len<81): %s", ErrInvalidParameter, h)
			}
		}
	}

	values := url.Values{}
	values.Set(`Action`, actionBatchDel)
	values.Set(`queueName`, queue)
	for i, h := range receiptHandles {
		values.Set(`receiptHandle.`+strconv.Itoa(i), h)
	}
	return c.call(values)
}
