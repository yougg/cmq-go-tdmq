package tdmq

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	id         uint64   // clientRequestId
	Url        *url.URL // ex: http://gateway.tdmq.io
	Method     string   // GET, POST
	SignMethod string   // HmacSHA1, HmacSHA256
	SecretId   string   // AKIDxxxxx
	SecretKey  string
	Token      string
	AppId      uint64 // appId for privatization, need gateway server option enabled
	Header     map[string]string

	Debug      bool // weather print request message
	HttpClient *http.Client
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewClient create TDMQ CMQ client
//  input: uri string request uri for TDMQ CMQ service
//  input: secretId string user secret id from tencent cloud account
//  input: secretKey string user secret key from tencent cloud account
//  input: t time.Duration http client request timeout
//  input: keepalive bool http client connection keep alive to server
//  return: *Client
func NewClient(uri, secretId, secretKey string, t time.Duration, keepalive ...bool) (c *Client, err error) {
	var shortLive bool
	if len(keepalive) > 0 {
		shortLive = !keepalive[0]
	}
	c = &Client{
		id:         uint64(rand.Uint32()),
		Method:     http.MethodPost,
		SignMethod: HmacSHA1,
		SecretId:   secretId,
		SecretKey:  secretKey,

		HttpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: shortLive,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: InsecureSkipVerify,
				},
			},
			Timeout: t,
		},
	}

	c.Url, err = url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if len(c.Url.Path) == 0 {
		c.Url.Path = `/`
	}
	return
}

func (c *Client) call(values url.Values) (msg *msgResponse, err error) {
	values.Set(`RequestClient`, currentVersion)
	if c.id > 0 {
		values.Set(`clientRequestId`, strconv.FormatUint(c.id, 10))
	}
	if c.AppId > 0 && c.SecretId == `` && c.SecretKey == `` {
		values.Set(`appId`, strconv.FormatUint(c.AppId, 10))
	} else {
		if c.Token != `` {
			values.Set(`Token`, c.Token)
		}
		values.Set(`SecretId`, c.SecretId)
		values.Set(`SignatureMethod`, c.SignMethod)
		values.Set(`Nonce`, strconv.FormatUint(uint64(rand.Uint32()), 10))
		values.Set(`Timestamp`, strconv.FormatInt(time.Now().Unix(), 10))
		values.Set(`Signature`, c.sign(c.Url.Host, values))
	}

	// https://cloud.tencent.com/document/product/406/5906
	var query string
	var reader io.Reader
	var u *url.URL
	switch c.Method {
	case http.MethodGet:
		// 请求方法是GET，对所有请求参数值做URL编码
		query = values.Encode()
		u = &(*c.Url) // copy url
		u.RawQuery = query
	case http.MethodPost:
		u = c.Url // reference url, avoid memory copy
		var plain []string
		for k, v := range values {
			if len(v) > 0 {
				v[0] = url.QueryEscape(strings.Join(v, ``))
			}
			plain = append(plain, k, `=`, strings.Join(v, ``), `&`)
		}
		if plain[len(plain)-1] == `&` {
			plain = plain[:len(plain)-1]
		}
		query = strings.Join(plain, ``)
		reader = strings.NewReader(query)
	default:
		return nil, errors.New("unsupported request method: " + c.Method)
	}
	var req *http.Request
	req, err = http.NewRequest(c.Method, u.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	for k, v := range c.Header {
		req.Header.Set(k, v)
	}
	var resp *http.Response
	if c.Debug {
		fmt.Printf("curl -i -X %s -H 'Content-Type:application/x-www-form-urlencoded' '%s'", c.Method, u.String())
		if c.Method == http.MethodPost {
			fmt.Printf(" -d '%s'", query)
		}
		fmt.Println()
	}
	resp, err = c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http client do request: %w", err)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	raw := string(data)
	if c.Debug {
		fmt.Println("Status:", resp.StatusCode)
		fmt.Println("Response:", raw)
	}
	msg = &msgResponse{
		Status: resp.StatusCode,
		Raw:    raw,
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("got invalid json response: %s", raw)
	}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, fmt.Errorf("json decode: %w, response: %s", err, raw)
	}
	return msg, nil
}
