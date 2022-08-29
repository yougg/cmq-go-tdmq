module github.com/yougg/cmq-go-tdmq/kit/cmd

go 1.19

require (
	github.com/integrii/flaggy v1.5.2
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq v1.0.480
	github.com/yougg/cmq-go-tdmq v0.0.0-00010101000000-000000000000
)

require github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.483

replace (
	github.com/integrii/flaggy => github.com/yougg/flaggy v0.0.0-20220829034919-0d1fbaa1a93b
	github.com/yougg/cmq-go-tdmq => ../..
)
