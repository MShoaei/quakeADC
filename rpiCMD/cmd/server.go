package cmd

import (
	"log"
	"os"
	"path"
	"runtime"

	"github.com/MShoaei/quakeADC/server"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "launch a server to execute command",
	Long:  `launch a server which listens on port 9090 and executes commands.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		wd, _ := os.Getwd()
		if err := os.MkdirAll(path.Join(wd, "data"), os.ModeDir|0755); err != nil {
			log.Fatalf("failed to create data directory: %v", err)
		}

		dataFS := afero.NewBasePathFs(afero.NewOsFs(), path.Join(wd, "data"))
		memFS := afero.NewMemMapFs()

		s := server.NewServer(dataFS, memFS, adcConnection, debug)
		if runtime.GOARCH == "arm" {
			if err := s.HardwareInitSeq(); err != nil {
				log.Fatalf("hardware init failed: %v", err)
			}
			log.Println("hardware init SUCCESSFUL")
		}

		_ = s.Run()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
