package ping

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type DnsPingResult struct {
	Time int
	Err  error
	IP   net.IP
}

func (this *DnsPingResult) Result() int {
	return this.Time
}

func (this *DnsPingResult) Error() error {
	return this.Err
}

func (this *DnsPingResult) String() string {
	if this.Err != nil {
		return fmt.Sprintf("%s", this.Err)
	} else {
		return fmt.Sprintf("%s: time=%d ms", this.IP.String(), this.Time)
	}
}

type DnsPing struct {
	host    string
	Port    uint16
	Timeout time.Duration

	// udp, tcp, tcp-tls，默认 udp
	Net string

	// A, AAAA, NS, ...，默认 NS
	Type string

	// 查询域名，默认 .
	Domain string

	// Net 为 tcp-tls 时，是否跳过证书验证
	Insecure bool

	ip net.IP
}

func (this *DnsPing) SetHost(host string) {
	this.host = host
	this.ip = net.ParseIP(host)
}

func (this *DnsPing) Host() string {
	return this.host
}

func (this *DnsPing) Ping() IPingResult {
	return this.PingContext(context.Background())
}

func (this *DnsPing) PingContext(ctx context.Context) IPingResult {
	ip := cloneIP(this.ip)
	if ip == nil {
		var err error
		ip, err = LookupFunc(this.host)
		if err != nil {
			return &DnsPingResult{0, err, nil}
		}
	}

	msg := &dns.Msg{}
	qtype, ok := dns.StringToType[this.Type]
	if !ok {
		return &DnsPingResult{0, errors.New("unknown type"), nil}
	}
	if !strings.HasSuffix(this.Domain, ".") {
		this.Domain += "."
	}
	msg.SetQuestion(this.Domain, qtype)
	msg.MsgHdr.RecursionDesired = true

	client := &dns.Client{}
	client.Net = this.Net
	client.Timeout = this.Timeout
	client.TLSConfig = &tls.Config{
		ServerName:         this.host,
		InsecureSkipVerify: this.Insecure,
	}

	t0 := time.Now()
	r, _, err := client.ExchangeContext(ctx, msg, net.JoinHostPort(ip.String(), strconv.Itoa(int(this.Port))))
	if err != nil {
		return &DnsPingResult{0, err, nil}
	}
	if r == nil || r.Response == false || r.Opcode != dns.OpcodeQuery {
		return &DnsPingResult{0, errors.New("response error"), nil}
	}
	return &DnsPingResult{int(time.Now().Sub(t0).Milliseconds()), nil, ip}
}

func NewDnsPing(host string, timeout time.Duration) *DnsPing {
	return &DnsPing{
		host:     host,
		Port:     53,
		Timeout:  timeout,
		Net:      "udp",
		Type:     "NS",
		Domain:   ".",
		Insecure: false,
		ip:       net.ParseIP(host),
	}
}

var (
	_ IPing       = (*DnsPing)(nil)
	_ IPingResult = (*DnsPingResult)(nil)
)
