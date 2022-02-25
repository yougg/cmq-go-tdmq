//go:generate go build -trimpath -buildmode pie -installsuffix netgo -tags "osusergo netgo static_build" -ldflags "-s -w" ${GOFILE}
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	tdmq "github.com/yougg/cmq-go-tdmq"
)

type list []string

func (l *list) String() string {
	return strings.Join(*l, ",")
}

func (l *list) Set(s string) error {
	*l = append(*l, s)
	return nil
}

var (
	uri string
	sid string
	key string

	queue  string
	topic  string
	action string

	msg  string
	msgs list

	handle  string
	handles list

	tags  list
	route string

	ack     bool
	timeout int
	number  int
	delay   int
	waits   int

	debug bool
)

var (
	client *tdmq.Client
)

func init() {
	flag.StringVar(&uri, "uri", "", "request uri")
	flag.StringVar(&sid, "sid", "", "secret id")
	flag.StringVar(&key, "key", "", "secret key")

	flag.StringVar(&queue, "q", "", "queue name")
	flag.StringVar(&topic, "t", "", "topic name")
	flag.StringVar(&route, "r", "", "routing key")
	flag.StringVar(&action, "a", "", "action: send, receive, delete, publish, sends, receives, deletes, publishes")
	flag.Var(&msgs, "m", "message(s), repeat '-m' 2~16 times to set multi messages")
	flag.Var(&tags, "tag", "tag(s), repeat '-tag' multi times to set multi tags")
	flag.Var(&handles, "handle", "handle(s), repeat '-handle' 2~16 times to set multi handles")

	flag.BoolVar(&ack, "ack", false, "receive message(s) with ack (default false)")
	flag.IntVar(&timeout, "timeout", 5, "client timeout")
	flag.IntVar(&number, "n", 16, "receives <number> messages")
	flag.IntVar(&delay, "delay", 0, "send message(s) <delay> second (default 0)")
	flag.IntVar(&waits, "wait", 5, "receive message(s) <wait> second")

	flag.BoolVar(&debug, "d", false, "print debug log (default false)")

	flag.Parse()
}

func main() {
	var err error
	client, err = tdmq.NewClient(uri, sid, key, time.Duration(timeout)*time.Second)
	if err != nil {
		fmt.Println("new TDMQ-CMQ client", err)
		return
	}
	client.Debug = debug

	switch action {
	case `send`:
		send()
	case `sends`:
		sends()
	case `receive`:
		receive()
	case `receives`:
		receives()
	case `delete`:
		remove()
	case `deletes`:
		removes()
	case `publish`:
		publish()
	case `publishes`:
		publishes()
	default:
		fmt.Println("invalid action:", action)
	}
}

func send() {
	if len(msgs) > 0 {
		msg = msgs[0]
	} else {
		fmt.Println("no message to send, use -m to set message")
		return
	}
	resp0, err := client.SendMessage(queue, msg, delay)
	if err != nil {
		fmt.Printf("send message: %v, response: %v\n", err, resp0)
		return
	}
}

func receive() {
	resp, err := client.ReceiveMessage(queue, waits)
	if err != nil {
		fmt.Println("receive message:", err)
		return
	}
	fmt.Println("Response:", resp)

	if !ack {
		return
	}
	resp1, err := client.DeleteMessage(queue, resp.Handle())
	if err != nil {
		fmt.Println("delete message:", err)
		return
	}
	fmt.Println("Response:", resp1)
}

func remove() {
	if len(handles) > 0 {
		handle = handles[0]
	} else {
		fmt.Println("no handle to delete, use -handle to set handle")
		return
	}
	resp, err := client.DeleteMessage(queue, handle)
	if err != nil {
		fmt.Println("delete message:", err)
		return
	}
	fmt.Println("Response:", resp)
}

func publish() {
	if len(msgs) > 0 {
		msg = msgs[0]
	} else {
		fmt.Println("no message to publish, use -m to set message")
		return
	}
	resp, err := client.PublishMessage(topic, msg, route, tags)
	if err != nil {
		fmt.Println("publish message:", err)
		return
	}
	fmt.Println("Response:", resp)
}

func sends() {
	resp, err := client.BatchSendMessage(queue, msgs, delay)
	if err != nil {
		fmt.Println("batch send message:", err)
		return
	}
	fmt.Println("Response:", resp)
}

func receives() {
	resp, err := client.BatchReceiveMessage(queue, waits, number)
	if err != nil {
		fmt.Println("batch receive message:", err)
		return
	}
	fmt.Println("Response:", resp)

	if !ack {
		return
	}
	handles = nil
	for _, m := range resp.MsgInfos() {
		if len(m.Handle()) > 0 {
			handles = append(handles, m.Handle())
		}
	}

	if len(handles) > 0 {
		res, err := client.BatchDeleteMessage(queue, handles)
		if err != nil {
			fmt.Println("batch delete message:", err)
			return
		}
		fmt.Println("batch delete result:", res)
	}
}

func removes() {
	resp, err := client.BatchDeleteMessage(queue, handles)
	if err != nil {
		fmt.Println("delete messages:", err)
		return
	}
	fmt.Println("Response:", resp)
}

func publishes() {
	resp, err := client.BatchPublishMessage(topic, route, msgs, tags)
	if err != nil {
		fmt.Println("publish message:", err)
		return
	}
	fmt.Println("Response:", resp)
}
