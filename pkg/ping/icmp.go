package ping

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
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

type icmpconn interface {
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
	tracker := r.Int63n(math.MaxInt64)
	id := os.Getpid() & 0xffff
	msg := this.getmsg(isipv6, tracker, id, 0, 64)
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return this.errorResult(err)
	}
	var dst net.Addr = &net.IPAddr{IP: ip}
	if network == "udp" {
		dst = &net.UDPAddr{IP: ip}
	}
	for {
		if _, err := conn.WriteTo(msgBytes, dst); err != nil {
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
		if isipv6 {
			var cm *ipv6.ControlMessage
			recvSize, cm, _, err = conn.IPv6PacketConn().ReadFrom(recvBytes)
			if cm != nil {
				ttl = cm.HopLimit
			}
		} else {
			var cm *ipv4.ControlMessage
			recvSize, cm, _, err = conn.IPv4PacketConn().ReadFrom(recvBytes)
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
		if recvMsg.Type != ipv4.ICMPTypeEchoReply && recvMsg.Type != ipv6.ICMPTypeEchoReply {
			// 不是 echo 回复，忽略并继续接收
			//fmt.Println("不是 echo 回复")
			continue
		}
		if echomsg, ok := recvMsg.Body.(*icmp.Echo); ok {
			// 收到 echo 回复
			if network == "ip" {
				// 如果 ID 不匹配则继续接收
				if echomsg.ID != id {
					//fmt.Println("ID 不匹配")
					continue
				}
			}
			if len(echomsg.Data) < 8+8 {
				//fmt.Println("接收数据过小")
				continue
			}
			recvTracker := bytesToInt(echomsg.Data[8:])
			timestamp := bytesToTime(echomsg.Data[:8])
			if recvTracker != tracker {
				//fmt.Println("tracker 不匹配")
				continue
			}
			return &IcmpPingResult{
				TTL:  ttl,
				Time: int(recvAt.Sub(timestamp).Milliseconds()),
				IP:   ip,
			}
		} else {
			return this.errorResult(errors.New("invalid ICMP echo reply"))
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

func (this *IcmpPing) getmsg(isipv6 bool, tracker int64, id, seq, msgsize int) *icmp.Message {
	var msgtype icmp.Type = ipv4.ICMPTypeEcho
	if isipv6 {
		msgtype = ipv6.ICMPTypeEchoRequest
	}
	t := append(timeToBytes(time.Now()), intToBytes(tracker)...)
	if remainsize := msgsize - 8 - 8; remainsize > 0 {
		t = append(t, bytes.Repeat([]byte{1}, remainsize)...)
	}
	body := &icmp.Echo{
		ID:   id,
		Seq:  seq,
		Data: t,
	}
	msg := &icmp.Message{
		Type: msgtype,
		Code: 0,
		Body: body,
	}
	return msg
}

func (this *IcmpPing) errorResult(err error) IPingResult {
	r := &IcmpPingResult{}
	r.Err = err
	return r
}

func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func bytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func intToBytes(tracker int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(tracker))
	return b
}

var (
	_ IPing       = (*IcmpPing)(nil)
	_ IPingResult = (*IcmpPingResult)(nil)
)
