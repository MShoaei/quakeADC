package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/MShoaei/quakeADC/driver/xmega"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var dataFS afero.Fs
var memFS afero.Fs

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "launch a server to execute command",
	Long:  `launch a server which listens on port 9090 and executes commands.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOARCH == "arm" {
			if err := HardwareInitSeq(); err != nil {
				log.Fatalf("hardware init failed: %v", err)
			}
			log.Println("hardware init SUCCESSFUL")
		}

		wd, _ := os.Getwd()
		if err := os.MkdirAll(path.Join(wd, "data"), os.ModeDir|0755); err != nil {
			log.Fatalf("failed to create directory: %v", err)
		}
		dataFS = afero.NewBasePathFs(afero.NewOsFs(), path.Join(wd, "data"))
		memFS = afero.NewMemMapFs()
		_, _ = memFS.Create("/data.raw")

		api := NewAPI()
		port := "9090"
		if os.Getenv("PORT") != "" {
			port = os.Getenv("PORT")
		}
		_ = api.Run(":" + port)
	},
}

func HardwareInitSeq() error {
	xmega.ReadID(adcConnection.Connection())
	time.Sleep(100 * time.Millisecond)

	if err := xmega.ResetAllADC(adcConnection.Connection()); err != nil {
		return fmt.Errorf("failed to reset ADCs: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	//TODO: Multi logic analyzer is not implemented. The function below SHOULD be implemented.
	list, err := xmega.DetectLogicConnString(adcConnection.Connection())
	if err != nil {
		return fmt.Errorf("failed to detect logic analyzers conn string: %v", err)
	}
	driverConnDigits = list
	time.Sleep(100 * time.Millisecond)

	if err := xmega.EnableMCLK(adcConnection.Connection()); err != nil {
		return fmt.Errorf("failed to enable MCLK: %v", err)
	}

	if err := xmega.StatusLED(adcConnection.Connection(), xmega.On); err != nil {
		return fmt.Errorf("failed to turn on LED: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	return nil
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
