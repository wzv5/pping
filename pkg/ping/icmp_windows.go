//go:build windows

package ping

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"syscall"
	"time"
	"unsafe"
)

type ip_option_information struct {
	ttl        uint8
	tos        uint8
	flags      uint8
	optionsize uint8
	optiondata uintptr
}

type icmpv6_echo_reply struct {
	address       ipv6_address_ex
	status        uint32
	roundtriptime uint32
}

type icmp_echo_reply struct {
	address       [4]byte
	status        uint32
	roundtriptime uint32
	datasize      uint16
	reserved      uint16
	data          uintptr
	option        ip_option_information
}

type ipv6_address_ex struct {
	// go 语言不支持手动指定内存对齐方式，sin6_flowinfo 字段错位
	data [26]byte
	// sin6_port     uint16
	// sin6_flowinfo uint32
	// sin6_addr     [16]byte
	// sin6_scope_id uint32
}

var (
	iphlpapi = syscall.MustLoadDLL("iphlpapi.dll")

	icmp6CreateFile = iphlpapi.MustFindProc("Icmp6CreateFile")
	icmp6SendEcho2  = iphlpapi.MustFindProc("Icmp6SendEcho2")

	icmpCreateFile = iphlpapi.MustFindProc("IcmpCreateFile")
	icmpSendEcho2  = iphlpapi.MustFindProc("IcmpSendEcho2")

	icmpCloseHandle = iphlpapi.MustFindProc("IcmpCloseHandle")
)

func (this *IcmpPing) ping_rootless(ctx context.Context) IPingResult {
	ip, isipv6, err := this.parseip()
	if err != nil {
		return this.errorResult(err)
	}
	handle := syscall.InvalidHandle
	defer func() {
		if handle != syscall.InvalidHandle {
			IcmpCloseHandle(handle)
		}
	}()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if isipv6 {
		handle = Icmp6CreateFile()
		if handle == syscall.InvalidHandle {
			return this.errorResult(errors.New("IcmpCreateFile failed"))
		}
		data := make([]byte, this.Size)
		r.Read(data)
		recv := Icmp6SendEcho(handle, ip, data, this.Timeout, this.TTL)
		if recv == nil {
			return this.errorResult(errors.New("IcmpSendEcho failed"))
		}
		recvmsg := (*icmpv6_echo_reply)(unsafe.Pointer(&recv[0]))
		var ip net.IP = recvmsg.address.data[6:22]
		if recvmsg.status != 0 {
			return this.errorResult(fmt.Errorf("%s: %s", ip.String(), icmpStatusToString(recvmsg.status)))
		}
		return &IcmpPingResult{
			Time: int(recvmsg.roundtriptime),
			TTL:  -1,
			IP:   ip,
		}
	} else {
		handle = IcmpCreateFile()
		if handle == syscall.InvalidHandle {
			return this.errorResult(errors.New("IcmpCreateFile failed"))
		}
		data := make([]byte, this.Size)
		r.Read(data)
		recv := IcmpSendEcho(handle, ip, data, this.Timeout, this.TTL)
		if recv == nil {
			return this.errorResult(errors.New("IcmpSendEcho failed"))
		}
		recvmsg := (*icmp_echo_reply)(unsafe.Pointer(&recv[0]))
		var ip net.IP = recvmsg.address[:]
		if recvmsg.status != 0 {
			return this.errorResult(fmt.Errorf("%s: %s", ip.String(), icmpStatusToString(recvmsg.status)))
		}
		return &IcmpPingResult{
			Time: int(recvmsg.roundtriptime),
			TTL:  int(recvmsg.option.ttl),
			IP:   ip,
		}
	}
}

func ipv4ToInt(ip net.IP) uint32 {
	return binary.LittleEndian.Uint32(ip.To4())
}

func IcmpCreateFile() syscall.Handle {
	h, _, _ := icmpCreateFile.Call()
	return syscall.Handle(h)
}

func IcmpCloseHandle(h syscall.Handle) uintptr {
	ret, _, _ := icmpCloseHandle.Call(uintptr(h))
	return ret
}

func IcmpSendEcho(handle syscall.Handle, ip net.IP, data []byte, timeout time.Duration, ttl int) []byte {
	buf := make([]byte, (int)(unsafe.Sizeof(icmp_echo_reply{}))+len(data))
	var pOptions *ip_option_information
	if ttl > 0 {
		pOptions = &ip_option_information{
			ttl: uint8(ttl),
		}
	}
	n, _, _ := icmpSendEcho2.Call(
		uintptr(handle),                   // icmphandle
		0,                                 // event
		0,                                 // apcroutine
		0,                                 // apccontext
		uintptr(ipv4ToInt(ip)),            // destinationaddress
		uintptr(unsafe.Pointer(&data[0])), // requestdata
		uintptr(len(data)),                // requestsize
		uintptr(unsafe.Pointer(pOptions)), // requestoptions
		uintptr(unsafe.Pointer(&buf[0])),  // replaybuffer
		uintptr(len(buf)),                 // replysize
		uintptr(timeout.Milliseconds()),   // timeout
	)
	if n == 0 {
		return nil
	}
	return buf
}

func Icmp6CreateFile() syscall.Handle {
	h, _, _ := icmp6CreateFile.Call()
	return syscall.Handle(h)
}

func Icmp6SendEcho(handle syscall.Handle, ip net.IP, data []byte, timeout time.Duration, ttl int) []byte {
	ip6source := syscall.RawSockaddrInet6{
		Family: syscall.AF_INET6,
	}
	ip6dest := syscall.RawSockaddrInet6{
		Family: syscall.AF_INET6,
	}
	copy(ip6dest.Addr[:], ip)
	buf := make([]byte, (int)(unsafe.Sizeof(icmpv6_echo_reply{}))+len(data))
	var pOptions *ip_option_information
	if ttl > 0 {
		pOptions = &ip_option_information{
			ttl: uint8(ttl),
		}
	}
	n, _, _ := icmp6SendEcho2.Call(
		uintptr(handle),                     // icmphandle
		0,                                   // event
		0,                                   // apcroutine
		0,                                   // apccontext
		uintptr(unsafe.Pointer(&ip6source)), // sourceaddress
		uintptr(unsafe.Pointer(&ip6dest)),   // destinationaddress
		uintptr(unsafe.Pointer(&data[0])),   // requestdata
		uintptr(len(data)),                  // requestsize
		uintptr(unsafe.Pointer(pOptions)),   // requestoptions
		uintptr(unsafe.Pointer(&buf[0])),    // replaybuffer
		uintptr(len(buf)),                   // replysize
		uintptr(timeout.Milliseconds()),     // timeout
	)
	if n == 0 {
		return nil
	}
	return buf
}

func icmpStatusToString(status uint32) string {
	switch status {
	case 11002:
		return "destination network was unreachable"
	case 11003:
		return "destination host was unreachable"
	case 11010:
		return "request timed out"
	case 11013:
		return "time exceeded"
	}
	return fmt.Sprintf("unknown error (%d)", status)
}
