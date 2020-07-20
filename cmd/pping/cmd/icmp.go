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
}

var icmpflag icmpFlags

func addIcmpCommand() {
	var cmd = &cobra.Command{
		Use:   "icmp <host>",
		Short: "icmp ping",
		Long:  "icmp ping",
		Args:  cobra.ExactArgs(1),
		Run:   runicmp,
	}

	cmd.Flags().DurationVarP(&icmpflag.timeout, "timeout", "w", time.Second*3, "timeout")
	cmd.Flags().BoolVarP(&icmpflag.privileged, "privileged", "p", false, "privileged")
	rootCmd.AddCommand(cmd)
}

func runicmp(cmd *cobra.Command, args []string) {
	host := args[0]
	fmt.Printf("Ping %s:\n", host)
	p := ping.NewIcmpPing(host, icmpflag.timeout)
	p.Privileged = icmpflag.privileged
	RunPing(p)
}
