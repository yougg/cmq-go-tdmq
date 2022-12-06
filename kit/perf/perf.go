//go:generate go build -trimpath -buildmode pie -installsuffix netgo -tags "osusergo netgo static_build" -ldflags "-s -w" ${GOFILE}
//go:generate sh -c "[ -z \"${GOEXE}\" ] && gzip -S _${GOOS}_${GOARCH}.gz perf || zip -mjq perf_${GOOS}_${GOARCH}.zip perf${GOEXE}"
package main

import (
	"container/ring"
	"encoding/json"
	"errors"
	"flag"
	"hash/crc32"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	dbg "runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/jamiealquiza/tachymeter"
	"github.com/pbnjay/memory"
	tcmq "github.com/yougg/cmq-go-tdmq"
	"github.com/yougg/pool"
)

type Case struct {
	Description        string           `json:"Description,omitempty"`        // 用例描述：执行1000次 向1个队列发送1条1KB的消息
	Enabled            bool             `json:"CaseEnabled,omitempty"`        // 是否启用本用例：true, false
	RepeatTimes        int              `json:"RepeatTimes,omitempty"`        // 用例重复次数
	RepeatTimeout      int              `json:"RepeatTimeout,omitempty"`      // 用例固定重复执行时间, 单位:秒, 非0时 RepeatTimes 配置无效
	Concurrent         int              `json:"Concurrent,omitempty"`         // 最大并发数量
	MaximumTPS         int              `json:"MaximumTPS,omitempty"`         // 最大限制TPS：0: 不限制，非0: 限制对应数量TPS
	ResourceType       string           `json:"ResourceType,omitempty"`       // 请求的资源类型：queue, topic
	ResourceName       string           `json:"ResourceName,omitempty"`       // 请求的资源名称：队列／主题的全名或者前缀，关联下面资源数量(1条时使用全名，多条时使用前缀)
	ResourceCount      int              `json:"ResourceCount,omitempty"`      // 请求的资源数量：1个或多个队列／主题
	ResourceStartIdx   int              `json:"ResourceStartIdx,omitempty"`   // 请求的资源列表开始索引
	RandMsgSize        bool             `json:"RandMsgSize,omitempty"`        // 请求的消息体积使用[1 ~ MessageSize]范围内的随机大小
	MessageSize        int              `json:"MessageSize,omitempty"`        // 请求的消息体积：1024B == 1KB，单条消息的体积，批量请求时总体积不能超过1MB
	MessageCount       int              `json:"MessageCount,omitempty"`       // 请求的消息数量：1条，每次Action请求消息数量，Batch批量Action请求为1~16条
	Action             string           `json:"Action,omitempty"`             // 请求消息的动作：QueryQueueRoute,SendMessage,BatchSendMessage,ReceiveMessage,BatchReceiveMessage,DeleteMessage,BatchDeleteMessage,QueryTopicRoute,PublishMessage,BatchPublishMessage
	AloneRecvTime      bool             `json:"AloneRecvTime,omitempty"`      // 拉取消息是否分隔Ack进行独立计时：true, false
	AckEnabled         bool             `json:"AckEnabled,omitempty"`         // 拉取到消息后是否向服务端Ack确认(删除)该条消息
	ReceiptHandles     []string         `json:"ReceiptHandles,omitempty"`     // 请求删除消息ID列表
	DelaySeconds       int              `json:"DelaySeconds,omitempty"`       // 单位为秒，消息发送到队列后，延时多久用户才可见该消息。
	PollingWaitSeconds int              `json:"PollingWaitSeconds,omitempty"` // 长轮询等待时间。取值范围0 - 30秒
	RoutingKey         string           `json:"RoutingKey,omitempty"`         // 发送消息的路由路径
	Tags               []string         `json:"Tags,omitempty"`               // 消息过滤标签
	Statistics         []*Statistics    `json:"-"`                            // 请求耗时与结果统计
	TPSes              sort.IntSlice    `json:"-"`                            // 每秒完成事务统计
	StatsChan          chan *Statistics `json:"-"`                            // 请求耗时与结果传递管道
}

type Statistics struct {
	CostTime time.Duration
	Succeed  bool
}

type list []string

func (l *list) String() string {
	return strings.Join(*l, ",")
}

func (l *list) Set(s string) error {
	*l = append(*l, s)
	return nil
}

const randMsgSize = 2 * 1024 * 1024 // 2M

var (
	addr string

	uris    list
	headers list
	sid     string
	key     string

	timeout   int
	keepalive bool

	insecure bool
	debug    bool
	showErr  bool
	showTPS  int
	succOnly bool

	clients []*tcmq.Client // multiple clients for load balance by consistent hash
	count   int            // created clients count

	kase  string
	cases []*Case

	randMsg string

	tps, tps1, tps2 = &atomic.Uint32{}, &atomic.Uint32{}, &atomic.Uint32{}
)

func init() {
	flag.StringVar(&addr, "http", "", "pprof listen address for perf tool, ex: 0.0.0.0:6666")
	flag.Var(&uris, "u", "URI(s), repeat '-u' multi times to set multi URIs")
	flag.Var(&headers, "H", "headers, repeat '-H' multi times to set multi headers")
	flag.StringVar(&sid, "i", "", "secret id")
	flag.StringVar(&key, "k", "", "secret key")
	flag.StringVar(&kase, "c", "cases.json", "test case file")
	flag.BoolVar(&keepalive, "keepalive", false, "keepalive connections from client server (default false)")
	flag.BoolVar(&insecure, "insecure", false, "whether client skip verifies server's certificate (default false)")
	flag.BoolVar(&debug, "d", false, "weather show client debug info (default false)")
	flag.BoolVar(&showErr, "e", false, "weather show error response (default false)")
	flag.IntVar(&timeout, "t", 30, "client timeout in seconds")
	flag.IntVar(&showTPS, "s", 0, "show current TPS every (s) seconds")
	flag.BoolVar(&succOnly, "succOnly", false, "only calculate cost time of succeed request (default false)")
	flag.Parse()
}

func init() {
	if kase == `` {
		return
	}

	data, err := os.ReadFile(kase)
	if err != nil {
		log.Println("read case file", err)
		return
	}
	err = json.Unmarshal(data, &cases)
	if err != nil {
		log.Println("unmarshal case file", err)
		return
	}

	if len(cases) == 0 {
		log.Println("no case found in case file", kase)
		return
	}

	// generate random 2MB length string
	// all visible ascii characters
	// !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~
	m := make([]byte, randMsgSize)
	for i := 0; i < randMsgSize; i++ {
		m[i] = byte('!' + rand.Intn('~'-'!'))
	}
	randMsg = string(m)
}

func main() {
	if len(uris) == 0 || sid == `` || key == `` {
		log.Printf("invalid parameters uris: %v, sid:%s, key:%s \n", uris, sid, key)
		return
	}

	// client load balance (multiple servers)
	for _, u := range uris {
		tcmq.InsecureSkipVerify = insecure
		client, err := tcmq.NewClient(u, sid, key, time.Duration(timeout)*time.Second, keepalive)
		if err != nil {
			log.Println("new TDMQ-CMQ client", err)
			return
		}
		client.Debug = debug
		if len(headers) > 0 {
			client.Header = map[string]string{}
			for _, h := range headers {
				kv := strings.Split(h, ":")
				if len(kv) != 2 {
					continue
				}
				client.Header[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
		clients = append(clients, client)
	}
	if count = len(clients); count == 0 {
		log.Println("no client created")
		return
	}

	go func() {
		if addr == `` {
			return
		}
		// go tool pprof -http=:9999 'http://127.0.0.1:6666/debug/pprof/heap'
		// go tool pprof -http=:9999 'http://127.0.0.1:6666/debug/pprof/profile?seconds=30'
		// go tool pprof -http=:9999 'http://127.0.0.1:6666/debug/pprof/block'
		// go tool pprof -http=:9999 'http://127.0.0.1:6666/debug/pprof/mutex'
		// wget -O trace.out 'http://127.0.0.1:6666/debug/pprof/trace?seconds=5'
		//  go tool trace 'trace.out'
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Println(err)
		}
	}()

	mem := memory.TotalMemory() * 6 / 10
	dbg.SetMemoryLimit(int64(mem))

	for _, c := range cases {
		if !c.Enabled {
			log.Println("Skip Case:", c.Description)
			continue
		}
		switch {
		case c.RepeatTimes < 0 && c.RepeatTimeout < 0:
			log.Printf("invalid case repeat times: %d, repeat timeout: %ds\n", c.RepeatTimes, c.RepeatTimeout)
			continue
		case c.Concurrent < 1:
			log.Printf("invalid case concurrent: %d\n", c.Concurrent)
			continue
		case c.ResourceCount < 1:
			log.Printf("invalid case resource count: %d\n", c.ResourceCount)
			continue
		case c.MessageSize < 0:
			log.Printf("invalid case message size: %d\n", c.MessageSize)
			continue
		case c.MessageCount < 0:
			log.Printf("invalid case message count: %d\n", c.MessageCount)
			continue
		}
		switch c.Action {
		case "QueryQueueRoute",
			"SendMessage",
			"BatchSendMessage",
			"ReceiveMessage",
			"BatchReceiveMessage",
			"DeleteMessage",
			"BatchDeleteMessage",
			"QueryTopicRoute",
			"PublishMessage",
			"BatchPublishMessage":
		default:
			log.Println("invalid action in case", c.Action)
			continue
		}
		log.Println("Test Case:", c.Description)
		test(c)
		l := len(c.Statistics)
		if l == 0 {
			log.Println("no statistics found in case result")
			continue
		}
		t := tachymeter.New(&tachymeter.Config{Size: l})
		var successes int
		for _, s := range c.Statistics {
			if succOnly {
				if s.Succeed {
					t.AddTime(s.CostTime)
				}
			} else {
				t.AddTime(s.CostTime)
			}
			if s.Succeed {
				successes++
			}
		}
		total := len(c.Statistics)
		rate := (float64(successes) / float64(total)) * 100
		log.Printf("Total Requests: %d, Succeed: %d, Rate: %.2f%%\n", total, successes, rate)
		// log.Printf("总请求数: %d, 成功: %d, 成功率: %.2f%%\n", total, successes, rate)
		if c.TPSes.Len() > 0 {
			c.TPSes.Sort()
			var sum int
			for _, v := range c.TPSes {
				sum += v
			}
			l = c.TPSes.Len()
			log.Printf("Max TPS: %d, Min TPS: %d, Avg TPS: %d, Median: %d\n", c.TPSes[l-1], c.TPSes[0], sum/l, c.TPSes[l/2])
			// log.Printf("最大TPS: %d, 最小TPS: %d, 平均TPS: %d, 中位数: %d\n", c.TPSes[l-1], c.TPSes[0], sum/l, c.TPSes[l/2])
		} else {
			log.Println("no TPS statistics")
		}
		log.Println("\n", t.Calc())
	}
}

func test(c *Case) {
	s, err := pool.NewBlocking(c.Concurrent)
	if err != nil {
		log.Println("create pool", err)
		return
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	var tpsShower *time.Ticker
	if showTPS > 0 {
		tpsShower = time.NewTicker(time.Duration(showTPS) * time.Second)
	}
	defer func() {
		// 等待2秒 便于TPS统计完成
		time.Sleep(2 * time.Second)
		ticker.Stop()
		if showTPS > 0 && tpsShower != nil {
			tpsShower.Stop()
		}
	}()
	r, r1, r2 := ring.New(100), ring.New(100), ring.New(100)
	for i := 0; i < r.Len(); i++ {
		r.Value, r1.Value, r2.Value = uint32(0), uint32(0), uint32(0)
		r, r1, r2 = r.Next(), r1.Next(), r2.Next()
	}
	var sumTPS = func() (sum int) {
		r.Do(func(i any) {
			v := int(i.(uint32))
			sum += v
		})
		return sum
	}
	var sumTPS12 = func() (sum1, sum2 int) {
		r1.Do(func(i any) {
			v := int(i.(uint32))
			sum1 += v
		})
		r2.Do(func(i any) {
			v := int(i.(uint32))
			sum2 += v
		})
		return sum1, sum2
	}
	go func() {
		for range ticker.C {
			// 每10毫秒统计一次当前TPS，统计后清零
			n, n1, n2 := tps.Swap(0), tps1.Swap(0), tps2.Swap(0)
			// 10ms/次 统计100次(1秒内) 的TPS
			r.Value, r1.Value, r2.Value = n, n1, n2
			r, r1, r2 = r.Next(), r1.Next(), r2.Next()
			c.TPSes = append(c.TPSes, sumTPS())
		}
	}()
	go func() {
		if showTPS > 0 && tpsShower != nil {
			for range tpsShower.C {
				sum1, sum2 := sumTPS12()
				total := sum1 + sum2
				rate := (float64(sum1) / float64(total)) * 100
				log.Printf("TPS: %-5d  Succeed: %-5d  Failed: %-5d  Rate: %.2f%%\n", total, sum1, sum2, rate)
			}
		}
	}()
	wg := &sync.WaitGroup{}
	var capacity int
	if c.RepeatTimeout > 0 {
		capacity = c.MaximumTPS * c.RepeatTimeout * c.ResourceCount
	} else {
		capacity = c.RepeatTimes * c.ResourceCount
	}
	if size := 100 * 1024 * 1024; capacity > size {
		capacity = size
	}
	c.Statistics = make([]*Statistics, 0, capacity)
	c.StatsChan = make(chan *Statistics, 1000)
	go func() {
		for s := range c.StatsChan {
			c.Statistics = append(c.Statistics, s)
		}
	}()
	ch := make(chan struct{}, 1)
	go func() {
		if c.RepeatTimeout > 0 {
			after := time.After(time.Duration(c.RepeatTimeout) * time.Second)
			for {
				select {
				case <-after:
					close(ch)
					return
				default:
					ch <- struct{}{}
				}
			}
		} else {
			for i := 0; i < c.RepeatTimes; i++ {
				ch <- struct{}{}
			}
			close(ch)
		}
	}()

	for range ch {
		wg.Add(c.ResourceCount)
		for j := 0; j < c.ResourceCount; j++ {
			if c.MaximumTPS > 0 {
				// 如果当前TPS超过了设定限制的最大TPS则等待直到TPS下降再继续
				for sumTPS() > c.MaximumTPS {
					time.Sleep(time.Millisecond)
				}
			}
			job := func(cc *Case) pool.Task {
				return func(args ...any) {
					var succeed bool
					defer func() {
						// 当前事务完成后增加TPS
						tps.Add(1)
						if succeed {
							tps1.Add(1)
						} else {
							tps2.Add(1)
						}
						wg.Done()
					}()
					name := cc.ResourceName
					if cc.ResourceCount > 1 {
						n := args[0].(int)
						n += cc.ResourceStartIdx
						name += strconv.Itoa(n)
					}
					begin := time.Now()
					k := cc.Action + name + begin.String()
					var client *tcmq.Client
					if count > 1 {
						client = clients[hash(k, count)]
					} else {
						client = clients[0]
					}
					var resp tcmq.Result
					var end *time.Time
					switch cc.Action {
					case "QueryQueueRoute":
						resp, err = client.QueryQueueRoute(name)
					case "SendMessage":
						head := rand.Intn(randMsgSize - cc.MessageSize)
						var tail int
						if c.RandMsgSize {
							// 指定范围内的随机消息体积
							tail = head + rand.Intn(cc.MessageSize)
							if tail == head {
								tail++
							}
						} else {
							tail = head + cc.MessageSize
						}
						msg := randMsg[head:tail]
						resp, err = client.SendMessage(name, msg, cc.DelaySeconds)
					case "BatchSendMessage":
						var msgs []string
						for x := 0; x < cc.MessageCount; x++ {
							head := rand.Intn(randMsgSize - cc.MessageSize)
							var tail int
							if c.RandMsgSize {
								tail = head + rand.Intn(cc.MessageSize)
								if tail == head {
									tail++
								}
							} else {
								tail = head + cc.MessageSize
							}
							msgs = append(msgs, randMsg[head:tail])
						}
						resp, err = client.BatchSendMessage(name, msgs, cc.DelaySeconds)
					case "ReceiveMessage":
						var res tcmq.ResponseRM
						res, err = client.ReceiveMessage(name, cc.PollingWaitSeconds)
						if c.AloneRecvTime {
							now := time.Now()
							end = &now
						}
						if err == nil && res.Code() == 0 && c.AckEnabled {
							resp, err = client.DeleteMessage(name, res.Handle())
						} else {
							resp = res
						}
					case "BatchReceiveMessage":
						var res tcmq.ResponseRMs
						res, err = client.BatchReceiveMessage(name, cc.PollingWaitSeconds, cc.MessageCount)
						if c.AloneRecvTime {
							now := time.Now()
							end = &now
						}
						if err == nil && res.Code() == 0 && c.AckEnabled {
							var handles []string
							for _, msg := range res.MsgInfos() {
								handles = append(handles, msg.Handle())
							}
							if len(handles) > 0 {
								resp, err = client.BatchDeleteMessage(name, handles)
							} else {
								err = errors.New("no message handles received")
							}
						} else {
							resp = res
						}
					case "DeleteMessage":
						var handle string
						if len(cc.ReceiptHandles) > 0 {
							handle = cc.ReceiptHandles[0]
						} else {
							err = errors.New("no receipt handle for delete")
							break
						}
						resp, err = client.DeleteMessage(name, handle)
					case "BatchDeleteMessage":
						resp, err = client.BatchDeleteMessage(name, cc.ReceiptHandles)
					case "QueryTopicRoute":
						resp, err = client.QueryTopicRoute(name)
					case "PublishMessage":
						head := rand.Intn(randMsgSize - cc.MessageSize)
						var tail int
						if c.RandMsgSize {
							tail = head + rand.Intn(cc.MessageSize)
							if tail == head {
								tail++
							}
						} else {
							tail = head + cc.MessageSize
						}
						msg := randMsg[head:tail]
						resp, err = client.PublishMessage(name, msg, cc.RoutingKey, cc.Tags)
					case "BatchPublishMessage":
						var msgs []string
						for x := 0; x < cc.MessageCount; x++ {
							head := rand.Intn(randMsgSize - cc.MessageSize)
							var tail int
							if c.RandMsgSize {
								tail = head + rand.Intn(cc.MessageSize)
								if tail == head {
									tail++
								}
							} else {
								tail = head + cc.MessageSize
							}
							msgs = append(msgs, randMsg[head:tail])
						}
						resp, err = client.BatchPublishMessage(name, cc.RoutingKey, msgs, cc.Tags)
					default:
						log.Println("invalid action in case", cc.Action)
						return
					}
					if err != nil {
						succeed = false
						if showErr {
							log.Println(client.Url, err)
						}
					} else {
						switch {
						case resp == nil || (*(*[2]uintptr)(unsafe.Pointer(&resp)))[1] == 0:
							succeed = false
							if showErr {
								log.Println(client.Url, resp, err)
							}
						case resp.Code() == 0:
							succeed = true
						default:
							succeed = false
							if showErr {
								log.Println(client.Url, resp)
							}
						}
					}
					var cost time.Duration
					if end == nil {
						cost = time.Since(begin)
					} else {
						cost = end.Sub(begin)
					}
					cc.StatsChan <- &Statistics{
						CostTime: cost,
						Succeed:  succeed,
					}
				}
			}
			if c.Concurrent > 1 {
				s.Join(job(c), j)
			} else {
				job(c)(j)
			}
		}
	}
	wg.Wait()
	close(c.StatsChan)
}

// hash consistent hash and mod to get client index
//
//	input: data string
//	input: length int
//	return: int
func hash(data string, length int) int {
	return int(crc32.ChecksumIEEE([]byte(data))) % length
}
