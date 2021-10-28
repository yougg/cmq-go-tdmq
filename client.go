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
	Id         uint64 // clientRequestId
	Uri        string // ex: http://gateway.tdmq.io
	Path       string // "/", "/v2/index.php"
	Method     string // GET, POST
	SignMethod string // HmacSHA1, HmacSHA256
	SecretId   string // AKIDxxxxx
	SecretKey  string
	AppId      uint64 // appId for privatization, need gateway server option enabled

	Debug  bool // weather print request message
	client *http.Client
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewClient(uri, secretId, secretKey string, t time.Duration) *Client {
	return &Client{
		Id:         uint64(rand.Uint32()),
		Uri:        uri,
		Path:       `/`,
		Method:     http.MethodPost,
		SignMethod: HmacSHA256,
		SecretId:   secretId,
		SecretKey:  secretKey,

		client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: t,
		},
	}
}

func (c *Client) call(values url.Values) (msg *msgResponse, err error) {
	var u *url.URL
	u, err = url.Parse(c.Uri)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	values.Set(`RequestClient`, currentVersion)
	if c.Id > 0 {
		values.Set(`clientRequestId`, strconv.FormatUint(c.Id, 10))
	}
	if c.AppId > 0 && c.SecretId == `` && c.SecretKey == `` {
		values.Set(`appId`, strconv.FormatUint(c.AppId, 10))
	} else {
		values.Set(`SecretId`, c.SecretId)
		values.Set(`SignatureMethod`, c.SignMethod)
		values.Set(`Nonce`, strconv.FormatUint(uint64(rand.Uint32()), 10))
		values.Set(`Timestamp`, strconv.FormatInt(time.Now().Unix(), 10))
		values.Set(`Signature`, c.sign(u.Host, values))
	}

	var query string
	query, err = url.QueryUnescape(values.Encode())
	if err != nil {
		return nil, fmt.Errorf("unescape data: %w", err)
	}
	var reader io.Reader
	switch c.Method {
	case http.MethodGet:
		u.RawQuery = query
	case http.MethodPost:
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
	var resp *http.Response
	if c.Debug {
		fmt.Printf("curl -i -X %s '%s'", c.Method, u.String())
		if c.Method == http.MethodPost {
			fmt.Printf(" -d '%s'", query)
		}
		fmt.Println()
	}
	resp, err = c.client.Do(req)
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
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, fmt.Errorf("json decode: %w, response: %s", err, raw)
	}
	return msg, nil
}
