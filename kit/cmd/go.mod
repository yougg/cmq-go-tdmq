module github.com/yougg/cmq-go-tdmq/kit/tcmqcli

go 1.20

require (
	github.com/integrii/flaggy v1.5.2
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq v1.0.514
	github.com/yougg/cmq-go-tdmq v0.0.0-00010101000000-000000000000
)

require github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.514

replace (
	github.com/integrii/flaggy => github.com/yougg/flaggy v0.0.0-20220927023241-44a00f282fe3
	github.com/yougg/cmq-go-tdmq => ../..
//github.com/yougg/cmq-go-tdmq/tcp => ../../tcp
)
