//go:generate go build -trimpath -buildmode pie -installsuffix netgo -tags "osusergo netgo static_build" -ldflags "-s -w" ${GOFILE}
//go:generate sh -c "[ -z \"${GOEXE}\" ] && gzip -S _${GOOS}_${GOARCH}.gz tcmqcli || zip -mjq tcmqcli_${GOOS}_${GOARCH}.zip tcmqcli${GOEXE}"
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	v20200217 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
	tcmq "github.com/yougg/cmq-go-tdmq"
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

	queue     string
	topic     string
	subscribe string
	action    string

	msg    string
	msgs   list
	length int

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
	region   string
	endpoint string

	filter string
)

var (
	client    *tcmq.Client
	mgrClient *v20200217.Client
)

func init() {
	flag.StringVar(&uri, "uri", "", "request uri")
	flag.StringVar(&sid, "sid", "", "secret id")
	flag.StringVar(&key, "key", "", "secret key")

	flag.StringVar(&queue, "q", "", "queue name")
	flag.StringVar(&topic, "t", "", "topic name")
	flag.StringVar(&route, "r", "", "routing key")
	flag.StringVar(&action, "a", "", "action: query, send, receive, delete, publish, sends, receives, deletes, publishes")
	flag.IntVar(&length, "l", 0, "length: send/publish message with specified length")
	flag.Var(&msgs, "m", "message(s), repeat '-m' 2~16 times to set multi messages")
	flag.Var(&tags, "tag", "tag(s), repeat '-tag' multi times to set multi tags")
	flag.Var(&handles, "handle", "handle(s), repeat '-handle' 2~16 times to set multi handles")

	flag.BoolVar(&ack, "ack", false, "receive message(s) with ack (default false)")
	flag.IntVar(&timeout, "timeout", 5, "client timeout")
	flag.IntVar(&number, "n", 16, "receives <number> messages")
	flag.IntVar(&delay, "delay", 0, "send message(s) <delay> second (default 0)")
	flag.IntVar(&waits, "wait", 5, "receive message(s) <wait> second")

	flag.StringVar(&region, "r", "ap-guangzhou", "region")
	flag.StringVar(&endpoint, "e", "", "endpoint")
	flag.StringVar(&subscribe, "s", "", "subscribe name")
	flag.StringVar(&filter, "f", "", "list filter resource type: queue/topic/subscribe")

	flag.BoolVar(&debug, "d", false, "print debug log (default false)")

	flag.Parse()
}

func main() {
	var err error
	client, err = tcmq.NewClient(uri, sid, key, time.Duration(timeout)*time.Second)
	if err != nil {
		fmt.Println("new TDMQ-CMQ client", err)
		return
	}
	client.Debug = debug

	// 管控API文档: https://cloud.tencent.com/document/product/1496/62819
	credential := common.NewCredential(sid, key)
	prof := profile.NewClientProfile()
	if endpoint != `` {
		prof.HttpProfile.Endpoint = endpoint // 在/etc/hosts中加入映射: 9.223.101.94 tdmq.ap-guangzhou.tencentyun.com
	}
	mgrClient, err = v20200217.NewClient(credential, region, prof)
	if err != nil {
		log.Println(err)
		return
	}
	mgrClient.WithHttpTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	switch action {
	case `query`:
		query()
	case `send`:
		send()
	case `sends`:
		sends()
	case `receive`:
		receive()
	case `receives`:
		receives()
	case `delete`:
		acknowledge()
	case `deletes`:
		acknowledges()
	case `publish`:
		publish()
	case `publishes`:
		publishes()
	case `create`:
		create()
	case `remove`:
		remove()
	case `modify`:
		modify()
	case `describe`:
		describe()
	case `list`:
		lists()
	default:
		fmt.Println("invalid action:", action)
	}
}

func query() {
	switch {
	case queue != ``:
		r, err := client.QueryQueueRoute(queue)
		if err != nil {
			fmt.Println("query topic route:", err)
			return
		}
		if !debug {
			fmt.Println("query queue route:", r)
		}
	case topic != ``:
		r, err := client.QueryTopicRoute(topic)
		if err != nil {
			fmt.Println("query topic route:", err)
			return
		}
		if !debug {
			fmt.Println("query topic route:", r)
		}
	default:
		fmt.Printf("invalid query parameters, queue: %s, topic: %s\n", queue, topic)
	}
}

func send() {
	if len(msgs) > 0 {
		msg = msgs[0]
	} else if length > 0 {
		msg = strings.Repeat("#", length)
	} else {
		fmt.Println("no message to send, use -m to set message")
		return
	}
	resp, err := client.SendMessage(queue, msg, delay)
	if err != nil {
		fmt.Println("send message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}
}

func receive() {
	resp, err := client.ReceiveMessage(queue, waits)
	if err != nil {
		fmt.Println("receive message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}

	if !ack || resp.Code() != 0 {
		return
	}
	resp1, err := client.DeleteMessage(queue, resp.Handle())
	if err != nil {
		fmt.Println("delete message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp1)
	}
}

func acknowledge() {
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
	if !debug {
		fmt.Println("Response:", resp)
	}
}

func publish() {
	if len(msgs) > 0 {
		msg = msgs[0]
	} else if length > 0 {
		msg = strings.Repeat("#", length)
	} else {
		fmt.Println("no message to publish, use -m to set message")
		return
	}
	resp, err := client.PublishMessage(topic, msg, route, tags)
	if err != nil {
		fmt.Println("publish message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}
}

func sends() {
	resp, err := client.BatchSendMessage(queue, msgs, delay)
	if err != nil {
		fmt.Println("batch send message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}
}

func receives() {
	resp, err := client.BatchReceiveMessage(queue, waits, number)
	if err != nil {
		fmt.Println("batch receive message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}

	if !ack || resp.Code() != 0 {
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
		if !debug {
			fmt.Println("batch delete result:", res)
		}
	}
}

func acknowledges() {
	resp, err := client.BatchDeleteMessage(queue, handles)
	if err != nil {
		fmt.Println("delete messages:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}
}

func publishes() {
	resp, err := client.BatchPublishMessage(topic, route, msgs, tags)
	if err != nil {
		fmt.Println("publish message:", err)
		return
	}
	if !debug {
		fmt.Println("Response:", resp)
	}
}

func create() {
	switch {
	case queue != ``:
		qr := v20200217.NewCreateCmqQueueRequest()
		qr.QueueName = &queue
		_, err := mgrClient.CreateCmqQueue(qr)
		if err != nil {
			log.Printf("create queue %s error: %v", *qr.QueueName, err)
			return
		}
	case topic != ``:
		tr := v20200217.NewCreateCmqTopicRequest()
		tr.TopicName = &topic
		_, err := mgrClient.CreateCmqTopic(tr)
		if err != nil {
			log.Printf("create topic %s error: %v", *tr.TopicName, err)
			return
		}
	case subscribe != ``:
		sr := v20200217.NewCreateCmqSubscribeRequest()
		p, ncf := `queue`, `SIMPLIFIED`
		sr.SubscriptionName = &subscribe
		sr.Protocol = &p
		sr.NotifyContentFormat = &ncf
		sr.TopicName = &topic
		sr.Endpoint = &queue
		_, err := mgrClient.CreateCmqSubscribe(sr)
		if err != nil {
			log.Printf("create subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
	}
}

func remove() {
	switch {
	case queue != ``:
		qr := v20200217.NewDeleteCmqQueueRequest()
		qr.QueueName = &queue
		_, err := mgrClient.DeleteCmqQueue(qr)
		if err != nil {
			log.Printf("delete queue %s error: %v", *qr.QueueName, err)
			return
		}
	case topic != ``:
		tr := v20200217.NewDeleteCmqTopicRequest()
		tr.TopicName = &topic
		_, err := mgrClient.DeleteCmqTopic(tr)
		if err != nil {
			log.Printf("delete topic %s error: %v", *tr.TopicName, err)
			return
		}
	case subscribe != ``:
		sr := v20200217.NewDeleteCmqSubscribeRequest()
		sr.SubscriptionName = &subscribe
		sr.TopicName = &topic
		_, err := mgrClient.DeleteCmqSubscribe(sr)
		if err != nil {
			log.Printf("delete subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
	}
}

func modify() {
	switch {
	case queue != ``:
		qr := v20200217.NewModifyCmqQueueAttributeRequest()
		resp, err := mgrClient.ModifyCmqQueueAttribute(qr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *qr.QueueName, err)
			return
		}
		log.Println(resp.ToJsonString())
	case topic != ``:
		tr := v20200217.NewModifyCmqTopicAttributeRequest()
		resp, err := mgrClient.ModifyCmqTopicAttribute(tr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *tr.TopicName, err)
			return
		}
		log.Println(resp.ToJsonString())
	case subscribe != ``:
		sr := v20200217.NewModifyCmqSubscriptionAttributeRequest()
		resp, err := mgrClient.ModifyCmqSubscriptionAttribute(sr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *sr.SubscriptionName, err)
			return
		}
		log.Println(resp.ToJsonString())
	}
}

func describe() {
	switch {
	case queue != ``:
		qr := v20200217.NewDescribeCmqQueueDetailRequest()
		qr.QueueName = &queue
		detail, err := mgrClient.DescribeCmqQueueDetail(qr)
		if err != nil {
			log.Printf("describe queue %s error: %v", *qr.QueueName, err)
			return
		}
		log.Println(detail.ToJsonString())
	case topic != ``:
		tr := v20200217.NewDescribeCmqTopicDetailRequest()
		tr.TopicName = &topic
		detail, err := mgrClient.DescribeCmqTopicDetail(tr)
		if err != nil {
			log.Printf("describe topic %s error: %v", *tr.TopicName, err)
			return
		}
		log.Println(detail.ToJsonString())
	case subscribe != ``:
		sr := v20200217.NewDescribeCmqSubscriptionDetailRequest()
		sr.TopicName = &topic
		sr.SubscriptionName = &subscribe
		detail, err := mgrClient.DescribeCmqSubscriptionDetail(sr)
		if err != nil {
			log.Printf("describe subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		log.Println(detail.ToJsonString())
	}
}

func lists() {
	switch filter {
	case `queue`:
		qr := v20200217.NewDescribeCmqQueuesRequest()
		resp, err := mgrClient.DescribeCmqQueues(qr)
		if err != nil {
			log.Printf("describe queue %s error: %v", *qr.QueueName, err)
			return
		}
		log.Println(resp.ToJsonString())
	case `topic`:
		tr := v20200217.NewDescribeCmqTopicsRequest()
		resp, err := mgrClient.DescribeCmqTopics(tr)
		if err != nil {
			return
		}
		log.Println(resp.ToJsonString())
	case `subscribe`:
		sr := v20200217.NewDescribeSubscriptionsRequest()
		resp, err := mgrClient.DescribeSubscriptions(sr)
		if err != nil {
			return
		}
		log.Println(resp.ToJsonString())
	}
}
