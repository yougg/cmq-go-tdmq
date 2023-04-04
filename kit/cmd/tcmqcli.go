//go:generate go build -trimpath -buildmode pie -installsuffix netgo -tags "osusergo netgo static_build" -ldflags "-s -w" ${GOFILE}
//go:generate sh -c "[ -z \"${GOEXE}\" ] && gzip -S _${GOOS}_${GOARCH}.gz tcmqcli || zip -mjq tcmqcli_${GOOS}_${GOARCH}.zip tcmqcli${GOEXE}"
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/integrii/flaggy"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	v20200217 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
	tcmq "github.com/yougg/cmq-go-tdmq"
)

type (
	Flag struct {
		Var      any
		Name     string
		FullName string
		Value    any
		Usage    string
	}

	SubCmd struct {
		ShortName string
		Name      string
		Do        func()
		Desc      string
		Flags     []Flag
		subCmds   []SubCmd
	}

	Queue struct {
		v20200217.CreateCmqQueueRequest
	}
	Topic struct {
		RoutingKey string
		v20200217.CreateCmqTopicRequest
	}
	Subscribe struct {
		Tag []string
		v20200217.CreateCmqSubscribeRequest
	}
)

const (
	lanUrl = `http://%s.mqadapter.cmq.tencentyun.com`
	wanUrl = `https://cmq-%s.public.tencenttdmq.com`
)

var regions = map[string]string{
	"gz":      "ap-guangzhou",
	"sh":      "ap-shanghai",
	"hk":      "ap-hongkong",
	"ca":      "na-toronto",
	"shjr":    "ap-shanghai-fsi",
	"bj":      "ap-beijing",
	"sg":      "ap-singapore",
	"szjr":    "ap-shenzhen-fsi",
	"gzopen":  "ap-guangzhou-open",
	"usw":     "na-siliconvalley",
	"cd":      "ap-chengdu",
	"de":      "eu-frankfurt",
	"kr":      "ap-seoul",
	"cq":      "ap-chongqing",
	"in":      "ap-mumbai",
	"use":     "na-ashburn",
	"th":      "ap-bangkok",
	"ru":      "eu-moscow",
	"jp":      "ap-tokyo",
	"jnec":    "ap-jinan-ec",
	"hzec":    "ap-hangzhou-ec",
	"nj":      "ap-nanjing",
	"fzec":    "ap-fuzhou-ec",
	"whec":    "ap-wuhan-ec",
	"dub":     "me-dubai",
	"la":      "na-losangeles",
	"sl":      "sl-saopaulo",
	"syd":     "au-sydney",
	"tj":      "ap-beijing-z1",
	"tsn":     "ap-tianjin",
	"szx":     "ap-shenzhen",
	"tpe":     "ap-taipei",
	"bjjr":    "ap-beijing-fsi",
	"others":  "ap-others",
	"csec":    "ap-changsha-ec",
	"sjwec":   "ap-shijiazhuang-ec",
	"qy":      "ap-qingyuan",
	"xbec":    "ap-xibei-ec",
	"hfeec":   "ap-hefei-ec",
	"sheec":   "ap-shenyang-ec",
	"xiyec":   "ap-xian-ec",
	"cgoec":   "ap-zhengzhou-ec",
	"jkt":     "ap-jakarta",
	"sao":     "sa-saopaulo",
	"qyxa":    "ap-qingyuan-xinan",
	"szsycft": "ap-shenzhen-sycft",
	"gy":      "ap-guiyang",
	"shadc":   "ap-shanghai-adc",
	"szjrtce": "ap-shanghai-fsitce",
	"shjrtce": "ap-shenzhen-fsitce",
}

var (
	uri   string
	sid   string
	key   string
	token string

	action string

	repeat    = 1
	holdTime  int
	interval  int
	keepalive bool

	msgs   = make([]string, 0, 16)
	length int

	handles = make([]string, 0, 16)

	ack     bool
	timeout = 5
	number  = 1
	delay   int
	waits   = 3

	insecure bool
	debug    bool

	network  = `public`
	method   = `POST`
	region   string
	endpoint string

	limit  uint64 = 20
	offset uint64
	filter = `queue`
)

// set VALUE_SEPARATOR=... environment variable for value separator
// the default slice flag parameters value separator is ','
// customize the VALUE_SEPARATOR to avoid the slice value be separate by ','
var valueSeparator string

var (
	client    *tcmq.Client
	mgrClient *v20200217.Client
)

var (
	q = Queue{
		CreateCmqQueueRequest: v20200217.CreateCmqQueueRequest{
			QueueName:           new(string),
			MaxMsgHeapNum:       new(uint64),
			PollingWaitSeconds:  new(uint64),
			VisibilityTimeout:   new(uint64),
			MaxMsgSize:          new(uint64),
			MsgRetentionSeconds: new(uint64),
			RewindSeconds:       new(uint64),
			Transaction:         new(uint64),
			FirstQueryInterval:  new(uint64),
			MaxQueryCount:       new(uint64),
			DeadLetterQueueName: new(string),
			Policy:              new(uint64),
			MaxReceiveCount:     new(uint64),
			MaxTimeToLive:       new(uint64),
			Trace:               new(bool),
			RetentionSizeInMB:   new(uint64),
			Tags:                make([]*v20200217.Tag, 0),
		},
	}
	t = Topic{
		CreateCmqTopicRequest: v20200217.CreateCmqTopicRequest{
			TopicName:           new(string),
			MaxMsgSize:          new(uint64),
			FilterType:          new(uint64),
			MsgRetentionSeconds: new(uint64),
			Trace:               new(bool),
			Tags:                make([]*v20200217.Tag, 0),
		},
	}
	s = Subscribe{
		Tag: make([]string, 0, 5),
		CreateCmqSubscribeRequest: v20200217.CreateCmqSubscribeRequest{
			TopicName:           new(string),
			SubscriptionName:    new(string),
			Protocol:            new(string),
			Endpoint:            new(string),
			NotifyStrategy:      new(string),
			FilterTag:           make([]*string, 0),
			BindingKey:          make([]*string, 0),
			NotifyContentFormat: new(string),
		},
	}

	qFlag   = Flag{q.QueueName, `q`, `queue`, ``, `queue name`}
	tFlag   = Flag{t.TopicName, `t`, `topic`, ``, `topic name`}
	sFlag   = Flag{s.SubscriptionName, `s`, `subscribe`, ``, `subscribe name`}
	rFlag   = Flag{&t.RoutingKey, `r`, `routingKey`, ``, `routing key`}
	mFlag   = Flag{&msgs, `m`, `msg`, &msgs, `message(s), repeat the flag 2~16 times to set multi messages`}
	lFlag   = Flag{&length, `l`, `length`, 0, `send/publish message(s) with specified length, unit: byte`}
	tagFlag = Flag{&s.Tag, `g`, `tag`, &s.Tag, `tag(s), repeat the flag multi times to set multi tags`}
	hFlag   = Flag{&handles, `n`, `handle`, &handles, `handle(s), repeat the flag 2~16 times to set multi handles`}
	ackFlag = Flag{&ack, `c`, `ack`, false, `receive message(s) with ack`}
	nFlag   = Flag{&number, `n`, `number`, 1, `send/receive/publish <number> message(s)`}
	dFlag   = Flag{&delay, `y`, `delay`, 0, `send message(s) <delay> second`}
	wFlag   = Flag{&waits, `w`, `wait`, 5, `receive message(s) <wait> seconds`}
	fFlag   = Flag{&filter, `f`, `filter`, ``, `list filter resource type: queue/topic/subscribe/region`}
	limitF  = Flag{&limit, `l`, `limit`, 20, `limit query page size`}
	offsetF = Flag{&offset, `o`, `offset`, 0, `begin index of query page`}

	queueFlags = []Flag{
		{q.QueueName, `n`, `name`, ``, `queue name`},
		{q.MaxMsgHeapNum, ``, `MaxMsgHeapNum`, 1000000, `max message heap number [1000000~1000000000]`},
		{q.PollingWaitSeconds, ``, `PollingWaitSeconds`, 0, `polling wait seconds [0~30]`},
		{q.VisibilityTimeout, ``, `VisibilityTimeout`, 30, `visibility timeout in seconds [0~43200]`},
		{q.MaxMsgSize, ``, `MaxMsgSize`, 65536, `max message size [1024~65536]`},
		{q.MsgRetentionSeconds, ``, `MsgRetentionSeconds`, 3600, `message retention seconds [30~43200]`},
		{q.RewindSeconds, ``, `RewindSeconds`, 0, `rewind seconds [0~1296000]`},
		{q.Transaction, ``, `Transaction`, 0, `transaction, 0:disable, 1:enable`},
		{q.FirstQueryInterval, ``, `FirstQueryInterval`, 0, `first query interval`},
		{q.MaxQueryCount, ``, `MaxQueryCount`, 0, `max query count`},
		{q.DeadLetterQueueName, ``, `DeadLetterQueueName`, ``, `dead letter queue name`},
		{q.Policy, ``, `Policy`, 1, `dead letter policy, 0:not acked after consume many times, 1:TTL expired`},
		{q.MaxReceiveCount, ``, `MaxReceiveCount`, 1, `max receive count [1~1000]`},
		{q.MaxTimeToLive, ``, `MaxTimeToLive`, 300, `max time to live [300~43200]`},
		{q.Trace, ``, `Trace`, false, `trace message, true:enable, false:disable`},
		{q.RetentionSizeInMB, ``, `RetentionSizeInMB`, 0, `retention size in MB [10240~512000]`},
	}
	topicFlags = []Flag{
		{t.TopicName, `n`, `name`, ``, `topic name`},
		{t.MaxMsgSize, ``, `MaxMsgSize`, 65536, `max message size [1024~65536]`},
		{t.FilterType, ``, `FilterType`, 1, `subscribe message filter type, 1:tag, 2:route`},
		{t.MsgRetentionSeconds, ``, `MsgRetentionSeconds`, 86400, `message retention seconds [60~86400]`},
		{t.Trace, ``, `Trace`, false, `trace message, true:enable, false:disable`},
	}
	subscribeFlags = []Flag{
		{s.SubscriptionName, `n`, `name`, ``, `subscribe name`},
		{s.TopicName, ``, `TopicName`, ``, `topic name`},
		{s.Protocol, ``, `Protocol`, ``, `deliver protocol, [http,queue]`},
		{s.Endpoint, ``, `Endpoint`, ``, `endpoint of deliver protocol, http url or queue name`},
		{s.NotifyStrategy, ``, `NotifyStrategy`, `EXPONENTIAL_DECAY_RETRY`, `deliver notify strategy: BACKOFF_RETRY, EXPONENTIAL_DECAY_RETRY`},
		{s.FilterTag, ``, `FilterTag`, nil, `message filter tag, max 5 count and each one max 16 chars`},
		{s.BindingKey, ``, `BindingKey`, nil, `message binding key, max 5 count and each one max 64 chars`},
		{s.NotifyContentFormat, ``, `NotifyContentFormat`, `JSON`, `notify content format: JSON, SIMPLIFIED`},
	}

	actionCmds = []SubCmd{
		{ShortName: `s`, Name: `send`, Do: send, Desc: "send message(s) to queue", Flags: []Flag{qFlag, mFlag, lFlag, dFlag, nFlag}},
		{ShortName: `r`, Name: `receive`, Do: receive, Desc: "receive message(s) from queue", Flags: []Flag{qFlag, wFlag, nFlag, ackFlag}},
		{ShortName: `d`, Name: `delete`, Do: acknowledge, Desc: "delete message by handle(s)", Flags: []Flag{qFlag, hFlag}},
		{ShortName: `p`, Name: `publish`, Do: publish, Desc: "publish message(s) to topic", Flags: []Flag{tFlag, mFlag, nFlag, lFlag, rFlag, tagFlag}},
		{ShortName: `q`, Name: `query`, Do: query, Desc: "query topic/queue route for tcp", Flags: []Flag{qFlag, tFlag}},
		{Name: ` `},
		{ShortName: `c`, Name: `create`, Desc: "create queue / topic / subscribe", subCmds: []SubCmd{
			{ShortName: `q`, Name: `queue`, Do: createQ, Desc: "create queue", Flags: queueFlags},
			{ShortName: `t`, Name: `topic`, Do: createT, Desc: "create topic", Flags: topicFlags},
			{ShortName: `s`, Name: `subscribe`, Do: createS, Desc: "create subscribe", Flags: subscribeFlags},
		}},
		{ShortName: `e`, Name: `remove`, Do: remove, Desc: "remove queue / topic / subscribe", Flags: []Flag{sFlag, qFlag, tFlag}},
		{ShortName: `m`, Name: `modify`, Desc: "modify queue / topic / subscribe", subCmds: []SubCmd{
			{ShortName: `q`, Name: `queue`, Do: modifyQ, Desc: "modify queue", Flags: queueFlags},
			{ShortName: `t`, Name: `topic`, Do: modifyT, Desc: "modify topic", Flags: topicFlags},
			{ShortName: `s`, Name: `subscribe`, Do: modifyS, Desc: "modify subscribe", Flags: subscribeFlags},
		}},
		{ShortName: `i`, Name: `describe`, Do: describe, Desc: "describe queue / topic / subscribe", Flags: []Flag{sFlag, qFlag, tFlag, fFlag, limitF, offsetF}},
		{ShortName: `l`, Name: `list`, Do: lists, Desc: "list -f  <queue | topic | subscribe | region>", Flags: []Flag{sFlag, qFlag, tFlag, fFlag, limitF, offsetF}},
	}
)

func init() {
	if separator := os.Getenv(`VALUE_SEPARATOR`); separator != `` {
		valueSeparator = separator
	}
	flaggy.SetVersion(`v0.2.1`)
	flaggy.SetDescription(`TDMQ-CMQ command line tool`)
	flagFn := func(subCmd *flaggy.Subcommand, flags []Flag) {
		for _, f := range flags {
			// fmt.Printf("name: %#v, var: %#v, value: %#v\n", f.FullName, f.Var, f.Value)
			isNilValue := f.Value == nil || (*[2]uintptr)(unsafe.Pointer(&f.Value))[1] == 0
			switch v := f.Var.(type) {
			case *string:
				if isNilValue {
					f.Value = new(string)
				}
				*v = f.Value.(string)
				subCmd.String(v, f.Name, f.FullName, f.Usage)
			case *int:
				if isNilValue {
					f.Value = new(int)
				}
				*v = f.Value.(int)
				subCmd.Int(v, f.Name, f.FullName, f.Usage)
			case *uint64:
				if isNilValue {
					*v = uint64(0)
				} else {
					*v = uint64(f.Value.(int))
				}
				subCmd.UInt64(v, f.Name, f.FullName, f.Usage)
			case *bool:
				if isNilValue {
					f.Value = new(bool)
				}
				*v = f.Value.(bool)
				subCmd.Bool(v, f.Name, f.FullName, f.Usage)
			case []string:
				if isNilValue {
					f.Value = make([]string, 0)
				}
				v = f.Value.([]string)
				subCmd.StringSlice(&v, f.Name, f.FullName, f.Usage, valueSeparator)
			case *[]string:
				if isNilValue {
					f.Value = &[]string{}
				}
				v = f.Value.(*[]string)
				subCmd.StringSlice(v, f.Name, f.FullName, f.Usage, valueSeparator)
			case []*string:
				if isNilValue {
					f.Value = make([]*string, 0)
				}
				v = f.Value.([]*string)
				subCmd.Flags = append(subCmd.Flags, &flaggy.Flag{
					AssignmentVar:  &v,
					ShortName:      f.Name,
					LongName:       f.FullName,
					Description:    f.Usage,
					ValueSeparator: valueSeparator,
				})
			default:
				log.Printf("invalid flag type: %#v", f)
			}
		}
	}
	for _, cmd := range actionCmds {
		subCmd := flaggy.NewSubcommand(cmd.Name)
		subCmd.ShortName = cmd.ShortName
		subCmd.Description = cmd.Desc
		flagFn(subCmd, cmd.Flags)
		flaggy.AttachSubcommand(subCmd, 1)
		for _, c := range cmd.subCmds {
			cs := flaggy.NewSubcommand(c.Name)
			cs.ShortName = c.ShortName
			cs.Description = c.Desc
			flagFn(cs, c.Flags)
			subCmd.AttachSubcommand(cs, 1)
		}
	}

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], `-`) {
		action = os.Args[1]
	}
	flaggy.String(&method, "X", "method", "http request method: GET/POST")
	flaggy.String(&region, "r", "region", "public cloud region, ex: gz/sh/bj/usw/jp/de...")
	switch action {
	case `query`, `send`, `receive`, `delete`, `publish`, `q`, `s`, `r`, `d`, `p`:
		flaggy.Bool(&debug, "d", `debug`, "print debug log (default false)")
		flaggy.String(&uri, "u", "uri", "request uri for message action")
		flaggy.String(&network, "net", "network", "access from public or private network")
		flaggy.Bool(&keepalive, "", `keepalive`, "keepalive for client connection")
	case `create`, `remove`, `modify`, `describe`, `list`, `c`, `e`, `m`, `i`, `l`:
		flaggy.String(&endpoint, "e", "endpoint", "special endpoint for manage action (disable region)")
	}

	flaggy.Bool(&insecure, "k", `insecure`, "skip verifies servers certificate")
	flaggy.Int(&timeout, "", `timeout`, "client request timeout in seconds")
	flaggy.Int(&repeat, "", "repeat", "repeat request count in serial mode (message flow only)")
	flaggy.Int(&holdTime, "", "holdTime", "repeat request util hold seconds timeout (override request count)")
	flaggy.Int(&interval, "", "interval", "interval milliseconds between each repeat request")
	flaggy.String(&token, "", "token", "token for temporary secretId/secretKey")
	flaggy.String(&sid, "sid", "secretId", "secret id")
	flaggy.String(&key, "key", "secretKey", "secret key")

	// flaggy.DebugMode = true
	flaggy.Parse()
}

func main() {
	var err error
	switch action {
	case `query`, `send`, `receive`, `delete`, `publish`, `q`, `s`, `r`, `d`, `p`:
		if uri == `` && region != `` {
			if network != `public` {
				uri = fmt.Sprintf(lanUrl, region)
			} else {
				uri = fmt.Sprintf(wanUrl, region)
			}
		}
		tcmq.InsecureSkipVerify = insecure
		client, err = tcmq.NewClient(uri, sid, key, time.Duration(timeout)*time.Second, keepalive)
		if err != nil {
			log.Println("new TCMQ client", err)
			return
		}
		client.Method = method
		client.Token = token
		client.Debug = debug
	case `create`, `remove`, `modify`, `describe`, `list`, `c`, `e`, `m`, `i`, `l`:
		// 管控API文档: https://cloud.tencent.com/document/product/1496/62819
		prof := profile.NewClientProfile()
		prof.HttpProfile.ReqTimeout = timeout
		if endpoint != `` {
			prof.HttpProfile.Endpoint = endpoint
		}
		credential := common.NewTokenCredential(sid, key, token)
		mgrClient, err = v20200217.NewClient(credential, regions[region], prof)
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

	var chooseCmd = make(map[string]bool)
	for _, subCmd := range flaggy.DefaultParser.Subcommands {
		chooseCmd[subCmd.Name] = subCmd.Used
		chooseCmd[subCmd.ShortName] = subCmd.Used
		for _, sc := range subCmd.Subcommands {
			chooseCmd[subCmd.Name+`.`+sc.Name] = sc.Used
			chooseCmd[subCmd.Name+`.`+sc.ShortName] = sc.Used
			chooseCmd[subCmd.ShortName+`.`+sc.Name] = sc.Used
			chooseCmd[subCmd.ShortName+`.`+sc.ShortName] = sc.Used
		}
	}

	// fmt.Printf("%#v\n", chooseCmd)

	for _, cmd := range actionCmds {
		if cmd.Name != action && cmd.ShortName != action {
			continue
		}
		if len(cmd.subCmds) > 0 {
			for _, c := range cmd.subCmds {
				if chooseCmd[cmd.Name+`.`+c.Name] ||
					chooseCmd[cmd.Name+`.`+c.ShortName] ||
					chooseCmd[cmd.ShortName+`.`+c.Name] ||
					chooseCmd[cmd.ShortName+`.`+c.ShortName] {
					repeatDo(c.Do)
					break
				}
			}
		} else {
			repeatDo(cmd.Do)
		}
	}
}

func repeatDo(do func()) {
	if holdTime > 0 {
		now := time.Now()
		for time.Since(now) <= time.Duration(holdTime)*time.Second {
			do()
			if interval > 0 {
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}
	} else {
		for i := 0; i < repeat; i++ {
			do()
			if interval > 0 {
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}
	}
}

func query() {
	switch {
	case *q.QueueName != ``:
		r, err := client.QueryQueueRoute(*q.QueueName)
		if err != nil {
			log.Println("query queue route:", err)
			return
		}
		if !debug {
			fmt.Println(r)
		}
	case *t.TopicName != ``:
		r, err := client.QueryTopicRoute(*t.TopicName)
		if err != nil {
			log.Println("query topic route:", err)
			return
		}
		if !debug {
			fmt.Println(r)
		}
	default:
		log.Printf("invalid query parameters, queue: %s, topic: %s\n", *q.QueueName, *t.TopicName)
	}
}

func send() {
	switch {
	case len(msgs) > 0:
		for i := range msgs {
			if len(msgs[i]) == 0 {
				log.Println("message is empty")
				return
			}
		}
	case length > 0:
		msg := strings.Repeat(`#`, length)
		if length > tcmq.MaxMessageSize {
			tcmq.MaxMessageSize = length
		}
		if number > 1 {
			msgs = make([]string, 0, number)
			for i := 0; i < number; i++ {
				msgs = append(msgs, msg)
			}
		} else {
			msgs = []string{msg}
		}
	default:
		log.Println("no message to send, set message(s) via -m flag")
		return
	}

	queue := &tcmq.Queue{
		Client:       client,
		Name:         *q.QueueName,
		DelaySeconds: delay,
	}
	var resp tcmq.Result
	var err error
	var s string
	if len(msgs) == 1 {
		resp, err = queue.Send(msgs[0])
	} else {
		resp, err = queue.BatchSend(msgs...)
		s = "s"
	}
	if err != nil {
		log.Println("send message"+s+":", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func receive() {
	queue := &tcmq.Queue{
		Client:             client,
		Name:               *q.QueueName,
		PollingWaitSeconds: waits,
	}
	var resp tcmq.Result
	var err error
	var s string
	if number == 1 {
		resp, err = queue.Receive()
	} else {
		resp, err = queue.BatchReceive(number)
		s = "s"
	}
	if err != nil {
		log.Println("receive message"+s+":", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}

	if !ack || resp.Code() != 0 {
		return
	}
	handles = nil
	if res, ok := resp.(tcmq.ResponseRM); ok {
		if handle := res.Handle(); len(handle) > 0 {
			handles = append(handles, handle)
		}
	}
	if res, ok := resp.(tcmq.ResponseRMs); ok {
		for _, m := range res.MsgInfos() {
			if handle := m.Handle(); len(handle) > 0 {
				handles = append(handles, handle)
			}
		}
	}
	switch len(handles) {
	case 1:
		resp, err = queue.Delete(handles[0])
	default:
		resp, err = queue.BatchDelete(handles...)
		s = "s"
	}
	if err != nil {
		log.Println("delete message"+s+":", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func acknowledge() {
	queue := &tcmq.Queue{
		Client: client,
		Name:   *q.QueueName,
	}
	var resp tcmq.Result
	var err error
	var s string
	if len(handles) == 1 {
		resp, err = queue.Delete(handles[0])
	} else {
		resp, err = queue.BatchDelete(handles...)
		s = "s"
	}
	if err != nil {
		log.Println("delete message"+s+":", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func publish() {
	switch {
	case len(msgs) > 0:
		for i := range msgs {
			if len(msgs[i]) == 0 {
				log.Println("message is empty")
				return
			}
		}
	case length > 0:
		msg := strings.Repeat(`#`, length)
		if length > tcmq.MaxMessageSize {
			tcmq.MaxMessageSize = length
		}
		if number > 1 {
			msgs = make([]string, 0, number)
			for i := 0; i < number; i++ {
				msgs = append(msgs, msg)
			}
		} else {
			msgs = []string{msg}
		}
	default:
		log.Println("no message to publish, use -m to set message")
		return
	}
	topic := &tcmq.Topic{
		Client:     client,
		Name:       *t.TopicName,
		RoutingKey: t.RoutingKey,
		Tags:       s.Tag,
	}
	var resp tcmq.Result
	var err error
	var s string
	if len(msgs) == 1 {
		resp, err = topic.Publish(msgs[0])
	} else {
		resp, err = topic.BatchPublish(msgs...)
		s = "s"
	}
	if err != nil {
		log.Println("publish message"+s+":", err)
		return
	}
	if !debug {
		fmt.Println(resp)
	}
}

func createQ() {
	filter = `queue`
	create()
}

func createT() {
	filter = `topic`
	create()
}

func createS() {
	filter = `subscribe`
	create()
}

func create() {
	switch filter {
	case `subscribe`:
		sr := v20200217.NewCreateCmqSubscribeRequest()
		_ = sr.FromJsonString(s.ToJsonString())
		resp, err := mgrClient.CreateCmqSubscribe(sr)
		if err != nil {
			log.Printf("create subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case `queue`:
		qr := v20200217.NewCreateCmqQueueRequest()
		_ = qr.FromJsonString(q.ToJsonString())
		resp, err := mgrClient.CreateCmqQueue(qr)
		if err != nil {
			log.Printf("create queue %s error: %v", *qr.QueueName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case `topic`:
		tr := v20200217.NewCreateCmqTopicRequest()
		_ = tr.FromJsonString(t.ToJsonString())
		resp, err := mgrClient.CreateCmqTopic(tr)
		if err != nil {
			log.Printf("create topic %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	}
}

func remove() {
	switch {
	case *s.SubscriptionName != ``:
		filter = `subscribe`
	case *q.QueueName != ``:
		filter = `queue`
	case *t.TopicName != ``:
		filter = `topic`
	}
	switch filter {
	case `subscribe`:
		sr := v20200217.NewDeleteCmqSubscribeRequest()
		sr.SubscriptionName = s.SubscriptionName
		sr.TopicName = t.TopicName
		resp, err := mgrClient.DeleteCmqSubscribe(sr)
		if err != nil {
			log.Printf("delete subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case `queue`:
		qr := v20200217.NewDeleteCmqQueueRequest()
		qr.QueueName = q.QueueName
		resp, err := mgrClient.DeleteCmqQueue(qr)
		if err != nil {
			log.Printf("delete queue %s error: %v", *qr.QueueName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case `topic`:
		tr := v20200217.NewDeleteCmqTopicRequest()
		tr.TopicName = t.TopicName
		resp, err := mgrClient.DeleteCmqTopic(tr)
		if err != nil {
			log.Printf("delete topic %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	}
}

func modifyQ() {
	filter = `queue`
	modify()
}

func modifyT() {
	filter = `topic`
	modify()
}

func modifyS() {
	filter = `subscribe`
	modify()
}

func modify() {
	switch filter {
	case `subscribe`:
		sr := v20200217.NewModifyCmqSubscriptionAttributeRequest()
		_ = sr.FromJsonString(s.ToJsonString())
		resp, err := mgrClient.ModifyCmqSubscriptionAttribute(sr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case `queue`:
		qr := v20200217.NewModifyCmqQueueAttributeRequest()
		_ = qr.FromJsonString(q.ToJsonString())
		resp, err := mgrClient.ModifyCmqQueueAttribute(qr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *qr.QueueName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	case `topic`:
		tr := v20200217.NewModifyCmqTopicAttributeRequest()
		_ = tr.FromJsonString(t.ToJsonString())
		resp, err := mgrClient.ModifyCmqTopicAttribute(tr)
		if err != nil {
			log.Printf("modify queue %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(resp.ToJsonString())
	}
}

func describe() {
	if filter == `` {
		switch {
		case *s.SubscriptionName != ``:
			filter = `subscribe`
		case *q.QueueName != ``:
			filter = `queue`
		case *t.TopicName != ``:
			filter = `topic`
		}
	}
	switch filter {
	case `subscribe`:
		sr := v20200217.NewDescribeCmqSubscriptionDetailRequest()
		sr.TopicName = t.TopicName
		sr.SubscriptionName = s.SubscriptionName
		sr.Limit = &limit
		sr.Offset = &offset
		detail, err := mgrClient.DescribeCmqSubscriptionDetail(sr)
		if err != nil {
			log.Printf("describe subscribe %s error: %v", *sr.SubscriptionName, err)
			return
		}
		fmt.Println(detail.ToJsonString())
	case `queue`:
		qr := v20200217.NewDescribeCmqQueueDetailRequest()
		qr.QueueName = q.QueueName
		detail, err := mgrClient.DescribeCmqQueueDetail(qr)
		if err != nil {
			log.Printf("describe queue %s error: %v", *qr.QueueName, err)
			return
		}
		fmt.Println(detail.ToJsonString())
	case `topic`:
		tr := v20200217.NewDescribeCmqTopicDetailRequest()
		tr.TopicName = t.TopicName
		detail, err := mgrClient.DescribeCmqTopicDetail(tr)
		if err != nil {
			log.Printf("describe topic %s error: %v", *tr.TopicName, err)
			return
		}
		fmt.Println(detail.ToJsonString())
	}
}

func lists() {
	if filter == `` {
		switch {
		case *s.SubscriptionName != ``:
			filter = `subscribe`
		case *q.QueueName != ``:
			filter = `queue`
		case *t.TopicName != ``:
			filter = `topic`
		case region != ``:
			filter = `region`
		}
	}
	switch filter {
	case `subscribe`:
		sr := v20200217.NewDescribeSubscriptionsRequest()
		sr.TopicName = t.TopicName
		sr.SubscriptionName = s.SubscriptionName
		sr.Limit = &limit
		sr.Offset = &offset
		resp, err := mgrClient.DescribeSubscriptions(sr)
		if err != nil {
			return
		}
		fmt.Println(resp.ToJsonString())
	case `queue`:
		qr := v20200217.NewDescribeCmqQueuesRequest()
		qr.QueueName = q.QueueName
		// qr.QueueNameList =   // TODO
		qr.Limit = &limit
		qr.Offset = &offset
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
	case `region`:
		var keys []string
		wg := &sync.WaitGroup{}
		status := &sync.Map{}
		http.DefaultClient.Timeout = time.Duration(timeout) * time.Second
		check := func(u string) {
			var available bool
			resp, err := http.Get(u)
			if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
				available = true
			}
			status.Store(u, available)
			wg.Done()
		}
		state := func(k string) (wanState, lanState string) {
			var wan, lan bool
			if v, ok := status.Load(fmt.Sprintf(wanUrl, k)); ok {
				wan = v.(bool)
			}
			if v, ok := status.Load(fmt.Sprintf(lanUrl, k)); ok {
				lan = v.(bool)
			}
			wanState, lanState = `-`, `-`
			if wan {
				wanState = `Arrive`
			}
			if lan {
				lanState = `Arrive`
			}
			return
		}
		format := "%7s  %-18s %-6s %-7s\n"
		fmt.Printf(format, `Region`, `AP Code`, `Public`, `Private`)
		if region != `` {
			regions = map[string]string{
				region: regions[region],
			}
		}
		wg.Add(len(regions) * 2)
		for k := range regions {
			go check(fmt.Sprintf(wanUrl, k))
			go check(fmt.Sprintf(lanUrl, k))
			keys = append(keys, k)
		}
		wg.Wait()
		sort.Strings(keys)
		for _, k := range keys {
			wanState, lanState := state(k)
			fmt.Printf(format, k, regions[k], wanState, lanState)
		}
	}
}
