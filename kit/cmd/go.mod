module github.com/yougg/cmq-go-tdmq/kit/cmd

go 1.19

require (
	github.com/integrii/flaggy v1.5.2
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq v1.0.476
	github.com/yougg/cmq-go-tdmq v0.0.0-00010101000000-000000000000
)

require github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.476

replace github.com/yougg/cmq-go-tdmq => ../..
