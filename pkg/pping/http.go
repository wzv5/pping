package pping

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpPingResult struct {
	Time   int
	Proto  string
	Status int
	Length int
	Err    error
}

func (this *HttpPingResult) Result() int {
	return this.Time
}

func (this *HttpPingResult) Error() error {
	return this.Err
}

func (this *HttpPingResult) String() string {
	if this.Err != nil {
		return fmt.Sprintf("%s", this.Err)
	} else {
		return fmt.Sprintf("proto = %s, status = %d, length = %d, time = %d ms", this.Proto, this.Status, this.Length, this.Time)
	}
}

type HttpPing struct {
	Method             string
	URL                string
	DisableHttp2       bool
	DisableCompression bool
	Insecure           bool
	Timeout            time.Duration

	// 以下参数全部为可选
	Referrer  string
	UserAgent string
	IP        net.IP
}

func (this *HttpPing) Ping() IPingResult {
	t0 := time.Now()

	u, err := url.Parse(this.URL)
	if err != nil {
		return this.ErrResult(err)
	}
	host := u.Hostname()

	dialer := &net.Dialer{
		Timeout:   this.Timeout,
		KeepAlive: -1,
	}

	var dialfunc = dialer.DialContext
	if this.IP != nil {
		dialfunc = func(ctx context.Context, network, address string) (net.Conn, error) {
			h, p, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			addr := address
			if strings.EqualFold(h, host) {
				addr = net.JoinHostPort(this.IP.String(), p)
			}
			return dialer.DialContext(ctx, network, addr)
		}
	}

	trans := &http.Transport{
		DialContext:           dialfunc,
		Proxy:                 http.ProxyFromEnvironment,
		DisableKeepAlives:     true,
		DisableCompression:    this.DisableCompression,
		ForceAttemptHTTP2:     !this.DisableHttp2,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: this.Insecure,
		},
	}

	req, err := http.NewRequest(this.Method, this.URL, nil)
	if err != nil {
		return this.ErrResult(err)
	}
	if this.UserAgent == "" {
		this.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"
	}
	req.Header.Set("User-Agent", this.UserAgent)
	if this.Referrer != "" {
		req.Header.Set("Referer", this.Referrer)
	}
	client := &http.Client{}
	client.Transport = trans
	client.Timeout = this.Timeout
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		return this.ErrResult(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return this.ErrResult(err)
	}
	return &HttpPingResult{int(time.Now().Sub(t0).Milliseconds()), resp.Proto, resp.StatusCode, len(body), nil}
}

func (this *HttpPing) ErrResult(err error) *HttpPingResult {
	r := &HttpPingResult{}
	r.Err = err
	return r
}

func NewHttpPing(method, url string, disablehttp2, disablecompression, insecure bool, timeout time.Duration, refer, ua string, ip net.IP) *HttpPing {
	return &HttpPing{method, url, disablehttp2, disablecompression, insecure, timeout, refer, ua, ip}
}
