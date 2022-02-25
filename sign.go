package tdmq

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"net/url"
	"sort"
	"strings"
)

const (
	HmacSHA1   = `HmacSHA1`
	HmacSHA256 = `HmacSHA256`
)

func (c *Client) sign(host string, values url.Values) string {
	plain := []string{c.Method}
	if len(host) > 0 {
		plain = append(plain, host)
	}
	if path := c.Url.Path; len(path) > 0 {
		idx := strings.Index(path, `?`)
		if idx >= 0 {
			path = path[:idx]
		}
		plain = append(plain, path)
	}
	plain = append(plain, `?`)
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		k = strings.Replace(k, `_`, `.`, -1)
		plain = append(plain, k, `=`, strings.Join(values[k], ``), `&`)
	}

	if plain[len(plain)-1] == `&` {
		plain = plain[:len(plain)-1]
	}
	if c.Debug {
		fmt.Println("String to sign:", strings.Join(plain, ``))
	}

	var h hash.Hash
	if c.SignMethod == HmacSHA256 {
		h = hmac.New(sha256.New, []byte(c.SecretKey))
	} else {
		h = hmac.New(sha1.New, []byte(c.SecretKey))
	}
	_, err := h.Write([]byte(strings.Join(plain, ``)))
	if err != nil {
		return ``
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
