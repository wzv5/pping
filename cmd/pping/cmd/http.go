package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/wzv5/pping/pkg/ping"

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
		RunE:  runhttp,
	}

	cmd.Flags().DurationVarP(&httpflag.timeout, "timeout", "w", time.Second*4, "timeout")
	cmd.Flags().StringVarP(&httpflag.method, "method", "m", "GET", "method")
	cmd.Flags().BoolVarP(&httpflag.disablehttp2, "nohttp2", "d", false, "disable HTTP/2")
	cmd.Flags().BoolVarP(&httpflag.disablecompression, "nocompression", "x", false, "disable compression")
	cmd.Flags().BoolVarP(&httpflag.insecure, "insecure", "k", false, "allow insecure server connections when using SSL")
	cmd.Flags().StringVarP(&httpflag.refer, "referrer", "r", "", "Referer header")
	cmd.Flags().StringVarP(&httpflag.ua, "useragent", "u", "", "User-Agent header")
	rootCmd.AddCommand(cmd)
}

func runhttp(cmd *cobra.Command, args []string) error {
	url := args[0]
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	var ip net.IP
	if len(args) == 2 {
		ip = net.ParseIP(args[1])
		if ip == nil {
			return errors.New("parse IP failed")
		}
	}
	fmt.Printf("Ping %s:\n", url)
	p := ping.NewHttpPing(httpflag.method, url, httpflag.timeout)
	p.DisableHttp2 = httpflag.disablehttp2
	p.DisableCompression = httpflag.disablecompression
	p.Insecure = httpflag.insecure
	p.Referrer = httpflag.refer
	p.UserAgent = httpflag.ua
	p.IP = ip
	return RunPing(p)
}
