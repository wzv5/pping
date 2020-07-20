package pping_test

import (
	"testing"
	"time"

	"github.com/wzv5/pping/pkg/ping"
)

const HOST = "www.baidu.com"

func TestTls(t *testing.T) {
	p := ping.NewTlsPing(HOST, 443, time.Second*1, time.Second*3)
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestTcp(t *testing.T) {
	p := ping.NewTcpPing(HOST, 80, time.Second*3)
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func BenchmarkIcmp(b *testing.B) {
	p := ping.NewIcmpPing("127.0.0.1", time.Second*1)
	p.Privileged = true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := p.Ping()
		if r.Error() != nil {
			b.Fatal(r.Error())
		}
	}
}
