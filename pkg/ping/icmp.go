package ping

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type IcmpPingResult struct {
	Time int
	Err  error
	IP   net.IP
	TTL  int
}

func (this *IcmpPingResult) Result() int {
	return this.Time
}

func (this *IcmpPingResult) Error() error {
	return this.Err
}

func (this *IcmpPingResult) String() string {
	if this.Err != nil {
		return fmt.Sprintf("%s", this.Err)
	} else {
		return fmt.Sprintf("%s: time=%d ms, TTL=%d", this.IP.String(), this.Time, this.TTL)
	}
}

type IcmpPing struct {
	host    string
	Timeout time.Duration

	ip         net.IP
	Privileged bool
}

func (this *IcmpPing) SetHost(host string) {
	this.host = host
	this.ip = net.ParseIP(host)
}

func (this *IcmpPing) Host() string {
	return this.host
}

func NewIcmpPing(host string, timeout time.Duration) *IcmpPing {
	p := &IcmpPing{
		Timeout: timeout,
	}
	p.SetHost(host)
	return p
}

func (this *IcmpPing) Ping() IPingResult {
	return this.PingContext(context.Background())
}

func (this *IcmpPing) PingContext(ctx context.Context) IPingResult {
	pingfunc := this.ping_rootless
	if this.Privileged {
		pingfunc = this.ping_root
	}
	return pingfunc(ctx)
}

func (this *IcmpPing) ping_root(ctx context.Context) IPingResult {
	return this.rawping("ip")
}

// https://github.com/sparrc/go-ping/blob/master/ping.go

func (this *IcmpPing) rawping(network string) IPingResult {
	// 解析IP
	ip, isipv6, err := this.parseip()
	if err != nil {
		return this.errorResult(err)
	}

	// 创建连接
	conn, err := this.getconn(network, ip, isipv6)
	if err != nil {
		return this.errorResult(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(this.Timeout))

	// 发送
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sendData := make([]byte, 32)
	r.Read(sendData)
	id := os.Getpid() & 0xffff
	sendMsg := this.getmsg(isipv6, id, 0, sendData)
	sendMsgBytes, err := sendMsg.Marshal(nil)
	if err != nil {
		return this.errorResult(err)
	}
	var dst net.Addr = &net.IPAddr{IP: ip}
	if network == "udp" {
		dst = &net.UDPAddr{IP: ip}
	}
	sendAt := time.Now()
	for {
		if _, err := conn.WriteTo(sendMsgBytes, dst); err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
		}
		break
	}

	recvBytes := make([]byte, 1500)
	recvSize := 0

	for {
		ttl := -1
		var peer net.Addr
		if isipv6 {
			var cm *ipv6.ControlMessage
			recvSize, cm, peer, err = conn.IPv6PacketConn().ReadFrom(recvBytes)
			if cm != nil {
				ttl = cm.HopLimit
			}
		} else {
			var cm *ipv4.ControlMessage
			recvSize, cm, peer, err = conn.IPv4PacketConn().ReadFrom(recvBytes)
			if cm != nil {
				ttl = cm.TTL
			}
		}
		if err != nil {
			return this.errorResult(err)
		}

		recvAt := time.Now()
		recvProto := 1
		if isipv6 {
			recvProto = 58
		}
		recvMsg, err := icmp.ParseMessage(recvProto, recvBytes[:recvSize])
		if err != nil {
			return this.errorResult(err)
		}
		recvData, recvID, recvType := this.parserecvmsg(isipv6, recvMsg)
		// 修正数据长度
		if len(recvData) > len(sendData) {
			recvData = recvData[len(recvData)-len(sendData):]
		}
		// 收到的数据和发送的数据不一致，继续接收
		if !bytes.Equal(recvData, sendData) {
			continue
		}
		// 是 echo 回复，但 ID 不一致，继续接收
		if recvType == 1 && network == "ip" && recvID != id {
			continue
		}

		if peer != nil {
			if _ip := net.ParseIP(peer.String()); _ip != nil {
				ip = _ip
			}
		}

		switch recvType {
		case 1:
			// echo
			return &IcmpPingResult{
				TTL:  ttl,
				Time: int(recvAt.Sub(sendAt).Milliseconds()),
				IP:   ip,
			}
		case 2:
			// destination unreachable
			return this.errorResult(errors.New(fmt.Sprintf("%s: destination unreachable", ip.String())))
		case 3:
			// time exceeded
			return this.errorResult(errors.New(fmt.Sprintf("%s: time exceeded", ip.String())))
		}
	}
}

func (this *IcmpPing) parseip() (ip net.IP, ipv6 bool, err error) {
	err = nil
	ip = cloneIP(this.ip)
	if ip == nil {
		ip, err = LookupFunc(this.host)
		if err != nil {
			return
		}
	}
	if isIPv4(ip) {
		ipv6 = false
	} else if isIPv6(ip) {
		ipv6 = true
	} else {
		err = errors.New("lookup ip failed")
	}
	return
}

func (this *IcmpPing) getconn(network string, ip net.IP, isipv6 bool) (*icmp.PacketConn, error) {
	ipv4Proto := map[string]string{"ip": "ip4:icmp", "udp": "udp4"}
	ipv6Proto := map[string]string{"ip": "ip6:ipv6-icmp", "udp": "udp6"}
	icmpnetwork := ""
	if isipv6 {
		icmpnetwork = ipv6Proto[network]
	} else {
		icmpnetwork = ipv4Proto[network]
	}
	conn, err := icmp.ListenPacket(icmpnetwork, "")
	if err != nil {
		return nil, err
	}
	if isipv6 {
		conn.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true)
	} else {
		conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
	}
	return conn, nil
}

func (this *IcmpPing) getmsg(isipv6 bool, id, seq int, data []byte) *icmp.Message {
	var msgtype icmp.Type = ipv4.ICMPTypeEcho
	if isipv6 {
		msgtype = ipv6.ICMPTypeEchoRequest
	}
	body := &icmp.Echo{
		ID:   id,
		Seq:  seq,
		Data: data,
	}
	msg := &icmp.Message{
		Type: msgtype,
		Code: 0,
		Body: body,
	}
	return msg
}

func (this *IcmpPing) parserecvmsg(isipv6 bool, msg *icmp.Message) (data []byte, id, msgtype int) {
	id = 0
	data = nil
	msgtype = 0
	if isipv6 {
		switch msg.Type {
		case ipv6.ICMPTypeEchoReply:
			msgtype = 1
		case ipv6.ICMPTypeDestinationUnreachable:
			msgtype = 2
		case ipv6.ICMPTypeTimeExceeded:
			msgtype = 3
		}
	} else {
		switch msg.Type {
		case ipv4.ICMPTypeEchoReply:
			msgtype = 1
		case ipv4.ICMPTypeDestinationUnreachable:
			msgtype = 2
		case ipv4.ICMPTypeTimeExceeded:
			msgtype = 3
		}
	}
	switch msgtype {
	case 1:
		if tempmsg, ok := msg.Body.(*icmp.Echo); ok {
			data = tempmsg.Data
			id = tempmsg.ID
		}
	case 2:
		if tempmsg, ok := msg.Body.(*icmp.DstUnreach); ok {
			data = tempmsg.Data
		}
	case 3:
		if tempmsg, ok := msg.Body.(*icmp.TimeExceeded); ok {
			data = tempmsg.Data
		}
	}
	return
}

func (this *IcmpPing) errorResult(err error) IPingResult {
	r := &IcmpPingResult{}
	r.Err = err
	return r
}

var (
	_ IPing       = (*IcmpPing)(nil)
	_ IPingResult = (*IcmpPingResult)(nil)
)
