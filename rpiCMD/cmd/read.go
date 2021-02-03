package cmd

import (
	"os"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/spf13/cobra"
)

//var input io.ReadCloser
//var buffer []byte
var driverConnDigits []string

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "command for operation",
}

func newMonitorLiveCommand() *cobra.Command {
	options := struct {
		sample int
	}{}
	cmd := &cobra.Command{
		Use: "monitor",
		Run: func(cmd *cobra.Command, args []string) {
			f, _ := os.Create("test.raw")
			driver.MonitorLive(f, options.sample)
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.IntVar(&options.sample, "sample", 0, "")
	_ = cmd.MarkFlagRequired("sample")

	return cmd
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.AddCommand(newMonitorLiveCommand())
}
