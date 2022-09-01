package tdmq

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type Queue struct {
	Client             *Client
	Name               string
	DelaySeconds       int // 消息延迟可见时间, 1 ~ 6048000 秒
	PollingWaitSeconds int // 消费消息长轮询等待时间, 0 ~ 30 秒
}

// Send message
//  input: message string
//  return: ResponseSM
//  return: error
func (q *Queue) Send(message string) (ResponseSM, error) {
	return q.Client.SendMessage(q.Name, message, q.DelaySeconds)
}

// BatchSend message(s)
//  input: messages ...string
//  return: ResponseSMs
//  return: error
func (q *Queue) BatchSend(messages ...string) (ResponseSMs, error) {
	return q.Client.BatchSendMessage(q.Name, messages, q.DelaySeconds)
}

// Receive message
//  return: ResponseRM
//  return: error
func (q *Queue) Receive() (ResponseRM, error) {
	return q.Client.ReceiveMessage(q.Name, q.PollingWaitSeconds)
}

// BatchReceive message(s)
//  input: numOfMsg int
//  return: *ResponseRMs
//  return: error
func (q *Queue) BatchReceive(numOfMsg int) (ResponseRMs, error) {
	return q.Client.BatchReceiveMessage(q.Name, q.PollingWaitSeconds, numOfMsg)
}

// Delete message handle
//  input: handle string
//  return: ResponseDM
//  return: error
func (q *Queue) Delete(handle string) (ResponseDM, error) {
	return q.Client.DeleteMessage(q.Name, handle)
}

// BatchDelete message handle(s)
//  input: handles ...string
//  return: ResponseDMs
//  return: error
func (q *Queue) BatchDelete(handles ...string) (ResponseDMs, error) {
	return q.Client.BatchDeleteMessage(q.Name, handles)
}

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
		return nil, fmt.Errorf("%w queue name(0<len<%d): %s", ErrInvalidParameter, MaxQueueNameSize+1, queue)
	case message == `` || len(message) > MaxMessageSize:
		return nil, fmt.Errorf("%w message length(0<len<%d): %d", ErrInvalidParameter, MaxMessageSize+1, len(message))
	case delaySeconds < 0 || delaySeconds > MaxDelaySeconds:
		return nil, fmt.Errorf("%w delay seconds[0~%d]: %d", ErrInvalidParameter, MaxDelaySeconds, delaySeconds)
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
		return nil, fmt.Errorf("%w queue name(0<len<%d): %s", ErrInvalidParameter, MaxQueueNameSize+1, queue)
	case len(messages) == 0 || len(messages) > MaxMessageCount:
		return nil, fmt.Errorf("%w message count(0<len<%d): %d", ErrInvalidParameter, MaxMessageCount+1, len(messages))
	case delaySeconds < 0 || delaySeconds > MaxDelaySeconds:
		return nil, fmt.Errorf("%w delay seconds[0~%d]: %d", ErrInvalidParameter, MaxDelaySeconds, delaySeconds)
	default:
		for _, v := range messages {
			if v == `` || len(v) > MaxMessageSize {
				return nil, fmt.Errorf("%w message length(0<len<%d): %s", ErrInvalidParameter, MaxMessageSize+1, v)
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
		return nil, fmt.Errorf("%w queue name(0<len<%d): %s", ErrInvalidParameter, MaxQueueNameSize+1, queue)
	case pollingWaitSeconds < 0 || pollingWaitSeconds > MaxWaitSeconds:
		return nil, fmt.Errorf("%w polling wait seconds[0~%d]: %d", ErrInvalidParameter, MaxWaitSeconds, pollingWaitSeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionRecvMsg)
	values.Set(`queueName`, queue)
	values.Set(`pollingWaitSeconds`, strconv.Itoa(pollingWaitSeconds))

	t := time.Duration(pollingWaitSeconds) * time.Second
	if t > c.HttpClient.Timeout {
		c.HttpClient.Timeout = t + time.Second
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
		return nil, fmt.Errorf("%w queue name(0<len<%d): %s", ErrInvalidParameter, MaxQueueNameSize+1, queue)
	case pollingWaitSeconds < 0 || pollingWaitSeconds > MaxWaitSeconds:
		return nil, fmt.Errorf("%w polling wait seconds[0~%d]: %d", ErrInvalidParameter, MaxWaitSeconds, pollingWaitSeconds)
	case numOfMsg < 1 || numOfMsg > 16:
		return nil, fmt.Errorf("%w number of message[1~%d]: %d", ErrInvalidParameter, MaxMessageCount, pollingWaitSeconds)
	}

	values := url.Values{}
	values.Set(`Action`, actionBatchRecv)
	values.Set(`queueName`, queue)
	values.Set(`pollingWaitSeconds`, strconv.Itoa(pollingWaitSeconds))
	values.Set(`numOfMsg`, strconv.Itoa(numOfMsg))

	t := time.Duration(pollingWaitSeconds) * time.Second
	if t > c.HttpClient.Timeout {
		c.HttpClient.Timeout = t + time.Second
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
		return nil, fmt.Errorf("%w queue name(0<len<%d): %s", ErrInvalidParameter, MaxQueueNameSize+1, queue)
	case !handleReg.MatchString(receiptHandle):
		return nil, fmt.Errorf("%w receipt handle(0<len<%d): %s", ErrInvalidParameter, MaxHandleLength+1, receiptHandle)
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
		return nil, fmt.Errorf("%w queue name(0<len<%d): %s", ErrInvalidParameter, MaxQueueNameSize+1, queue)
	case len(receiptHandles) == 0 || len(receiptHandles) > MaxHandleCount:
		return nil, fmt.Errorf("%w receipt handle count[0~%d]: %v", ErrInvalidParameter, MaxHandleCount, receiptHandles)
	default:
		for _, h := range receiptHandles {
			if !handleReg.MatchString(h) {
				return nil, fmt.Errorf("%w receipt handle(0<len<%d): %s", ErrInvalidParameter, MaxHandleLength+1, h)
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
