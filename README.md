# Tencent TDMQ-CMQ Go SDK

Manage API: https://cloud.tencent.com/document/product/1496/62819

Data Flow API: https://cloud.tencent.com/document/product/1496/61039

Only support these data flow actions:

- Queue
    - QueryQueueRoute
    - SendMessage
    - BatchSendMessage
    - ReceiveMessage
    - BatchReceiveMessage
    - DeleteMessage
    - BatchDeleteMessage

- Topic
    - QueryTopicRoute
    - PublishMessage
    - BatchPublishMessage

Example:

```shell
go get -u github.com/yougg/cmq-go-tdmq@main
```

```go
package main

import (
    "fmt"
    "time"

    tcmq "github.com/yougg/cmq-go-tdmq"
)

func main() {
    // get your own secretId/secretKey: https://console.cloud.tencent.com/cam/capi
    client, err := tcmq.NewClient("https://cmq-gz.public.tencenttdmq.com","AKIDxxxxx","xxxxx",5*time.Second)
    if err != nil {
        fmt.Println("new TDMQ-CMQ client", err)
        return
    }
    // client.AppId = 12345  // for privatization request without authentication
    // client.Method = `GET` // default: POST
    // client.Token = `your_token` // for temporary secretId/secretKey auth with token 
    client.Debug = true // verbose print each request

    queue := &tcmq.Queue{
        Client: client,
        Name:   `queue0`,
        DelaySeconds: 0,
        PollingWaitSeconds: 5,
    }
    resp0, err := queue.Send(`message test 0`)
    if err != nil {
        fmt.Println("send message:", err)
        return
    }
    fmt.Println("Status:", resp0.StatusCode())
    fmt.Println("Response:", resp0)

    msg, err := queue.Receive()
    if err != nil {
        fmt.Println("receive message:", err)
        return
    }
    fmt.Println("Response:", msg)

    resp1, err := queue.Delete(msg.Handle())
    if err != nil {
        fmt.Println("delete message:", err)
        return
    }
    fmt.Println("Response:", resp1)

    resp2, err := queue.BatchSend("a", "b", "c")
    if err != nil {
        fmt.Println("batch send message:", err)
        return
    }
    fmt.Println("Response:", resp2)

    msgs, err := queue.BatchReceive(5)
    if err != nil {
        fmt.Println("batch receive message:", err)
        return
    }
    fmt.Println("Response:", msgs)
    var handles []string
    for _, msg := range msgs.MsgInfos() {
        if len(msg.Handle()) > 0 {
            handles = append(handles, msg.Handle())
        }
    }

    if len(handles) > 0 {
        res, err := queue.BatchDelete(handles...)
        if err != nil {
            fmt.Println("batch delete message:", err)
            return
        }
        fmt.Println("batch delete result:", res)
    }

    topic := &tcmq.Topic{
        Client:     client,
        Name:       `topic0`,
        RoutingKey: ``,
        Tags:       nil,
    }
    resp5, err := topic.Publish(`message test 1`)
    if err != nil {
        fmt.Println("publish message:", err)
        return
    }
    fmt.Println("Response:", resp5)

    msgS, err := topic.BatchPublish("x", "y", "z")
    if err != nil {
        fmt.Println("publish message:", err)
        return
    }
    fmt.Println("Response:", msgS)
}
```