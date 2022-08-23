# TDMQ-CMQ Command Line

## 安装/编译

### 下载预编译的二进制安装

https://github.com/yougg/cmq-go-tdmq/releases

<details>
  <summary>从源码安装</summary>

### 安装二进制到`$GOPATH/bin`

```bash
go install github.com/yougg/cmq-go-tdmq/kit/cmd@main
```

### 从源码编译二进制

```bash
git clone https://github.com/yougg/cmq-go-tdmq.git
cd cmq-go-tdmq/kit/cmd/
GOOS=linux go generate tdmqcli.go
GOOS=darwin go generate tdmqcli.go
GOOS=windows go generate tdmqcli.go
```

> 编译输出文件: `tdmqcli` 或 `tdmqcli.exe`

</details>

## 查看帮助

### `tcmqcli -h` 或 `tcmqcli --help`

```bash
tcmqcli - TDMQ-CMQ command line tool

  Usage:
    tcmqcli [send|receive|delete|publish|query| |create|remove|modify|describe|list]

  Subcommands: 
    send (s)       send message(s) to queue
    receive (r)    receive message(s) from queue
    delete (d)     delete message by handle(s)
    publish (p)    publish message(s) to topic
    query (q)      query topic/queue route for tcp
     
    create (c)     create queue/topic/subscribe
    remove (e)     remove queue/topic/subscribe
    modify (m)     modify queue/topic/subscribe
    describe (i)   describe queue/topic/subscribe
    list (l)       list queue/topic/subscribe/region

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -k --insecure    whether client skip verifies server's certificate
    -t --timeout     client timeout in seconds (default: 5)
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli s -h` 或 `tcmqcli send -h`

```bash
send - send message(s) to queue

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -q --queue       queue name
    -m --msg         message(s), repeat the flag 2~16 times to set multi messages
    -l --length      send/publish message(s) with specified length, unit: byte (default: 0)
    -y --delay       send message(s) <delay> second (default: 0)
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -d --debug       print debug log (default false)
    -u --uri         request uri for message action
    -net --network     access from public or private network (default: public)
    -k --insecure    whether client skip verifies server's certificate
    -t --timeout     client timeout in seconds (default: 5)
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli r -h` 或 `tcmqcli receive -h`

```bash
receive - receive message(s) from queue

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -q --queue       queue name
    -w --wait        receive message(s) <wait> seconds (default: 5)
    -n --number      send/receive/publish <number> message(s) with special <length> (default: 1)
    -c --ack         receive message(s) with ack
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -d --debug       print debug log (default false)
    -u --uri         request uri for message action
    -k --insecure    whether client skip verifies server's certificate
    -t --timeout     client timeout in seconds (default: 5)
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli d -h` 或 `tcmqcli delete -h`

```bash
delete - delete message by handle(s)

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -q --queue       queue name
    -n --handle      handle(s), repeat the flag 2~16 times to set multi handles
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -d --debug       print debug log (default false)
    -u --uri         request uri for message action
    -k --insecure    whether client skip verifies server's certificate
    -t --timeout     client timeout in seconds (default: 5)
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli p -h` 或 `tcmqcli publish -h`

```bash
publish - publish message(s) to topic

  Flags: 
       --version      Displays the program version string.
    -h --help         Displays help with available flag, subcommand, and positional value parameters.
    -t --topic        topic name
    -m --msg          message(s), repeat the flag 2~16 times to set multi messages
    -n --number       send/receive/publish <number> message(s) with special <length> (default: 1)
    -l --length       send/publish message(s) with specified length, unit: byte (default: 0)
    -r --routingKey   routing key
    -g --tag          tag(s), repeat the flag multi times to set multi tags
    -d --debug        print debug log (default false)
    -u --uri          request uri for message action
    -k --insecure     whether client skip verifies server's certificate
    -sid --secretId     secret id
    -key --secretKey    secret key
```

#### `tcmqcli q -h` 或 `tcmqcli query -h`

```bash
query - query topic/queue route for tcp

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -q --queue       queue name
    -t --topic       topic name
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -d --debug       print debug log (default false)
    -u --uri         request uri for message action
    -net --network     access from public or private network (default: public)
    -k --insecure    whether client skip verifies server's certificate
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli c -h` 或 `tcmqcli create -h`

```bash
create - create queue/topic/subscribe

  Usage:
    create [queue|topic|subscribe]

  Subcommands: 
    queue (q)       create queue
    topic (t)       create topic
    subscribe (s)   create subscribe

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint    special endpoint for manage action (disable region)
    -k --insecure    whether client skip verifies server's certificate
    -t --timeout     client timeout in seconds (default: 5)
    -sid --secretId    secret id
    -key --secretKey   secret key
```

##### `tcmqcli create queue -h`

```bash
queue - create queue

  Flags: 
       --version               Displays the program version string.
    -h --help                  Displays help with available flag, subcommand, and positional value parameters.
    -n --name                  queue name
       --MaxMsgHeapNum         max message heap number [1000000~1000000000] (default: 1000000)
       --PollingWaitSeconds    polling wait seconds [0~30] (default: 0)
       --VisibilityTimeout     visibility timeout in seconds [0~43200] (default: 30)
       --MaxMsgSize            max message size [1024~65536] (default: 65536)
       --MsgRetentionSeconds   message retention seconds [30~43200] (default: 3600)
       --RewindSeconds         rewind seconds [0~1296000] (default: 0)
       --Transaction           transaction, 0:disable, 1:enable (default: 0)
       --FirstQueryInterval    first query interval (default: 0)
       --MaxQueryCount         max query count (default: 0)
       --DeadLetterQueueName   dead letter queue name
       --Policy                dead letter policy, 0:not acked after consume many times, 1:TTL expired (default: 1)
       --MaxReceiveCount       max receive count [1~1000] (default: 1)
       --MaxTimeToLive         max time to live [300~43200] (default: 300)
       --Trace                 trace message, true:enable, false:disable
       --RetentionSizeInMB     retention size in MB [10240~512000] (default: 0)
    -r --region                public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint              special endpoint for manage action (disable region)
    -k --insecure              whether client skip verifies server's certificate
    -t --timeout               client timeout in seconds (default: 5)
    -sid --secretId              secret id
    -key --secretKey             secret key
```

##### `tcmqcli create topic -h`

```bash
topic - create topic

  Flags: 
       --version               Displays the program version string.
    -h --help                  Displays help with available flag, subcommand, and positional value parameters.
    -n --name                  topic name
       --MaxMsgSize            max message size [1024~65536] (default: 65536)
       --FilterType            subscribe message filter type, 1:tag, 2:route (default: 1)
       --MsgRetentionSeconds   message retention seconds [60~86400] (default: 86400)
       --Trace                 trace message, true:enable, false:disable
    -r --region                public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint              special endpoint for manage action (disable region)
    -k --insecure              whether client skip verifies server's certificate
    -t --timeout               client timeout in seconds (default: 5)
    -sid --secretId              secret id
    -key --secretKey             secret key
```

##### `tcmqcli create subscribe -h`

```bash
subscribe - create subscribe

  Flags: 
       --version               Displays the program version string.
    -h --help                  Displays help with available flag, subcommand, and positional value parameters.
    -n --name                  subscribe name
       --TopicName             topic name
       --Protocol              deliver protocol, [http,queue]
       --Endpoint              endpoint of deliver protocol, http url or queue name
       --NotifyStrategy        deliver notify strategy, 1:BACKOFF_RETRY, 2:EXPONENTIAL_DECAY_RETRY (default: 2)
       --FilterTag             message filter tag, max 5 count and each one max 16 chars
       --BindingKey            message binding key, max 5 count and each one max 64 chars
       --NotifyContentFormat   notify content format, 1:JSON, 2:SIMPLIFIED (default: 2)
    -r --region                public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint              special endpoint for manage action (disable region)
    -k --insecure              whether client skip verifies server's certificate
    -t --timeout               client timeout in seconds (default: 5)
    -sid --secretId              secret id
    -key --secretKey             secret key
```

#### `tcmqcli e -h` 或 `tcmqcli remove -h`

```bash
remove - remove queue/topic/subscribe

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -s --subscribe   subscribe name
    -q --queue       queue name
    -t --topic       topic name
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint    special endpoint for manage action (disable region)
    -k --insecure    whether client skip verifies server's certificate
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli m -h` 或 `tcmqcli modify -h`

```bash
modify - modify queue/topic/subscribe

  Usage:
    modify [queue|topic|subscribe]

  Subcommands: 
    queue (q)       modify queue
    topic (t)       modify topic
    subscribe (s)   modify subscribe

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint    special endpoint for manage action (disable region)
    -k --insecure    whether client skip verifies server's certificate
    -t --timeout     client timeout in seconds (default: 5)
    -sid --secretId    secret id
    -key --secretKey   secret key
```

##### `tcmqcli modify queue -h`

```bash
queue - modify queue

  Flags: 
       --version               Displays the program version string.
    -h --help                  Displays help with available flag, subcommand, and positional value parameters.
    -n --name                  queue name
       --MaxMsgHeapNum         max message heap number [1000000~1000000000] (default: 1000000)
       --PollingWaitSeconds    polling wait seconds [0~30] (default: 0)
       --VisibilityTimeout     visibility timeout in seconds [0~43200] (default: 30)
       --MaxMsgSize            max message size [1024~65536] (default: 65536)
       --MsgRetentionSeconds   message retention seconds [30~43200] (default: 3600)
       --RewindSeconds         rewind seconds [0~1296000] (default: 0)
       --Transaction           transaction, 0:disable, 1:enable (default: 0)
       --FirstQueryInterval    first query interval (default: 0)
       --MaxQueryCount         max query count (default: 0)
       --DeadLetterQueueName   dead letter queue name
       --Policy                dead letter policy, 0:not acked after consume many times, 1:TTL expired (default: 1)
       --MaxReceiveCount       max receive count [1~1000] (default: 1)
       --MaxTimeToLive         max time to live [300~43200] (default: 300)
       --Trace                 trace message, true:enable, false:disable
       --RetentionSizeInMB     retention size in MB [10240~512000] (default: 0)
    -r --region                public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint              special endpoint for manage action (disable region)
    -k --insecure              whether client skip verifies server's certificate
    -t --timeout               client timeout in seconds (default: 5)
    -sid --secretId              secret id
    -key --secretKey             secret key
```

##### `tcmqcli modify topic -h`

```bash
topic - modify topic

  Flags: 
       --version               Displays the program version string.
    -h --help                  Displays help with available flag, subcommand, and positional value parameters.
    -n --name                  topic name
       --MaxMsgSize            max message size [1024~65536] (default: 65536)
       --FilterType            subscribe message filter type, 1:tag, 2:route (default: 1)
       --MsgRetentionSeconds   message retention seconds [60~86400] (default: 86400)
       --Trace                 trace message, true:enable, false:disable
    -r --region                public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint              special endpoint for manage action (disable region)
    -k --insecure              whether client skip verifies server's certificate
    -t --timeout               client timeout in seconds (default: 5)
    -sid --secretId              secret id
    -key --secretKey             secret key
```

##### `tcmqcli modify subscribe -h`

```bash
subscribe - modify subscribe

  Flags: 
       --version               Displays the program version string.
    -h --help                  Displays help with available flag, subcommand, and positional value parameters.
    -n --name                  subscribe name
       --TopicName             topic name
       --Protocol              deliver protocol, [http,queue]
       --Endpoint              endpoint of deliver protocol, http url or queue name
       --NotifyStrategy        deliver notify strategy, 1:BACKOFF_RETRY, 2:EXPONENTIAL_DECAY_RETRY (default: 2)
       --FilterTag             message filter tag, max 5 count and each one max 16 chars
       --BindingKey            message binding key, max 5 count and each one max 64 chars
       --NotifyContentFormat   notify content format, 1:JSON, 2:SIMPLIFIED (default: 2)
    -r --region                public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint              special endpoint for manage action (disable region)
    -k --insecure              whether client skip verifies server's certificate
    -t --timeout               client timeout in seconds (default: 5)
    -sid --secretId              secret id
    -key --secretKey             secret key
```

#### `tcmqcli i -h` 或 `tcmqcli describe -h`

```bash
describe - describe queue/topic/subscribe

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -s --subscribe   subscribe name
    -q --queue       queue name
    -t --topic       topic name
    -f --filter      list filter resource type: queue/topic/subscribe/region
    -l --limit       limit query page size (default: 0)
    -o --offset      begin index of query page (default: 0)
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint    special endpoint for manage action (disable region)
    -k --insecure    whether client skip verifies server's certificate
    -sid --secretId    secret id
    -key --secretKey   secret key
```

#### `tcmqcli l -h` 或 `tcmqcli list -h`

```bash
list - list queue/topic/subscribe/region

  Flags: 
       --version     Displays the program version string.
    -h --help        Displays help with available flag, subcommand, and positional value parameters.
    -s --subscribe   subscribe name
    -q --queue       queue name
    -t --topic       topic name
    -f --filter      list filter resource type: queue/topic/subscribe/region
    -l --limit       limit query page size (default: 0)
    -o --offset      begin index of query page (default: 0)
    -r --region      public cloud region, ex: gz/sh/bj/usw/jp/de...
    -e --endpoint    special endpoint for manage action (disable region)
    -k --insecure    whether client skip verifies server's certificate
    -sid --secretId    secret id
    -key --secretKey   secret key
```

## 执行命令

> 数据流请求 启用region参数默认请求公有云公网地址  
> 使用 --network private 请求内网地址  
> 使用 --uri 参数指定测试环境/私有环境或小集群等请求地址

### 发送消息到队列

```bash
./tcmqcli send --region gz -sid "AKIDxxxxx" -key "xxxxx" -q "myqueue" -m "hello world"
```

### 发送多条消息到队列

```bash
./tcmqcli send -r gz -sid "AKIDxxxxx" -key "xxxxx" -q "myqueue" -m "msg1" -m "msg2" -m "msg3"
```

### 从队列接收多条消息并同时确认消息

```bash
./tcmqcli receive -r gz -sid "AKIDxxxxx" -key "xxxxx" -q "myqueue" -n 10 -ack
```

### 发布多条消息到主题

```bash
./tcmqcli publish -r gz -sid "AKIDxxxxx" -key "xxxxx" -t "mytopic" -tag "TAG" -m "msg 3" -m "msg 4"
```

> 管控流请求 启用region参数默认请求公有云公网地址  
> 使用 --endpoint 参数指定测试环境/私有环境等请求地址

### 创建队列

```bash
# 使用默认参数创建队列
./tcmqcli create queue -n myqueue -r gz -sid AKIDxxxxx -key xxxxx
```

### 创建主题

```bash
# 使用默认参数创建队列
./tcmqcli create topic -n mytopic -r gz -sid AKIDxxxxx -key xxxxx
```

### 创建订阅

```bash
./tcmqcli create subscribe -n mysubscribe --TopicName mytopic --Protocol queue --Endpoint myqueue -r gz -sid AKIDxxxxx -key xxxxx
```

### 删除队列/主题/订阅

```bash
./tcmqcli remove -q myqueue -r gz -sid AKIDxxxxx -key xxxxx
./tcmqcli remove -t mytopic -r gz -sid AKIDxxxxx -key xxxxx
./tcmqcli remove -s mysubscribe -r gz -sid AKIDxxxxx -key xxxxx
```

### 修改队列/主题/订阅

> 参考 `tcmqcli modify --help`

### 描述队列/主题/订阅

```bash
./tcmqcli describe --queue myqueue -r gz -sid AKIDxxxxx -key xxxxx
./tcmqcli describe --topic mytopic -r gz -sid AKIDxxxxx -key xxxxx
./tcmqcli describe --subscribe mysubscribe -r gz -sid AKIDxxxxx -key xxxxx
```

### 列表队列/主题/订阅

```bash
./tcmqcli list --queue myqueue -r gz -sid AKIDxxxxx -key xxxxx
./tcmqcli list --topic mytopic -r gz -sid AKIDxxxxx -key xxxxx
./tcmqcli list --subscribe mysubscribe -r gz -sid AKIDxxxxx -key xxxxx
```

> 列出TCMQ数据流所有可访问地域

```bash
./tcmqcli list -f region
./tcmqcli list -f region -r gz
```

```
 Region  AP Code            Public Private
     bj  ap-beijing         Arrive Arrive 
   bjjr  ap-beijing-fsi     Arrive Arrive 
     ca  na-toronto         Arrive Arrive 
     cd  ap-chengdu         Arrive Arrive 
  cgoec  ap-zhengzhou-ec    -      -      
     cq  ap-chongqing       Arrive Arrive 
   csec  ap-changsha-ec     -      -      
     de  eu-frankfurt       Arrive Arrive 
    dub  me-dubai           -      -      
   fzec  ap-fuzhou-ec       -      -      
     gy  ap-guiyang         -      -      
     gz  ap-guangzhou       Arrive Arrive 
 gzopen  ap-guangzhou-open  -      -      
  hfeec  ap-hefei-ec        -      -      
     hk  ap-hongkong        Arrive Arrive 
   hzec  ap-hangzhou-ec     -      -      
     in  ap-mumbai          Arrive Arrive 
    jkt  ap-jakarta         -      -      
   jnec  ap-jinan-ec        -      -      
     jp  ap-tokyo           Arrive Arrive 
     kr  ap-seoul           Arrive Arrive 
     la  na-losangeles      -      -      
     nj  ap-nanjing         -      -      
 others  ap-others          -      -      
     qy  ap-qingyuan        -      -      
   qyxa  ap-qingyuan-xinan  -      -      
     ru  eu-moscow          -      -      
    sao  sa-saopaulo        Arrive Arrive 
     sg  ap-singapore       Arrive Arrive 
     sh  ap-shanghai        Arrive Arrive 
  shadc  ap-shanghai-adc    -      -      
  sheec  ap-shenyang-ec     -      -      
   shjr  ap-shanghai-fsi    Arrive Arrive 
shjrtce  ap-shenzhen-fsitce -      -      
  sjwec  ap-shijiazhuang-ec -      -      
     sl  sl-saopaulo        -      -      
    syd  au-sydney          -      -      
   szjr  ap-shenzhen-fsi    Arrive Arrive 
szjrtce  ap-shanghai-fsitce -      -      
szsycft  ap-shenzhen-sycft  -      -      
    szx  ap-shenzhen        -      -      
     th  ap-bangkok         Arrive Arrive 
     tj  ap-beijing-z1      -      -      
    tpe  ap-taipei          -      -      
    tsn  ap-tianjin         -      -      
    use  na-ashburn         Arrive Arrive 
    usw  na-siliconvalley   Arrive Arrive 
   whec  ap-wuhan-ec        -      -      
   xbec  ap-xibei-ec        -      -      
  xiyec  ap-xian-ec         -      -
```
