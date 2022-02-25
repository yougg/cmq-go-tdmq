# TDMQ-CMQ Command Line

## 安装/编译

```bash
go install github.com/yougg/cmq-go-tdmq/cmd/cli@main
```

```bash
git clone https://github.com/yougg/cmq-go-tdmq.git
cd cmq-go-tdmq/cmd/cli/
GOOS=linux go generate tdmqcli.go
GOOS=darwin go generate tdmqcli.go
GOOS=windows go generate tdmqcli.go
```

## 执行命令

```bash
./tcmqcli -h # 查看命令参数帮助

# 发送消息
./tcmqcli -d -a send -uri 'https://cmq-gz.public.tencenttdmq.com' -sid 'AKIDxxxxx' -key 'xxxxx' -q 'myqueue' -m 'hello world'
# 发送多条消息
./tcmqcli -d -a sends -uri 'https://cmq-gz.public.tencenttdmq.com' -sid 'AKIDxxxxx' -key 'xxxxx' -q 'myqueue' -m 'msg 0' -m 'msg 1' -m 'msg 2'

# 接收多条消息
./tcmqcli -a receives -uri 'https://cmq-gz.public.tencenttdmq.com' -sid 'AKIDxxxxx' -key 'xxxxx' -q 'myqueue' -n 10 -ack

# 发布多条消息
./tcmqcli -a publishes -uri 'https://cmq-gz.public.tencenttdmq.com' -sid 'AKIDxxxxx' -key 'xxxxx' -t 'mytopic' -tag 'TAG' -m 'msg 3' -m 'msg 4'
```