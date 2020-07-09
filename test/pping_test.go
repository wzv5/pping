package pping_test

import (
	"testing"
	"time"

	"github.com/wzv5/pping/pkg/pping"
)

const HOST = "www.baidu.com"

func TestTls(t *testing.T) {
	ping := pping.NewTlsPing(HOST, 443, time.Second*1, time.Second*3)
	result := ping.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}

func TestTcp(t *testing.T) {
	ping := pping.NewTcpPing(HOST, 80, time.Second*3)
	result := ping.Ping()
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
}
