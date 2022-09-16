module github.com/yougg/cmq-go-tdmq/kit/tcmqcli

go 1.19

require (
	github.com/integrii/flaggy v1.5.2
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq v1.0.498
	github.com/yougg/cmq-go-tdmq v0.0.0-00010101000000-000000000000
)

require github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.498

replace (
	github.com/integrii/flaggy => github.com/yougg/flaggy v0.0.0-20220916091504-e64624f2dc64
	github.com/yougg/cmq-go-tdmq => ../..
)
