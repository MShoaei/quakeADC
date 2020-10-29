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
	"gobot.io/x/gobot/drivers/spi"
)

var dataFS afero.Fs
var memFS afero.Fs

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "launch a server to execute command",
	Long:  `launch a server which listens on port 9090 and executes commands.`,
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
		_ = api.Listen(":" + port)
	},
}

func HardwareInitSeq() error {
	//if err := xmega.Reset(); err != nil {
	//	return fmt.Errorf("reset failed: %v", err)
	//}
	//time.Sleep(2 * time.Second)

	conn, err := spi.GetSpiConnection(0, 0, 0, 8, 50000)
	if err != nil {
		return fmt.Errorf("failed to create spi connection: %v", err)
	}
	defer conn.Close()

	xmega.ReadID(conn)
	time.Sleep(100 * time.Millisecond)

	if err := xmega.StatusLED(conn, xmega.On); err != nil {
		return fmt.Errorf("failed to turn on LED: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := xmega.ResetAllADC(conn); err != nil {
		return fmt.Errorf("failed to reset ADCs: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	//TODO: Multi logic analyzer is not implemented. The function below SHOULD be implemented.
	//xmega.DetectLogicConnString(conn)
	//time.Sleep(100 * time.Millisecond)

	if err := xmega.EnableMCLK(conn); err != nil {
		return fmt.Errorf("failed to enable MCLK: %v", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
