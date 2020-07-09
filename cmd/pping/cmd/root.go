package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/wzv5/pping/pkg/pping"

	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

type globalFlags struct {
	v    bool
	t    bool
	n    int
	i    time.Duration
	ipv4 bool
	ipv6 bool
}

var globalflag globalFlags

func init() {
	rootCmd = &cobra.Command{Use: filepath.Base(os.Args[0])}
	rootCmd.Version = "2020.07.09"
	rootCmd.PersistentFlags().BoolVarP(&globalflag.t, "infinite", "t", false, "ping the specified target until stopped")
	rootCmd.PersistentFlags().IntVarP(&globalflag.n, "count", "c", 4, "number of requests to send")
	rootCmd.PersistentFlags().DurationVarP(&globalflag.i, "interval", "i", time.Second*1, "delay between each request")
	rootCmd.PersistentFlags().BoolVarP(&globalflag.ipv4, "ipv4", "4", false, "use IPv4")
	rootCmd.PersistentFlags().BoolVarP(&globalflag.ipv6, "ipv6", "6", false, "use IPv6")

	rootCmd.PersistentPreRun = func(*cobra.Command, []string) {
		if globalflag.ipv4 && !globalflag.ipv6 {
			pping.LookupFunc = pping.LookupIPv4
		} else if !globalflag.ipv4 && globalflag.ipv6 {
			pping.LookupFunc = pping.LookupIPv6
		} else {
			pping.LookupFunc = pping.LookupIP
		}
	}

	addTcpCommand()
	addTlsCommand()
	addHttpCommand()
}

func Execute() error {
	return rootCmd.Execute()
}

func PingToChan(ctx context.Context, ping pping.IPing) <-chan pping.IPingResult {
	c := make(chan pping.IPingResult)
	go func() {
		c <- ping.PingContext(ctx)
	}()
	return c
}

func RunPing(ping pping.IPing) {
	// 预热，由于某些资源需要初始化，首次运行会耗时较长
	ping.Ping()

	resultlist := make([]pping.IPingResult, 0)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	for i := 1; i <= globalflag.n || globalflag.t; i++ {
		select {
		case result := <-PingToChan(ctx, ping):
			PrintResult(i, result)
			resultlist = append(resultlist, result)
		case <-c:
			goto end
		}

		// 最后一次 ping 结束后不再等待
		if i == globalflag.n && !globalflag.t {
			break
		}

		select {
		case <-c:
			goto end
		case <-time.After(globalflag.i):
		}
	}

end:
	cancel()
	PrintStatistics(resultlist)
}

func PrintResult(i int, r pping.IPingResult) {
	log.Printf("[%d] %v\n", i, r)
}

func PrintStatistics(r []pping.IPingResult) {
	if len(r) == 0 {
		return
	}
	var max, min, avg, a, ok, err int
	min = 9999
	for _, i := range r {
		if i.Error() != nil {
			err += 1
			continue
		}
		ok += 1
		t := i.Result()
		if t > max {
			max = t
		}
		if t < min {
			min = t
		}
		a += t
	}
	fmt.Println()
	fmt.Printf("\tsent = %d, ok = %d, failed = %d (%d%%)\n", len(r), ok, err, 100*err/len(r))
	if ok > 0 {
		avg = a / ok
		fmt.Printf("\tmin = %d ms, max = %d ms, avg = %d ms\n", min, max, avg)
	}
}
