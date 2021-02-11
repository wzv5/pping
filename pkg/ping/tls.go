package ping

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"
)

type TlsPingResult struct {
	ConnectionTime int
	HandshakeTime  int
	TLSVersion     uint16
	Err            error
	IP             net.IP
}

func (this *TlsPingResult) Result() int {
	return this.ConnectionTime + this.HandshakeTime
}

func (this *TlsPingResult) Error() error {
	return this.Err
}

func (this *TlsPingResult) String() string {
	if this.Err != nil {
		return fmt.Sprintf("%s", this.Err)
	} else {
		return fmt.Sprintf("%s: protocol=%s, connection=%d ms, handshake=%d ms, time=%d ms", this.IP.String(), tlsVersionToString(this.TLSVersion), this.ConnectionTime, this.HandshakeTime, this.Result())
	}
}

type TlsPing struct {
	Host              string
	Port              uint16
	ConnectionTimeout time.Duration
	HandshakeTimeout  time.Duration

	// 以下为可选参数
	TlsVersion uint16
	Insecure   bool
	IP         net.IP
}

func (this *TlsPing) Ping() IPingResult {
	return this.PingContext(context.Background())
}

func (this *TlsPing) PingContext(ctx context.Context) IPingResult {
	ip := cloneIP(this.IP)
	if ip == nil {
		var err error
		ip, err = LookupFunc(this.Host)
		if err != nil {
			return this.errorResult(err)
		}
	}

	dialer := &net.Dialer{
		Timeout:   this.ConnectionTimeout,
		KeepAlive: -1,
	}
	t0 := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.FormatUint(uint64(this.Port), 10)))
	if err != nil {
		return this.errorResult(err)
	}
	defer conn.Close()
	t1 := time.Now()
	config := &tls.Config{
		ServerName:         this.Host,
		MinVersion:         this.TlsVersion,
		MaxVersion:         this.TlsVersion,
		InsecureSkipVerify: this.Insecure,
	}
	client := tls.Client(conn, config)
	client.SetDeadline(time.Now().Add(this.HandshakeTimeout))
	err = client.Handshake()
	if err != nil {
		return this.errorResult(err)
	}
	defer client.Close()
	t2 := time.Now()
	return &TlsPingResult{int(t1.Sub(t0).Milliseconds()), int(t2.Sub(t1).Milliseconds()), client.ConnectionState().Version, nil, ip}
}

func NewTlsPing(host string, port uint16, ct, ht time.Duration) *TlsPing {
	return &TlsPing{
		Host:              host,
		Port:              port,
		ConnectionTimeout: ct,
		HandshakeTimeout:  ht,
	}
}

func (this *TlsPing) errorResult(err error) *TlsPingResult {
	r := &TlsPingResult{}
	r.Err = err
	return r
}

var (
	_ IPing       = (*TlsPing)(nil)
	_ IPingResult = (*TlsPingResult)(nil)
)
