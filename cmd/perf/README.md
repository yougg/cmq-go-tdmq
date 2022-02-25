# TDMQ-CMQ Benchmark

## 编译构建

> 要求使用最新[Go SDK](https://golang.google.cn/dl/)

```shell
GOOS=linux go generate perf.go
GOOS=darwin go generate perf.go
GOOS=windows go generate perf.go
```

> 编译输出文件: `perf` 或 `perf.exe`

## 编写性能测试用例

> 完整用例配置文件参考  
> 一个用例文件可以包含多条用例, 用例串行执行, 可设置是否启用一条用例

```yaml
---
- Description: example # 用例描述：执行1000次 向1个队列发送1条1KB的消息
  CaseEnabled: false   # 是否启用本用例：true, false
  RepeatTimes: 0       # 用例重复次数：1000次
  RepeatTimeout: 600   # 用例固定重复执行时间, 单位:秒, 非0时 RepeatTimes 配置无效
  Concurrent: 100      # 最大并发数量：100
  MaximumTPS: 0        # 最大限制TPS：0: 不限制，非0: 限制对应数量TPS
  ResourceType: queue  # 请求的资源类型：queue, topic
  ResourceName: queue_ # 请求的资源名称：队列／主题的全名或者前缀，关联下面资源数量(1条时使用全名，多条时使用前缀)
  ResourceCount: 10    # 请求的资源数量：1个或多个队列／主题
  RandMsgSize: true    # 请求的消息体积使用[1 ~ MessageSize]范围内的随机大小
  MessageSize: 1024    # 请求的消息体积：1024B == 1KB，单条消息的体积，批量请求时总体积不能超过64KB
  MessageCount: 1      # 请求的消息数量：1条，每次Action请求消息数量，Batch批量Action请求为1~16条
  Action: SendMessage  # 请求消息的动作：QueryQueueRoute,SendMessage,BatchSendMessage,ReceiveMessage,BatchReceiveMessage,DeleteMessage,BatchDeleteMessage,QueryTopicRoute,PublishMessage,BatchPublishMessage
  ReceiptHandles:      # 请求删除消息ID列表
    - '111'
    - '222'
  DelaySeconds: 123       # 单位为秒，消息发送到队列后，延时多久用户才可见该消息。
  PollingWaitSeconds: 123 # 长轮询等待时间。取值范围0 - 30秒
  RoutingKey: routing_key # 发送消息的路由路径
  Tags:                   # 消息过滤标签
    - tag0
    - tag1
- Description: 执行1000次 向1个队列发送1条1KB的消息
  CaseEnabled: true    # 是否启用本用例：true, false
  RepeatTimes: 1000    # 用例重复次数：1000次
  RepeatTimeout: 600   # 用例固定重复执行时间, 单位:秒, 非0时 RepeatTimes 配置无效
  Concurrent: 100      # 最大并发数量：100
  MaximumTPS: 0        # 最大限制TPS：0: 不限制，非0: 限制对应数量TPS
  ResourceType: queue  # 请求的资源类型：queue, topic
  ResourceName: test   # 请求的资源名称：队列／主题的全名或者前缀，关联下面资源数量(1条时使用全名，多条时使用前缀)
  ResourceCount: 1     # 请求的资源数量：1个或多个队列／主题
  RandMsgSize: true    # 请求的消息体积使用[1 ~ MessageSize]范围内的随机大小
  MessageSize: 1024    # 请求的消息体积：1024B == 1KB，单条消息的体积，批量请求时总体积不能超过64KB
  MessageCount: 1      # 请求的消息数量：1条，每次Action请求消息数量，Batch批量Action请求为1~16条
  Action: SendMessage  # 请求消息的动作：QueryQueueRoute,SendMessage,BatchSendMessage,ReceiveMessage,BatchReceiveMessage,DeleteMessage,BatchDeleteMessage,QueryTopicRoute,PublishMessage,BatchPublishMessage
  DelaySeconds: 0      # 单位为秒，消息发送到队列后，延时多久用户才可见该消息。
- Description: 执行10次 从1个队列接收1条消息
  CaseEnabled: true    # 是否启用本用例：true, false
  RepeatTimes: 10      # 用例重复次数：1000次
  RepeatTimeout: 600   # 用例固定重复执行时间, 单位:秒, 非0时 RepeatTimes 配置无效
  Concurrent: 100      # 最大并发数量：100
  MaximumTPS: 0        # 最大限制TPS：0: 不限制，非0: 限制对应数量TPS
  ResourceType: queue  # 请求的资源类型：queue, topic
  ResourceName: test   # 请求的资源名称：队列／主题的全名或者前缀，关联下面资源数量(1条时使用全名，多条时使用前缀)
  Action: ReceiveMessage  # 请求消息的动作：QueryQueueRoute,SendMessage,BatchSendMessage,ReceiveMessage,BatchReceiveMessage,DeleteMessage,BatchDeleteMessage,QueryTopicRoute,PublishMessage,BatchPublishMessage
  AloneRecvTime: true   # 拉取消息是否分隔Ack进行独立计时
  AckEnabled: true      # 拉取到消息后是否向服务端Ack确认(删除)该条消息
  PollingWaitSeconds: 5 # 长轮询等待时间。取值范围0 - 30秒
```

复制编辑以上`yaml`内容到 https://www.json2yaml.com/ 进行转换`json`  
复制转换后的`json`内容到本地保存为`cases.json`文件

## 执行测试用例

```shell
./perf -h # 查看命令参数帮助

./perf -u 'http://12.34.56.78:9990' -i 'AKIDxxxxx' -k 'abcdefghijk' -c cases.json
```