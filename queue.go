package tdmq

import (
	"fmt"
	"net/url"
	"strconv"
)

// SendMessage
//  API: https://cloud.tencent.com/document/product/406/5837
//  input: queue string
//  input: message string
//  input: delaySeconds int
//  return: *ResponseSM
//  return: error
func (c *Client) SendMessage(queue, message string, delaySeconds int) (*ResponseSM, error) {
	switch {
	case queue == `` || len(queue) > 64:
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

// BatchSendMessage
//  API: https://cloud.tencent.com/document/product/406/5838
//  input: queue string
//  input: messages []string
//  input: delaySeconds int
//  return: *ResponseSMs
//  return: error
func (c *Client) BatchSendMessage(queue string, messages []string, delaySeconds int) (*ResponseSMs, error) {
	switch {
	case queue == `` || len(queue) > 64:
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

// ReceiveMessage
//  API: https://cloud.tencent.com/document/product/406/5839
//  input: queue string
//  input: pollingWaitSeconds int
//  return: *ResponseRM
//  return: error
func (c *Client) ReceiveMessage(queue string, pollingWaitSeconds int) (*ResponseRM, error) {
	switch {
	case queue == `` || len(queue) > 64:
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case pollingWaitSeconds < 0 || pollingWaitSeconds > 30:
		return nil, fmt.Errorf("%w polling wait seconds[0~30]: %d", ErrInvalidParameter, pollingWaitSeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionRecvMsg)
	values.Set(`queueName`, queue)
	values.Set(`pollingWaitSeconds`, strconv.Itoa(pollingWaitSeconds))
	msg, err := c.call(values)
	if err != nil {
		return nil, fmt.Errorf("client call: %w", err)
	}
	resp := &ResponseRM{
		Code:             msg.Code,
		Message:          msg.Message,
		RequestId:        msg.RequestId,
		ClientId:         msg.ClientId,
		MsgId:            msg.MsgId,
		MsgBody:          msg.MsgBody,
		Handle:           msg.Handle,
		EnqueueTime:      msg.EnqueueTime,
		FirstDequeueTime: msg.FirstDequeueTime,
		NextVisibleTime:  msg.NextVisibleTime,
		DequeueCount:     msg.DequeueCount,
	}
	return resp, nil
}

// BatchReceiveMessage
//  API: https://cloud.tencent.com/document/product/406/5924
//  input: queue string
//  input: pollingWaitSeconds int
//  input: numOfMsg int
//  return: *ResponseRMs
//  return: error
func (c *Client) BatchReceiveMessage(queue string, pollingWaitSeconds, numOfMsg int) (*ResponseRMs, error) {
	switch {
	case queue == `` || len(queue) > 64:
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case pollingWaitSeconds < 0 || pollingWaitSeconds > 30:
		return nil, fmt.Errorf("%w polling wait seconds[0~30]: %d", ErrInvalidParameter, pollingWaitSeconds)
	case numOfMsg < 0 || numOfMsg > 16:
		return nil, fmt.Errorf("%w number of message[1~16]: %d", ErrInvalidParameter, pollingWaitSeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionBatchRecv)
	values.Set(`queueName`, queue)
	values.Set(`pollingWaitSeconds`, strconv.Itoa(pollingWaitSeconds))
	values.Set(`numOfMsg`, strconv.Itoa(numOfMsg))
	msg, err := c.call(values)
	if err != nil {
		return nil, err
	}
	resp := &ResponseRMs{
		Code:      msg.Code,
		Message:   msg.Message,
		RequestId: msg.RequestId,
		ClientId:  msg.ClientId,
		MsgInfos:  msg.MsgInfos,
	}
	return resp, nil
}

// DeleteMessage
//  API: https://cloud.tencent.com/document/product/406/5840
//  input: queue string
//  input: receiptHandle string
//  return: *ResponseDM
//  return: error
func (c *Client) DeleteMessage(queue, receiptHandle string) (*ResponseDM, error) {
	switch {
	case queue == `` || len(queue) > 64:
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case receiptHandle == `` || len(receiptHandle) > 80:
		return nil, fmt.Errorf("%w receipt handle(0<len<81): %s", ErrInvalidParameter, receiptHandle)
	}

	values := url.Values{}
	values.Set(`Action`, actionDelMsg)
	values.Set(`queueName`, queue)
	values.Set(`receiptHandle`, receiptHandle)
	msg, err := c.call(values)
	if err != nil {
		return nil, fmt.Errorf("client call: %w", err)
	}
	resp := &ResponseDM{
		Code:      msg.Code,
		Message:   msg.Message,
		RequestId: msg.RequestId,
		ClientId:  msg.ClientId,
	}
	return resp, nil
}

// BatchDeleteMessage
//  API: https://cloud.tencent.com/document/product/406/5841
//  input: queue string
//  input: receiptHandles []string
//  return: *ResponseDMs
//  return: error
func (c *Client) BatchDeleteMessage(queue string, receiptHandles []string) (*ResponseDMs, error) {
	switch {
	case queue == `` || len(queue) > 64:
		return nil, fmt.Errorf("%w queue name(0<len<65): %s", ErrInvalidParameter, queue)
	case len(receiptHandles) == 0 || len(receiptHandles) > 16:
		return nil, fmt.Errorf("%w receipt handle count[0~16]: %v", ErrInvalidParameter, receiptHandles)
	default:
		for _, h := range receiptHandles {
			if h == `` || len(h) > 80 {
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
	msg, err := c.call(values)
	if err != nil {
		return nil, err
	}
	resp := &ResponseDMs{
		Code:      msg.Code,
		Message:   msg.Message,
		RequestId: msg.RequestId,
		ClientId:  msg.ClientId,
		Errors:    msg.Errors,
	}
	return resp, nil
}
