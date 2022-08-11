package tdmq

import (
	"errors"
	"fmt"
)

type (
	// Result common result of request TDMQ-CMQ
	Result interface {
		StatusCode() int   // HTTP Response status code
		Code() int         // 0：表示成功，others：错误
		Message() string   // 错误提示信息
		RequestId() string // 服务器生成的请求ID
		ClientId() uint64  // 客户端发送ID
		fmt.Stringer
	}

	// Msg message ID
	Msg interface {
		MsgId() string // 消费的消息唯一标识 ID
	}

	// Message information of response message
	Message interface {
		Msg
		MsgBody() string         // 消费的消息正文
		Handle() string          // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
		EnqueueTime() int64      // 消费被生产出来，进入队列的时间。返回 Unix 时间戳，精确到秒
		FirstDequeueTime() int64 // 保留字段
		NextVisibleTime() int64  // 消息的下次可见（可再次被消费）时间。返回 Unix 时间戳，精确到秒
		DequeueCount() int64     // 保留字段
	}

	// MsgError errors in response
	MsgError interface {
		Code() int       // 0：表示成功，others：错误
		Message() string // 错误提示信息
		Handle() string  // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
	}

	// Route Response of query route
	Route interface {
		Result
		Addr() []string // TDMQ gateway tcp 服务地址
	}

	// ResponseSM Response of send message
	ResponseSM interface {
		Result
		Msg
	}

	// ResponseSMs Response of send messages
	ResponseSMs interface {
		Result
		MsgIDs() []Msg // 服务器生成消息的唯一标识 ID 列表，每个元素是一条消息的信息
	}

	// ResponseRM Response of receive message
	ResponseRM interface {
		Result
		Message
	}

	// ResponseRMs Response of receive messages
	ResponseRMs interface {
		Result
		MsgInfos() []Message // messages 信息列表，每个元素是一条消息的具体信息
	}

	// ResponseDM Response of delete message
	ResponseDM interface {
		Result
	}

	// ResponseDMs Response of delete messages
	ResponseDMs interface {
		Result
		Errors() []MsgError // 无法成功删除的错误列表。每个元素列出了消息无法成功被删除的错误及原因
	}
)

type (
	msgResponse struct {
		Status            int       `json:"-"`                          // HTTP Response status code
		Code_             int       `json:"code"`                       // 0：表示成功，others：错误
		Message_          string    `json:"Message"`                    // 错误提示信息
		RequestId_        string    `json:"requestId"`                  // 服务器生成的请求ID
		ClientId_         uint64    `json:"clientRequestId"`            // 客户端发送ID
		Addr_             []string  `json:"addr,omitempty"`             // TDMQ gateway tcp 服务地址
		MsgId_            string    `json:"msgId,omitempty"`            // 本次的消息唯一标识ID
		MsgBody_          string    `json:"msgBody,omitempty"`          // 本次的消息正文
		Handle_           string    `json:"receiptHandle,omitempty"`    // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
		EnqueueTime_      int64     `json:"enqueueTime,omitempty"`      // 消费被生产出来，进入队列的时间。返回 Unix 时间戳，精确到秒
		FirstDequeueTime_ int64     `json:"firstDequeueTime,omitempty"` // 保留字段
		NextVisibleTime_  int64     `json:"nextVisibleTime,omitempty"`  // 消息的下次可见（可再次被消费）时间。返回 Unix 时间戳，精确到秒
		DequeueCount_     int64     `json:"dequeueCount,omitempty"`     // 保留字段
		MsgIDs_           []msgID   `json:"msgList,omitempty"`          // 服务器生成消息的唯一标识 ID 列表，每个元素是一条消息的信息
		MsgInfos_         []msgInfo `json:"msgInfoList,omitempty"`      // Message 信息列表，每个元素是一条消息的具体信息
		Errors_           []msgErr  `json:"errorList,omitempty"`        // 无法成功删除的错误列表。每个元素列出了消息无法成功被删除的错误及原因
		Raw               string    `json:"-"`
	}

	msgID struct {
		MsgId_ string `json:"msgId,omitempty"` // 消费的消息唯一标识 ID
	}

	msgInfo struct {
		MsgId_            string `json:"msgId"`            // 消费的消息唯一标识 ID
		MsgBody_          string `json:"msgBody"`          // 消费的消息正文
		Handle_           string `json:"receiptHandle"`    // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
		EnqueueTime_      int64  `json:"enqueueTime"`      // 消费被生产出来，进入队列的时间。返回 Unix 时间戳，精确到秒
		FirstDequeueTime_ int64  `json:"firstDequeueTime"` // 保留字段
		NextVisibleTime_  int64  `json:"nextVisibleTime"`  // 消息的下次可见（可再次被消费）时间。返回 Unix 时间戳，精确到秒
		DequeueCount_     int64  `json:"dequeueCount"`     // 保留字段
	}

	msgErr struct {
		Code_    int    `json:"code"`          // 0：表示成功，others：错误
		Message_ string `json:"Message"`       // 错误提示信息
		Handle_  string `json:"receiptHandle"` // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
	}
)

func (m *msgResponse) StatusCode() int         { return m.Status }
func (m *msgResponse) Code() int               { return m.Code_ }
func (m *msgResponse) Message() string         { return m.Message_ }
func (m *msgResponse) RequestId() string       { return m.RequestId_ }
func (m *msgResponse) ClientId() uint64        { return m.ClientId_ }
func (m *msgResponse) Addr() []string          { return m.Addr_ }
func (m *msgResponse) MsgId() string           { return m.MsgId_ }
func (m *msgResponse) MsgBody() string         { return m.MsgBody_ }
func (m *msgResponse) Handle() string          { return m.Handle_ }
func (m *msgResponse) EnqueueTime() int64      { return m.EnqueueTime_ }
func (m *msgResponse) FirstDequeueTime() int64 { return m.FirstDequeueTime_ }
func (m *msgResponse) NextVisibleTime() int64  { return m.NextVisibleTime_ }
func (m *msgResponse) DequeueCount() int64     { return m.DequeueCount_ }
func (m *msgResponse) MsgIDs() (ids []Msg) {
	for i := range m.MsgIDs_ {
		ids = append(ids, &m.MsgIDs_[i])
	}
	return
}
func (m *msgResponse) MsgInfos() (msgs []Message) {
	for i := range m.MsgInfos_ {
		msgs = append(msgs, &m.MsgInfos_[i])
	}
	return
}
func (m *msgResponse) Errors() (errs []MsgError) {
	for i := range m.Errors_ {
		errs = append(errs, &m.Errors_[i])
	}
	return
}
func (m *msgResponse) String() string { return m.Raw }

func (m *msgID) MsgId() string             { return m.MsgId_ }
func (m *msgInfo) MsgId() string           { return m.MsgId_ }
func (m *msgInfo) MsgBody() string         { return m.MsgBody_ }
func (m *msgInfo) Handle() string          { return m.Handle_ }
func (m *msgInfo) EnqueueTime() int64      { return m.EnqueueTime_ }
func (m *msgInfo) FirstDequeueTime() int64 { return m.FirstDequeueTime_ }
func (m *msgInfo) NextVisibleTime() int64  { return m.NextVisibleTime_ }
func (m *msgInfo) DequeueCount() int64     { return m.DequeueCount_ }
func (m *msgErr) Code() int                { return m.Code_ }
func (m *msgErr) Message() string          { return m.Message_ }
func (m *msgErr) Handle() string           { return m.Handle_ }

const (
	currentVersion = "SDK_GO_1.2.0"

	actionQueueRoute = "QueryQueueRoute"
	actionTopicRoute = "QueryTopicRoute"

	actionSendMsg = "SendMessage"
	actionRecvMsg = "ReceiveMessage"
	actionDelMsg  = "DeleteMessage"
	actionPubMsg  = "PublishMessage"

	actionBatchSend = "BatchSendMessage"
	actionBatchRecv = "BatchReceiveMessage"
	actionBatchDel  = "BatchDeleteMessage"
	actionBatchPub  = "BatchPublishMessage"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)
