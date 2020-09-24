package cmd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/wzv5/pping/pkg/ping"

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

func addTlsCommand() {
	var cmd = &cobra.Command{
		Use:   "tls <host> [ip]",
		Short: "tls ping",
		Long:  "tls ping",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runtls,
	}

	cmd.Flags().Uint16VarP(&tlsflag.tlsver, "tlsversion", "s", 0, "TLS version, one of 13, 12, 11, 10")
	cmd.Flags().DurationVarP(&tlsflag.conntime, "connection", "w", time.Second*4, "connection timeout")
	cmd.Flags().DurationVarP(&tlsflag.handtime, "handshake", "x", time.Second*10, "handshake timeout")
	cmd.Flags().Uint16VarP(&tlsflag.port, "port", "p", 443, "port")
	cmd.Flags().BoolVarP(&tlsflag.insecure, "insecure", "k", false, "allow insecure server connections")

	rootCmd.AddCommand(cmd)
}

func runtls(cmd *cobra.Command, args []string) error {
	host := args[0]
	var ip net.IP
	if len(args) == 2 {
		ip = net.ParseIP(args[1])
		if ip == nil {
			return errors.New("parse IP failed")
		}
	}

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
		return errors.New("unknown TLS version")
	}
	fmt.Printf("Ping %s (%d):\n", host, tlsflag.port)
	p := ping.NewTlsPing(host, tlsflag.port, tlsflag.conntime, tlsflag.handtime)
	p.IP = ip
	return RunPing(p)
}
