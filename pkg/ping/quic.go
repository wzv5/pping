package ping

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type QuicPingResult struct {
	Time        int
	Err         error
	IP          net.IP
	QUICVersion uint32
	TLSVersion  uint16
}

func (this *QuicPingResult) Result() int {
	return this.Time
}

func (this *QuicPingResult) Error() error {
	return this.Err
}

func (this *QuicPingResult) String() string {
	if this.Err != nil {
		return fmt.Sprintf("%s", this.Err)
	} else {
		return fmt.Sprintf("%s: quic=%s, tls=%s, time=%d ms", this.IP.String(), quic.VersionNumber(this.QUICVersion).String(), tlsVersionToString(this.TLSVersion), this.Time)
	}
}

type QuicPing struct {
	Host    string
	Port    uint16
	Timeout time.Duration

	// 以下为可选参数
	Insecure bool
	ALPN     string
	IP       net.IP
}

func (this *QuicPing) Ping() IPingResult {
	return this.PingContext(context.Background())
}

func (this *QuicPing) PingContext(ctx context.Context) IPingResult {
	ip := cloneIP(this.IP)
	if ip == nil {
		var err error
		ip, err = LookupFunc(this.Host)
		if err != nil {
			return this.errorResult(err)
		}
	}
	addr := net.JoinHostPort(ip.String(), fmt.Sprint(this.Port))

	alpn := http3.NextProtoH3
	if this.ALPN != "" {
		alpn = this.ALPN
	}
	tlsconf := tls.Config{
		ServerName:         this.Host,
		InsecureSkipVerify: this.Insecure,
		NextProtos:         []string{alpn},
	}
	quicconf := quic.Config{
		HandshakeIdleTimeout: this.Timeout,
	}
	t0 := time.Now()
	conn, err := quic.DialAddr(ctx, addr, &tlsconf, &quicconf)
	if err != nil {
		return this.errorResult(err)
	}
	closecode := uint64(http3.ErrCodeNoError)
	if alpn != http3.NextProtoH3 {
		closecode = 0
	}
	defer conn.CloseWithError(quic.ApplicationErrorCode(closecode), "")
	return &QuicPingResult{int(time.Since(t0).Milliseconds()), nil, ip, uint32(conn.ConnectionState().Version), conn.ConnectionState().TLS.Version}
}

func NewQuicPing(host string, port uint16, timeout time.Duration) *QuicPing {
	return &QuicPing{
		Host:    host,
		Port:    port,
		Timeout: timeout,
	}
}

func (this *QuicPing) errorResult(err error) *QuicPingResult {
	r := &QuicPingResult{}
	r.Err = err
	return r
}

var (
	_ IPing       = (*QuicPing)(nil)
	_ IPingResult = (*QuicPingResult)(nil)
)
