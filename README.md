# Tencent CMQ to TDMQ compatible Go SDK

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


```shell
go get github.com/yougg/cmq-go-tdmq
```

Example:

```go
package main

import (
    "fmt"

    tdmq "github.com/yougg/cmq-go-tdmq"
)

func main() {
    c := tdmq.NewClient("http://cmq.to.tdmq:12345","AKIDxxxxxxxxxx","ABCDEFGHIJKLMN")
    //c.AppId = 12345 // for privatization without authentication
	//c.Method = `GET` // default: POST
	c.Debug = true
    msg, err := c.SendMessage(`queue0`, `message test`, 0)
    if err != nil {
        fmt.Println("send message:", err)
        return
    }
    fmt.Printf("Response: %#v\n", msg)
}
```