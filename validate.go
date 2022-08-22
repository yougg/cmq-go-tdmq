package tdmq

import "regexp"

var (
	MaxQueueNameSize  = 64
	MaxTopicNameSize  = 64
	MaxMessageSize    = 1 * 1024 * 1024 // default max message size: 1MB
	MaxMessageCount   = 16
	MaxDelaySeconds   = 70 * 24 * 60 * 60 // 70 day
	MaxWaitSeconds    = 30
	MaxHandleCount    = 16
	MaxHandleLength   = 256
	MaxRouteKeyLength = 64
	MaxRouteKeyDots   = 15
	MaxTagCount       = 5
	MaxTagLength      = 16
)

var (
	InsecureSkipVerify bool
)

var (
	nameReg   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z\d_-]{0,63}$`)
	handleReg = regexp.MustCompile(`^[a-zA-Z\d%#:_-]{0,256}$`)
)
