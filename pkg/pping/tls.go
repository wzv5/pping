package pping

import (
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
		return fmt.Sprintf("proto = %s, connection = %d ms, handshake = %d ms, time = %d ms", TlsVersionToString(this.TLSVersion), this.ConnectionTime, this.HandshakeTime, this.Result())
	}
}

type TlsPing struct {
	Host              string
	IP                net.IP
	Port              uint16
	ConnectionTimeout time.Duration
	HandshakeTimeout  time.Duration
	TlsVersion        uint16
	Insecure          bool
}

func (this *TlsPing) Ping() IPingResult {
	dialer := &net.Dialer{
		Timeout:   this.ConnectionTimeout,
		KeepAlive: -1,
	}
	t0 := time.Now()
	conn, err := dialer.Dial("tcp", net.JoinHostPort(this.IP.String(), strconv.FormatUint(uint64(this.Port), 10)))
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
	return &TlsPingResult{int(t1.Sub(t0).Milliseconds()), int(t2.Sub(t1).Milliseconds()), client.ConnectionState().Version, nil}
}

func NewTlsPing(host string, ip net.IP, port uint16, ct, ht time.Duration, tlsver uint16, insecure bool) *TlsPing {
	return &TlsPing{host, ip, port, ct, ht, tlsver, insecure}
}

func (this *TlsPing) errorResult(err error) *TlsPingResult {
	r := &TlsPingResult{}
	r.Err = err
	return r
}
