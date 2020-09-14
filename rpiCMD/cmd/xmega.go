package cmd

import (
	"fmt"
	"strconv"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/spf13/cobra"
	"gobot.io/x/gobot/drivers/spi"
)

// xmegaCmd represents the adc command
var xmegaCmd = &cobra.Command{
	Use:   "xmega",
	Short: "command to control XMega over SPI",
	Long:  "command to control XMega over SPI",
	RunE: func(cmd *cobra.Command, args []string) error {
		h, _ := strconv.ParseUint(args[0], 16, 8)
		l, _ := strconv.ParseUint(args[1], 16, 16)

		tx := []byte{uint8(h), uint8(l), 0}
		rx := make([]byte, 3)

		conn, err := spi.GetSpiConnection(0, 0, 0, 8, 50000)
		if err != nil {
			return nil
		}
		defer conn.Close()

		_ = driver.EnableChipSelect(0)
		if err := conn.Tx(tx, rx); err != nil {
			return err
		}
		_ = driver.DisableChipSelect(0)
		fmt.Println(tx, rx)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(xmegaCmd)
}
