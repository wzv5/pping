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
	IP     net.IP
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
		return fmt.Sprintf("%s: proto=%s, status=%d, length=%d, time=%d ms", this.IP.String(), this.Proto, this.Status, this.Length, this.Time)
	}
}

type HttpPing struct {
	Method  string
	URL     string
	Timeout time.Duration

	// 以下参数全部为可选
	DisableHttp2       bool
	DisableCompression bool
	Insecure           bool
	Referrer           string
	UserAgent          string
	IP                 net.IP
}

func (this *HttpPing) Ping() IPingResult {
	return this.PingContext(context.Background())
}

func (this *HttpPing) PingContext(ctx context.Context) IPingResult {
	u, err := url.Parse(this.URL)
	if err != nil {
		return this.errorResult(err)
	}
	host := u.Hostname()
	var ip net.IP
	if this.IP != nil {
		ip = make(net.IP, len(this.IP))
		copy(ip, this.IP)
	} else {
		var err error
		ip, err = LookupFunc(host)
		if err != nil {
			return this.errorResult(err)
		}
	}

	dialer := &net.Dialer{
		Timeout:   this.Timeout,
		KeepAlive: -1,
	}

	dialfunc := func(ctx context.Context, network, address string) (net.Conn, error) {
		h, p, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		if ip == nil || !strings.EqualFold(h, host) {
			var err error
			ip, err = LookupFunc(h)
			if err != nil {
				return nil, err
			}
		}
		addr := net.JoinHostPort(ip.String(), p)
		return dialer.DialContext(ctx, network, addr)
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

	req, err := http.NewRequestWithContext(ctx, this.Method, this.URL, nil)
	if err != nil {
		return this.errorResult(err)
	}
	if this.UserAgent == "" {
		this.UserAgent = "httping"
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
	t0 := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return this.errorResult(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return this.errorResult(err)
	}
	return &HttpPingResult{int(time.Now().Sub(t0).Milliseconds()), resp.Proto, resp.StatusCode, len(body), nil, ip}
}

func (this *HttpPing) errorResult(err error) *HttpPingResult {
	r := &HttpPingResult{}
	r.Err = err
	return r
}

func NewHttpPing(method, url string, timeout time.Duration) *HttpPing {
	return &HttpPing{
		Method:  method,
		URL:     url,
		Timeout: timeout,
	}
}

var (
	_ IPing       = (*HttpPing)(nil)
	_ IPingResult = (*HttpPingResult)(nil)
)
