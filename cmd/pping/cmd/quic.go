package cmd

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/spf13/cobra"
	"github.com/wzv5/pping/pkg/ping"
)

type quicFlags struct {
	timeout  time.Duration
	port     uint16
	insecure bool
	alpn     string
}

var quicflag quicFlags

func addQuicCommand() {
	var cmd = &cobra.Command{
		Use:   "quic <host> [ip]",
		Short: "quic ping",
		Long:  "quic ping",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runquic,
	}

	cmd.Flags().DurationVarP(&quicflag.timeout, "timeout", "w", time.Second*4, "timeout")
	cmd.Flags().Uint16VarP(&quicflag.port, "port", "p", 443, "port")
	cmd.Flags().BoolVarP(&quicflag.insecure, "insecure", "k", false, "allow insecure server connections")
	cmd.Flags().StringVarP(&quicflag.alpn, "alpn", "a", http3.NextProtoH3, "ALPN")
	rootCmd.AddCommand(cmd)
}

func runquic(cmd *cobra.Command, args []string) error {
	host := args[0]
	var ip net.IP
	if len(args) == 2 {
		ip = net.ParseIP(args[1])
		if ip == nil {
			return errors.New("parse IP failed")
		}
	}

	fmt.Printf("Ping %s (%d):\n", host, quicflag.port)
	p := ping.NewQuicPing(host, quicflag.port, quicflag.timeout)
	p.Insecure = quicflag.insecure
	p.ALPN = quicflag.alpn
	p.IP = ip
	return RunPing(p)
}
