package cmd

import (
	"fmt"
	"time"

	"github.com/wzv5/pping/pkg/ping"

	"github.com/spf13/cobra"
)

type icmpFlags struct {
	privileged bool
	timeout    time.Duration
	ttl        int
	size       int
}

var icmpflag icmpFlags

func addIcmpCommand() {
	var cmd = &cobra.Command{
		Use:   "icmp <host>",
		Short: "icmp ping",
		Long:  "icmp ping",
		Args:  cobra.ExactArgs(1),
		RunE:  runicmp,
	}

	cmd.Flags().DurationVarP(&icmpflag.timeout, "timeout", "w", time.Second*4, "timeout")
	cmd.Flags().BoolVarP(&icmpflag.privileged, "privileged", "p", false, "privileged")
	cmd.Flags().IntVarP(&icmpflag.ttl, "ttl", "l", 0, "time to live")
	cmd.Flags().IntVarP(&icmpflag.size, "size", "s", 0, "send buffer size")
	rootCmd.AddCommand(cmd)
}

func runicmp(cmd *cobra.Command, args []string) error {
	host := args[0]
	fmt.Printf("Ping %s:\n", host)
	p := ping.NewIcmpPing(host, icmpflag.timeout)
	p.Privileged = icmpflag.privileged
	if icmpflag.ttl > 0 {
		p.TTL = icmpflag.ttl
	}
	if icmpflag.size > 0 {
		p.Size = icmpflag.size
	}
	return RunPing(p)
}
