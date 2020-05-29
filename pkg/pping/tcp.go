package pping

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type TcpPingResult struct {
	Time int
	Err  error
}

func (this *TcpPingResult) Result() int {
	return this.Time
}

func (this *TcpPingResult) Error() error {
	return this.Err
}

func (this *TcpPingResult) String() string {
	if this.Err != nil {
		return fmt.Sprintf("%s", this.Err)
	} else {
		return fmt.Sprintf("time = %d ms", this.Time)
	}
}

type TcpPing struct {
	IP      net.IP
	Port    uint16
	Timeout time.Duration
}

func (this *TcpPing) Ping() IPingResult {
	dialer := &net.Dialer{
		Timeout:   this.Timeout,
		KeepAlive: -1,
	}
	t0 := time.Now()
	conn, err := dialer.Dial("tcp", net.JoinHostPort(this.IP.String(), strconv.FormatUint(uint64(this.Port), 10)))
	if err != nil {
		return &TcpPingResult{0, err}
	}
	defer conn.Close()
	return &TcpPingResult{int(time.Now().Sub(t0).Milliseconds()), nil}
}

func NewTcpPing(ip net.IP, port uint16, timeout time.Duration) *TcpPing {
	return &TcpPing{ip, port, timeout}
}
