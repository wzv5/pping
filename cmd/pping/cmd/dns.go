package cmd

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/wzv5/pping/pkg/ping"

	"github.com/spf13/cobra"
)

type dnsFlags struct {
	port     uint16
	timeout  time.Duration
	tcp      bool
	tls      bool
	qtype    string
	domain   string
	insecure bool
}

var dnsflag dnsFlags

func addDnsCommand() {
	var cmd = &cobra.Command{
		Use:   "dns <host>",
		Short: "dns ping",
		Long:  "dns ping",
		Args:  cobra.ExactArgs(1),
		RunE:  rundns,
	}

	cmd.Flags().DurationVarP(&dnsflag.timeout, "timeout", "w", time.Second*4, "timeout")
	cmd.Flags().Uint16VarP(&dnsflag.port, "port", "p", 0, "port")
	cmd.Flags().BoolVar(&dnsflag.tcp, "tcp", false, "use TCP")
	cmd.Flags().BoolVar(&dnsflag.tls, "tls", false, "use DNS-over-TLS")
	cmd.Flags().StringVar(&dnsflag.qtype, "type", "NS", "A, AAAA, NS, ...")
	cmd.Flags().StringVar(&dnsflag.domain, "domain", ".", "domain")
	cmd.Flags().BoolVarP(&dnsflag.insecure, "insecure", "k", false, "allow insecure server connections")

	rootCmd.AddCommand(cmd)
}

func rundns(cmd *cobra.Command, args []string) error {
	host := args[0]
	Net := "udp"
	if dnsflag.tls {
		Net = "tcp-tls"
	} else if dnsflag.tcp {
		Net = "tcp"
	}
	if dnsflag.port == 0 {
		switch Net {
		case "udp", "tcp":
			dnsflag.port = 53
		case "tcp-tls":
			dnsflag.port = 853
		}
	}
	fmt.Printf("Ping %s://%s:\n", Net, net.JoinHostPort(host, strconv.Itoa(int(dnsflag.port))))
	p := ping.NewDnsPing(host, dnsflag.timeout)
	p.Port = dnsflag.port
	p.Net = Net
	p.Type = dnsflag.qtype
	p.Domain = dnsflag.domain
	p.Insecure = dnsflag.insecure
	return RunPing(p)
}
