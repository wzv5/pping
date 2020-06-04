package cmd

import (
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
	v bool
	t bool
	n int
	i time.Duration
}

var globalflag globalFlags

func generalPing(ping pping.IPing) {
	// 预热，由于某些资源需要初始化，首次运行会耗时较长
	ping.Ping()

	resultlist := make([]pping.IPingResult, 0)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for i := 1; i <= globalflag.n || globalflag.t; i++ {
		result := ping.Ping()
		PrintResult(i, result)
		resultlist = append(resultlist, result)
		// 最后一次 ping 结束后不再等待
		if i == globalflag.n && !globalflag.t {
			break
		}
		// 可被 ctrl + c 中止的等待
		if WaitSignal(c, globalflag.i) {
			break
		}
	}

	PrintSummary(resultlist)
}

func init() {
	rootCmd = &cobra.Command{Use: filepath.Base(os.Args[0])}
	rootCmd.Version = "2020.05.29"
	rootCmd.PersistentFlags().BoolVarP(&globalflag.t, "infinite", "t", false, "ping the specified target until stopped")
	rootCmd.PersistentFlags().IntVarP(&globalflag.n, "count", "c", 4, "number of requests to send")
	rootCmd.PersistentFlags().DurationVarP(&globalflag.i, "interval", "i", time.Second*1, "delay between each request")

	AddTcpCommand()
	AddTlsCommand()
	AddHttpCommand()
}

func Execute() error {
	return rootCmd.Execute()
}

func PrintResult(i int, r pping.IPingResult) {
	log.Printf("[%d] %v\n", i, r)
}

func PrintSummary(r []pping.IPingResult) {
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

// true：收到信号，false：超时
func WaitSignal(c chan os.Signal, timeout time.Duration) bool {
	select {
	case <-c:
		return true
	case <-time.After(timeout):
	}
	return false
}
