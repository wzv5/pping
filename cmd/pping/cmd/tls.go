package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/wzv5/pping/pkg/pping"

	"github.com/spf13/cobra"
)

type tlsFlags struct {
	conntime time.Duration
	handtime time.Duration
	port     uint16
	insecure bool
	tlsver   uint16
}

var tlsflag tlsFlags

func AddTlsCommand() {
	var cmd = &cobra.Command{
		Use:   "tls <host> [ip]",
		Short: "tls ping",
		Long:  "tls ping",
		Args:  cobra.RangeArgs(1, 2),
		Run:   runtls,
	}

	cmd.Flags().Uint16VarP(&tlsflag.tlsver, "tlsversion", "s", 0, "TLS version, one of 13, 12, 11, 10")
	cmd.Flags().DurationVarP(&tlsflag.conntime, "connection", "w", time.Second*3, "connection timeout")
	cmd.Flags().DurationVarP(&tlsflag.handtime, "handshake", "x", time.Second*10, "handshake timeout")
	cmd.Flags().Uint16VarP(&tlsflag.port, "port", "p", 443, "port")
	cmd.Flags().BoolVarP(&tlsflag.insecure, "insecure", "k", false, "allow insecure server connections")

	rootCmd.AddCommand(cmd)
}

func runtls(cmd *cobra.Command, args []string) {
	host := args[0]
	ip := host
	if len(args) == 2 {
		ip = args[1]
	}
	addr, err := net.LookupHost(ip)
	if err != nil {
		fmt.Println(err)
		return
	}
	ip = addr[0]
	switch tlsflag.tlsver {
	case 0:
	case 13:
		tlsflag.tlsver = tls.VersionTLS13
	case 12:
		tlsflag.tlsver = tls.VersionTLS12
	case 11:
		tlsflag.tlsver = tls.VersionTLS11
	case 10:
		tlsflag.tlsver = tls.VersionTLS10
	default:
		fmt.Println("unknown TLS version")
		return
	}
	fmt.Printf("Ping %s (%s):\n", host, ip)
	ping := pping.NewTlsPing(host, net.ParseIP(ip), tlsflag.port, tlsflag.conntime, tlsflag.handtime, tlsflag.tlsver, tlsflag.insecure)
	generalPing(ping)
}
