//go:generate go build -trimpath -buildmode pie -installsuffix netgo -tags "osusergo netgo static_build" -ldflags "-s -w" ${GOFILE}
//go:generate sh -c "[ -z \"${GOEXE}\" ] && gzip -S _${GOOS}_${GOARCH}.gz tcmqcli || zip -mjq tcmqcli_${GOOS}_${GOARCH}.zip tcmqcli${GOEXE}"
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	v20200217 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
	tcmq "github.com/yougg/cmq-go-tdmq"
)

type (
	Flag struct {
		Var   any
		Name  string
		Value any
		Usage string
	}

	SubCmd struct {
		Do    func()
		Flags []Flag
	}

	list []string
)

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

	insecure bool
	debug    bool
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

var (
	qFlag   = Flag{&queue, `q`, ``, `queue name`}
	tFlag   = Flag{&topic, `t`, ``, `topic name`}
	sFlag   = Flag{&subscribe, `s`, ``, `subscribe name`}
	rFlag   = Flag{&route, `r`, ``, `routing key`}
	mFlag   = Flag{&msgs, `m`, nil, `message(s), repeat '-m' 2~16 times to set multi messages`}
	lFlag   = Flag{&length, `l`, 0, `length: send/publish message with specified length, unit: byte`}
	tagFlag = Flag{&tags, `tag`, nil, `tag(s), repeat '-tag' multi times to set multi tags`}
	hFlag   = Flag{&handles, `handle`, nil, `handle(s), repeat '-handle' 2~16 times to set multi handles`}
	ackFlag = Flag{&ack, `ack`, false, `receive message(s) with ack (default false)`}
	nFlag   = Flag{&number, `n`, 16, `sends/receives <number> messages`}
	dFlag   = Flag{&delay, `delay`, 0, `send message(s) <delay> second (default 0)`}
	wFlag   = Flag{&waits, `wait`, 5, `receive message(s) <wait> second`}
	fFlag   = Flag{&filter, `f`, ``, `list filter resource type: queue/topic/subscribe`}
	dbgFlag = Flag{&debug, "d", false, "print debug log (default false)"}
	uFlag   = Flag{&uri, "uri", "", "request uri"}
	rgFlag  = Flag{&region, "region", "ap-guangzhou", "region"}
	epFlag  = Flag{&endpoint, "e", "", "endpoint"}

	actionCmds = map[string]SubCmd{
		`query`:     {Do: query, Flags: []Flag{dbgFlag, uFlag, qFlag, tFlag}},
		`send`:      {Do: send, Flags: []Flag{dbgFlag, uFlag, qFlag, mFlag, lFlag, dFlag}},
		`sends`:     {Do: sends, Flags: []Flag{dbgFlag, uFlag, qFlag, mFlag, lFlag, dFlag}},
		`receive`:   {Do: receive, Flags: []Flag{dbgFlag, uFlag, qFlag, wFlag, ackFlag}},
		`receives`:  {Do: receives, Flags: []Flag{dbgFlag, uFlag, qFlag, wFlag, nFlag, ackFlag}},
		`delete`:    {Do: acknowledge, Flags: []Flag{dbgFlag, uFlag, qFlag, hFlag}},
		`deletes`:   {Do: acknowledges, Flags: []Flag{dbgFlag, uFlag, qFlag, hFlag}},
		`publish`:   {Do: publish, Flags: []Flag{dbgFlag, uFlag, tFlag, mFlag, lFlag, rFlag, tagFlag}},
		`publishes`: {Do: publishes, Flags: []Flag{dbgFlag, uFlag, tFlag, mFlag, nFlag, lFlag, rFlag, tagFlag}},
		`create`:    {Do: create, Flags: []Flag{rgFlag, epFlag, qFlag, tFlag, sFlag}},   // TODO replenish flags
		`remove`:    {Do: remove, Flags: []Flag{rgFlag, epFlag, qFlag, tFlag, sFlag}},   // TODO replenish flags
		`modify`:    {Do: modify, Flags: []Flag{rgFlag, epFlag, qFlag, tFlag, sFlag}},   // TODO replenish flags
		`describe`:  {Do: describe, Flags: []Flag{rgFlag, epFlag, qFlag, tFlag, sFlag}}, // TODO replenish flags
		`list`:      {Do: lists, Flags: []Flag{rgFlag, epFlag, qFlag, tFlag, sFlag, fFlag}},
	}

	flagSets = map[string]*flag.FlagSet{}
)

func init() {
	for a, cmd := range actionCmds {
		fs := flag.NewFlagSet(a, flag.ExitOnError)
		for _, f := range cmd.Flags {
			switch v := f.Var.(type) {
			case *string:
				fs.StringVar(v, f.Name, f.Value.(string), f.Usage)
			case *int:
				fs.IntVar(v, f.Name, f.Value.(int), f.Usage)
			case *bool:
				fs.BoolVar(v, f.Name, f.Value.(bool), f.Usage)
			case flag.Value:
				fs.Var(v, f.Name, f.Usage)
			}
		}
		fs.BoolVar(&insecure, "insecure", false, "whether client verifies server's certificate")
		fs.IntVar(&timeout, "timeout", 5, "client timeout in seconds")
		fs.StringVar(&sid, "sid", "", "secret id")
		fs.StringVar(&key, "key", "", "secret key")
		flagSets[a] = fs
	}
}

func init() {
	flag.Usage = func() {
		name := filepath.Base(os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(),
			"Usage of %s with sub commands:\n  %s\n\nShow sub cmd usage:\n  %s %s\n  %s %s\n\nExample:\n  %s %s\n  %s %s\n",
			name,
			"query,send,sends,receive,receives,delete,deletes,publish,publishes\n  create,remove,modify,describe,list",
			name, `<subcmd> -h`, name, `<subcmd> --help`,
			name, `send -d -uri https://cmq-gz.public.tencenttdmq.com -sid AKID... -key xxx -q test -l 10`,
			name, `receives -uri https://cmq-gz.public.tencenttdmq.com -sid AKID... -key xxx -q test -n 5`)
	}
	flag.Parse()

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], `-`) {
		action = os.Args[1]
	}
	if cmd, ok := flagSets[action]; ok {
		err := cmd.Parse(os.Args[2:])
		if err != nil {
			log.Println(err)
			cmd.Usage()
		}
	} else {
		log.Fatalln("invalid action:", action)
	}
}

func main() {
	var err error
	switch action {
	case `query`, `send`, `sends`, `receive`, `receives`, `delete`, `deletes`, `publish`, `publishes`:
		tcmq.InsecureSkipVerify = insecure
		client, err = tcmq.NewClient(uri, sid, key, time.Duration(timeout)*time.Second)
		if err != nil {
			log.Println("new TCMQ client", err)
			return
		}
		client.Debug = debug
	case `create`, `remove`, `modify`, `describe`, `list`:
		// 管控API文档: https://cloud.tencent.com/document/product/1496/62819
		prof := profile.NewClientProfile()
		prof.HttpProfile.ReqTimeout = timeout
		if endpoint != `` {
			prof.HttpProfile.Endpoint = endpoint
		}
		mgrClient, err = v20200217.NewClient(common.NewCredential(sid, key), region, prof)
		if err != nil {
			log.Println("new TCMQ manager client", err)
			return
		}
		mgrClient.WithHttpTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		})
	}

	actionCmds[action].Do()
}

func query() {
	switch {
	case queue != ``:
		r, err := client.QueryQueueRoute(queue)
		if err != nil {
			log.Println("query queue route:", err)
			return
		}
		if !debug {
			fmt.Println(r)
		}
	case topic != ``:
		r, err := client.QueryTopicRoute(topic)
		if err != nil {
			log.Println("query topic route:", err)
			return
		}
		if !debug {
			fmt.Println(r)
		}
	default:
		log.Printf("invalid query parameters, queue: %s, topic: %s\n", queue, topic)
	}
}

func send() {
	if len(msgs) > 0 {
		msg = msgs[0]
	} else if length > 0 {
		msg = strings.Repeat("#", length)
		if length > tcmq.MaxMessageSize {
			tcmq.MaxMessageSize = length
		}
	} else {
		log.Println("no message to send, use -m to set message")
		return
	}
	resp, err := client.SendMessage(queue, msg, delay)
	if err != nil {
		log.Println("send message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func receive() {
	resp, err := client.ReceiveMessage(queue, waits)
	if err != nil {
		log.Println("receive message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}

	if !ack || resp.Code() != 0 {
		return
	}
	resp1, err := client.DeleteMessage(queue, resp.Handle())
	if err != nil {
		log.Println("delete message:", err)
		return
	}
	if !debug {
		fmt.Println(resp1)
	}
}

func acknowledge() {
	if len(handles) > 0 {
		handle = handles[0]
	} else {
		log.Println("no handle to delete, use -handle to set handle")
		return
	}
	resp, err := client.DeleteMessage(queue, handle)
	if err != nil {
		log.Println("delete message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func publish() {
	switch {
	case len(msgs) > 0 && len(msgs[0]) > 0:
		msg = msgs[0]
	case length > 0:
		msg = strings.Repeat("#", length)
		if length > tcmq.MaxMessageSize {
			tcmq.MaxMessageSize = length
		}
	default:
		log.Println("no message to publish, use -m to set message")
		return
	}
	resp, err := client.PublishMessage(topic, msg, route, tags)
	if err != nil {
		log.Println("publish message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func sends() {
	switch {
	case len(msgs) > 0:
		for i := range msgs {
			if len(msgs[i]) == 0 {
				log.Println("message is empty")
				return
			}
		}
	case length > 0:
		msg = strings.Repeat(`#`, length)
		if length > tcmq.MaxMessageSize {
			tcmq.MaxMessageSize = length
		}
		if number > 1 {
			msgs = make(list, 0, number)
			for i := 0; i < number; i++ {
				msgs = append(msgs, msg)
			}
		} else {
			msgs = list{msg}
		}
	default:
		log.Println("no message to send, use -m to set message")
		return
	}
	resp, err := client.BatchSendMessage(queue, msgs, delay)
	if err != nil {
		log.Println("batch send message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func receives() {
	resp, err := client.BatchReceiveMessage(queue, waits, number)
	if err != nil {
		log.Println("batch receive message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
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
			log.Println("batch delete message:", err)
			return
		}
		if !debug {
			fmt.Println(res)
		}
	}
}

func acknowledges() {
	resp, err := client.BatchDeleteMessage(queue, handles)
	if err != nil {
		log.Println("delete messages:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func publishes() {
	switch {
	case len(msgs) > 0:
		for i := range msgs {
			if len(msgs[i]) == 0 {
				log.Println("message is empty")
				return
			}
		}
	case length > 0:
		msg = strings.Repeat(`#`, length)
		if length > tcmq.MaxMessageSize {
			tcmq.MaxMessageSize = length
		}
		if number > 1 {
			msgs = make(list, 0, number)
			for i := 0; i < number; i++ {
				msgs = append(msgs, msg)
			}
		} else {
			msgs = list{msg}
		}
	default:
		log.Println("no message to publish, use -m to set message")
		return
	}
	resp, err := client.BatchPublishMessage(topic, route, msgs, tags)
	if err != nil {
		log.Println("publish message:", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func create() {
	switch {
	case queue != ``:
		qr := v20200217.NewCreateCmqQueueRequest()
		qr.QueueName = &queue
		resp, err := mgrClient.CreateCmqQueue(qr)
		if err != nil {
			log.Printf("create queue %s error: %v", *qr.QueueName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case topic != ``:
		tr := v20200217.NewCreateCmqTopicRequest()
		tr.TopicName = &topic
		resp, err := mgrClient.CreateCmqTopic(tr)
		if err != nil {
			log.Printf("create topic %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case subscribe != ``:
		sr := v20200217.NewCreateCmqSubscribeRequest()
		p, ncf := `queue`, `SIMPLIFIED`
		sr.SubscriptionName = &subscribe
		sr.Protocol = &p
		sr.NotifyContentFormat = &ncf
		sr.TopicName = &topic
		sr.Endpoint = &queue
		resp, err := mgrClient.CreateCmqSubscribe(sr)
		if err != nil {
			log.Printf("create subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	}
}

func remove() {
	switch {
	case queue != ``:
		qr := v20200217.NewDeleteCmqQueueRequest()
		qr.QueueName = &queue
		resp, err := mgrClient.DeleteCmqQueue(qr)
		if err != nil {
			log.Printf("delete queue %s error: %v", *qr.QueueName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case topic != ``:
		tr := v20200217.NewDeleteCmqTopicRequest()
		tr.TopicName = &topic
		resp, err := mgrClient.DeleteCmqTopic(tr)
		if err != nil {
			log.Printf("delete topic %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case subscribe != ``:
		sr := v20200217.NewDeleteCmqSubscribeRequest()
		sr.SubscriptionName = &subscribe
		sr.TopicName = &topic
		resp, err := mgrClient.DeleteCmqSubscribe(sr)
		if err != nil {
			log.Printf("delete subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
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
		fmt.Println(resp.ToJsonString())
	case topic != ``:
		tr := v20200217.NewModifyCmqTopicAttributeRequest()
		resp, err := mgrClient.ModifyCmqTopicAttribute(tr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case subscribe != ``:
		sr := v20200217.NewModifyCmqSubscriptionAttributeRequest()
		resp, err := mgrClient.ModifyCmqSubscriptionAttribute(sr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
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
		fmt.Println(detail.ToJsonString())
	case topic != ``:
		tr := v20200217.NewDescribeCmqTopicDetailRequest()
		tr.TopicName = &topic
		detail, err := mgrClient.DescribeCmqTopicDetail(tr)
		if err != nil {
			log.Printf("describe topic %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(detail.ToJsonString())
	case subscribe != ``:
		sr := v20200217.NewDescribeCmqSubscriptionDetailRequest()
		sr.TopicName = &topic
		sr.SubscriptionName = &subscribe
		detail, err := mgrClient.DescribeCmqSubscriptionDetail(sr)
		if err != nil {
			log.Printf("describe subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(detail.ToJsonString())
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
		fmt.Println(resp.ToJsonString())
	case `topic`:
		tr := v20200217.NewDescribeCmqTopicsRequest()
		resp, err := mgrClient.DescribeCmqTopics(tr)
		if err != nil {
			return
		}
		fmt.Println(resp.ToJsonString())
	case `subscribe`:
		sr := v20200217.NewDescribeSubscriptionsRequest()
		resp, err := mgrClient.DescribeSubscriptions(sr)
		if err != nil {
			return
		}
		fmt.Println(resp.ToJsonString())
	}
}
