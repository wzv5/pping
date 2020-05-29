package pping_test

import (
	"net"
	"testing"
	"time"

	"github.com/wzv5/pping/pkg/pping"
)

var HOST = "www.baidu.com"
var IP net.IP

func init() {
	ip, _ := net.LookupHost(HOST)
	IP = net.ParseIP(ip[0])
}

func TestTls(t *testing.T) {
	ping := pping.NewTlsPing(HOST, IP, 443, time.Second*1, time.Second*3, false, false)
	result := ping.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestTcp(t *testing.T) {
	ping := pping.NewTcpPing(IP, 80, time.Second*3)
	result := ping.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}
