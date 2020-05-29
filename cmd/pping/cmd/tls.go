package cmd

import (
	"fmt"
	"net"
	"time"

	"github.com/wzv5/pping/pkg/pping"

	"github.com/spf13/cobra"
)

type tlsFlags struct {
	tls13    bool
	conntime time.Duration
	handtime time.Duration
	port     uint16
	insecure bool
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

	cmd.Flags().BoolVarP(&tlsflag.tls13, "tls13", "s", false, "force use TLS 1.3")
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
	fmt.Printf("Ping %s (%s):\n", host, ip)
	ping := pping.NewTlsPing(host, net.ParseIP(ip), tlsflag.port, tlsflag.conntime, tlsflag.handtime, tlsflag.tls13, tlsflag.insecure)
	generalPing(ping)
}
