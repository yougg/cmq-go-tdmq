module github.com/yougg/cmq-go-tdmq/kit/cmd

go 1.19

require (
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq v1.0.431
	github.com/yougg/cmq-go-tdmq v0.0.0-00010101000000-000000000000
)

require github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.431 // indirect

replace github.com/yougg/cmq-go-tdmq => ../..
