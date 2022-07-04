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

## 执行命令

```bash
./tcmqcli -h # 查看命令参数帮助

# 发送消息到队列
./tcmqcli -d -a send -uri "https://cmq-gz.public.tencenttdmq.com" -sid "AKIDxxxxx" -key "xxxxx" -q "myqueue" -m "hello world"
# 发送多条消息到队列
./tcmqcli -d -a sends -uri "https://cmq-gz.public.tencenttdmq.com" -sid "AKIDxxxxx" -key "xxxxx" -q "myqueue" -m "msg 0" -m "msg 1" -m "msg 2"

# 从队列接收多条消息并同时确认消息
./tcmqcli -a receives -uri "https://cmq-gz.public.tencenttdmq.com" -sid "AKIDxxxxx" -key "xxxxx" -q "myqueue" -n 10 -ack

# 发布多条消息到主题
./tcmqcli -a publishes -uri "https://cmq-gz.public.tencenttdmq.com" -sid "AKIDxxxxx" -key "xxxxx" -t "mytopic" -tag "TAG" -m "msg 3" -m "msg 4"
```