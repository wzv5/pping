package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wzv5/pping/pkg/pping"

	"github.com/spf13/cobra"
)

type tcpFlags struct {
	port    uint16
	timeout time.Duration
}

var tcpflag tcpFlags

func addTcpCommand() {
	var cmd = &cobra.Command{
		Use:   "tcp <host> <port>",
		Short: "tcp ping",
		Long:  "tcp ping",
		Args:  cobra.ExactArgs(2),
		Run:   runtcp,
	}

	cmd.Flags().DurationVarP(&tcpflag.timeout, "timeout", "w", time.Second*3, "timeout")
	rootCmd.AddCommand(cmd)
}

func runtcp(cmd *cobra.Command, args []string) {
	host := args[0]
	ip, err := pping.LookupFunc(host)
	if err != nil {
		fmt.Println(err)
		return
	}
	port, err := strconv.ParseUint(args[1], 10, 16)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Ping %s (%d):\n", host, port)
	ping := pping.NewTcpPing(host, uint16(port), tcpflag.timeout)
	ping.IP = ip
	RunPing(ping)
}
