package ping

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
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
	Http3              bool
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
	orighost := u.Host
	host := u.Hostname()
	port := u.Port()
	ip := cloneIP(this.IP)
	if ip == nil {
		var err error
		ip, err = LookupFunc(host)
		if err != nil {
			return this.errorResult(err)
		}
	}
	ipstr := ip.String()
	if isIPv6(ip) {
		ipstr = fmt.Sprintf("[%s]", ipstr)
	}
	if port != "" {
		u.Host = fmt.Sprintf("%s:%s", ipstr, port)
	} else {
		u.Host = ipstr
	}
	url2 := u.String()

	var transport http.RoundTripper
	if this.Http3 {
		trans := &http3.Transport{
			DisableCompression: this.DisableCompression,
			QUICConfig: &quic.Config{
				KeepAlivePeriod: 0,
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: this.Insecure,
				ServerName:         host,
			},
		}
		defer trans.Close()
		transport = trans
	} else {
		trans := http.DefaultTransport.(*http.Transport).Clone()
		trans.DisableKeepAlives = true
		trans.MaxIdleConnsPerHost = -1
		trans.DisableCompression = this.DisableCompression
		trans.ForceAttemptHTTP2 = !this.DisableHttp2
		trans.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: this.Insecure,
			ServerName:         host,
		}
		transport = trans
	}

	req, err := http.NewRequestWithContext(ctx, this.Method, url2, nil)
	if err != nil {
		return this.errorResult(err)
	}
	ua := "httping"
	if this.UserAgent != "" {
		ua = this.UserAgent
	}
	req.Header.Set("User-Agent", ua)
	if this.Referrer != "" {
		req.Header.Set("Referer", this.Referrer)
	}
	req.Host = orighost
	client := &http.Client{}
	client.Transport = transport
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
