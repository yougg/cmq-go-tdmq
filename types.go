package tdmq

import "errors"

type (
	MsgResponse struct {
		Code             int        `json:"code"`                       // 0：表示成功，others：错误
		Message          string     `json:"message"`                    // 错误提示信息
		RequestId        string     `json:"requestId"`                  // 服务器生成的请求ID
		ClientId         uint64     `json:"clientRequestId"`            // 客户端发送ID
		Addr             []string   `json:"addr,omitempty"`             // TDMQ gateway tcp 服务地址
		MsgId            string     `json:"msgId,omitempty"`            // 本次的消息唯一标识ID
		MsgBody          string     `json:"msgBody,omitempty"`          // 本次的消息正文
		Handle           string     `json:"receiptHandle,omitempty"`    // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
		EnqueueTime      int64      `json:"enqueueTime,omitempty"`      // 消费被生产出来，进入队列的时间。返回 Unix 时间戳，精确到秒
		FirstDequeueTime int64      `json:"firstDequeueTime,omitempty"` // 保留字段
		NextVisibleTime  int64      `json:"nextVisibleTime,omitempty"`  // 消息的下次可见（可再次被消费）时间。返回 Unix 时间戳，精确到秒
		DequeueCount     int64      `json:"dequeueCount,omitempty"`     // 保留字段
		MsgIDs           []MsgID    `json:"msgList,omitempty"`          // 服务器生成消息的唯一标识 ID 列表，每个元素是一条消息的信息
		MsgInfos         []MsgInfo  `json:"msgInfoList,omitempty"`      // message 信息列表，每个元素是一条消息的具体信息
		Errors           []MsgError `json:"errorList,omitempty"`        // 无法成功删除的错误列表。每个元素列出了消息无法成功被删除的错误及原因
	}

	MsgID struct {
		MsgId string `json:"msgId,omitempty"` // 消费的消息唯一标识 ID
	}

	MsgInfo struct {
		MsgId            string `json:"msgId"`            // 消费的消息唯一标识 ID
		MsgBody          string `json:"msgBody"`          // 消费的消息正文
		Handle           string `json:"receiptHandle"`    // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
		EnqueueTime      int64  `json:"enqueueTime"`      // 消费被生产出来，进入队列的时间。返回 Unix 时间戳，精确到秒
		FirstDequeueTime int64  `json:"firstDequeueTime"` // 保留字段
		NextVisibleTime  int64  `json:"nextVisibleTime"`  // 消息的下次可见（可再次被消费）时间。返回 Unix 时间戳，精确到秒
		DequeueCount     int64  `json:"dequeueCount"`     // 保留字段
	}

	MsgError struct {
		Code    int    `json:"code"`          // 0：表示成功，others：错误
		Message string `json:"message"`       // 错误提示信息
		Handle  string `json:"receiptHandle"` // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
	}

	// ResponseRoute Response of query route
	ResponseRoute struct {
		Code      int      `json:"code"`            // 0：表示成功，others：错误
		Message   string   `json:"message"`         // 错误提示信息
		RequestId string   `json:"requestId"`       // 服务器生成的请求ID
		ClientId  uint64   `json:"clientRequestId"` // 客户端发送ID
		Addr      []string `json:"addr,omitempty"`  // TDMQ gateway tcp 服务地址
	}

	// ResponseSM Response of send message
	ResponseSM struct {
		Code      int    `json:"code"`            // 0：表示成功，others：错误
		Message   string `json:"message"`         // 错误提示信息
		RequestId string `json:"requestId"`       // 服务器生成的请求ID
		ClientId  uint64 `json:"clientRequestId"` // 客户端发送ID
		MsgId     string `json:"msgId"`           // 本次的消息唯一标识ID
	}

	// ResponseSMs Response of send messages
	ResponseSMs struct {
		Code      int     `json:"code"`            // 0：表示成功，others：错误
		Message   string  `json:"message"`         // 错误提示信息
		RequestId string  `json:"requestId"`       // 服务器生成的请求ID
		ClientId  uint64  `json:"clientRequestId"` // 客户端发送ID
		MsgIDs    []MsgID `json:"msgList"`         // 服务器生成消息的唯一标识 ID 列表，每个元素是一条消息的信息
	}

	// ResponseRM Response of receive message
	ResponseRM struct {
		Code             int    `json:"code"`                       // 0：表示成功，others：错误
		Message          string `json:"message"`                    // 错误提示信息
		RequestId        string `json:"requestId"`                  // 服务器生成的请求ID
		ClientId         uint64 `json:"clientRequestId"`            // 客户端发送ID
		MsgId            string `json:"msgId,omitempty"`            // 本次的消息唯一标识ID
		MsgBody          string `json:"msgBody,omitempty"`          // 本次的消息正文
		Handle           string `json:"receiptHandle,omitempty"`    // 每次消费返回唯一的消息句柄，用于删除消费。仅上一次消费该消息产生的句柄能用于删除消息。且有效期是 visibilityTimeout，即取出消息隐藏时长，超过该时间后该句柄失效。
		EnqueueTime      int64  `json:"enqueueTime,omitempty"`      // 消费被生产出来，进入队列的时间。返回 Unix 时间戳，精确到秒
		FirstDequeueTime int64  `json:"firstDequeueTime,omitempty"` // 保留字段
		NextVisibleTime  int64  `json:"nextVisibleTime,omitempty"`  // 消息的下次可见（可再次被消费）时间。返回 Unix 时间戳，精确到秒
		DequeueCount     int64  `json:"dequeueCount,omitempty"`     // 保留字段
	}

	// ResponseRMs Response of receive messages
	ResponseRMs struct {
		Code      int       `json:"code"`                  // 0：表示成功，others：错误
		Message   string    `json:"message"`               // 错误提示信息
		RequestId string    `json:"requestId"`             // 服务器生成的请求ID
		ClientId  uint64    `json:"clientRequestId"`       // 客户端发送ID
		MsgInfos  []MsgInfo `json:"msgInfoList,omitempty"` // message 信息列表，每个元素是一条消息的具体信息
	}

	// ResponseDM Response of delete message
	ResponseDM struct {
		Code      int    `json:"code"`            // 0：表示成功，others：错误
		Message   string `json:"message"`         // 错误提示信息
		RequestId string `json:"requestId"`       // 服务器生成的请求ID
		ClientId  uint64 `json:"clientRequestId"` // 客户端发送ID
	}

	// ResponseDMs Response of delete messages
	ResponseDMs struct {
		Code      int        `json:"code"`                // 0：表示成功，others：错误
		Message   string     `json:"message"`             // 错误提示信息
		RequestId string     `json:"requestId"`           // 服务器生成的请求ID
		ClientId  uint64     `json:"clientRequestId"`     // 客户端发送ID
		Errors    []MsgError `json:"errorList,omitempty"` // 无法成功删除的错误列表。每个元素列出了消息无法成功被删除的错误及原因
	}
)

const (
	currentVersion = "SDK_GO_1.0.0"

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
