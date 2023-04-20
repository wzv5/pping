package ping

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
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
		return fmt.Sprintf("%s: protocol=%s, status=%d, length=%d, time=%d ms", this.IP.String(), this.Proto, this.Status, this.Length, this.Time)
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
	ip := cloneIP(this.IP)
	if ip == nil {
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

	trans := http.DefaultTransport.(*http.Transport).Clone()
	trans.DialContext = dialfunc
	trans.DisableKeepAlives = true
	trans.MaxIdleConnsPerHost = -1
	trans.DisableCompression = this.DisableCompression
	trans.ForceAttemptHTTP2 = !this.DisableHttp2
	trans.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: this.Insecure,
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return this.errorResult(err)
	}
	return &HttpPingResult{int(time.Since(t0).Milliseconds()), resp.Proto, resp.StatusCode, len(body), nil, ip}
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
