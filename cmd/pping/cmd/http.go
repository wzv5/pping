package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/wzv5/pping/pkg/pping"

	"github.com/spf13/cobra"
)

type httpFlags struct {
	method             string
	disablehttp2       bool
	disablecompression bool
	insecure           bool
	timeout            time.Duration
	refer              string
	ua                 string
}

var httpflag httpFlags

func addHttpCommand() {
	var cmd = &cobra.Command{
		Use:   "http <url> [ip]",
		Short: "http ping",
		Long:  "http ping",
		Args:  cobra.RangeArgs(1, 2),
		Run:   runhttp,
	}

	cmd.Flags().DurationVarP(&httpflag.timeout, "timeout", "w", time.Second*3, "timeout")
	cmd.Flags().StringVarP(&httpflag.method, "method", "m", "GET", "method")
	cmd.Flags().BoolVarP(&httpflag.disablehttp2, "nohttp2", "d", false, "disable HTTP/2")
	cmd.Flags().BoolVarP(&httpflag.disablecompression, "nocompression", "x", false, "disable compression")
	cmd.Flags().BoolVarP(&httpflag.insecure, "insecure", "k", false, "allow insecure server connections when using SSL")
	cmd.Flags().StringVarP(&httpflag.refer, "referrer", "r", "", "Referer header")
	cmd.Flags().StringVarP(&httpflag.ua, "useragent", "u", "", "User-Agent header")
	rootCmd.AddCommand(cmd)
}

func runhttp(cmd *cobra.Command, args []string) {
	url := args[0]
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	var ip net.IP
	if len(args) == 2 {
		ip = net.ParseIP(args[1])
		if ip == nil {
			fmt.Println("parse IP failed")
			return
		}
	}
	fmt.Printf("Ping %s:\n", url)
	ping := pping.NewHttpPing(httpflag.method, url, httpflag.timeout)
	ping.DisableHttp2 = httpflag.disablehttp2
	ping.DisableCompression = httpflag.disablecompression
	ping.Insecure = httpflag.insecure
	ping.Referrer = httpflag.refer
	ping.UserAgent = httpflag.ua
	ping.IP = ip
	RunPing(ping)
}
