package pping_test

import (
	"strings"
	"testing"
	"time"

	"github.com/wzv5/pping/pkg/ping"
)

const HOST = "www.microsoft.com"

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

func TestIcmp(t *testing.T) {
	p := ping.NewIcmpPing("127.0.0.1", time.Second*1)
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestIcmp6(t *testing.T) {
	p := ping.NewIcmpPing("::1", time.Second*1)
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestIcmp_ttl(t *testing.T) {
	p := ping.NewIcmpPing(HOST, time.Second*1)
	p.TTL = 1
	result := p.Ping()
	if !strings.Contains(result.Error().Error(), "exceeded") {
		t.Fatal(result.Error())
	}
}

func TestIcmp_ttl_root(t *testing.T) {
	p := ping.NewIcmpPing(HOST, time.Second*1)
	p.Privileged = true
	p.TTL = 1
	result := p.Ping()
	if !strings.Contains(result.Error().Error(), "exceeded") {
		t.Fatal(result.Error())
	}
}

func TestIcmp_size(t *testing.T) {
	p := ping.NewIcmpPing(HOST, time.Second*1)
	p.Size = 60000
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestIcmp_size_root(t *testing.T) {
	p := ping.NewIcmpPing(HOST, time.Second*1)
	p.Privileged = true
	p.Size = 60000
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestQuic(t *testing.T) {
	p := ping.NewQuicPing("quic.nginx.org", 443, time.Second*5)
	result := p.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}
