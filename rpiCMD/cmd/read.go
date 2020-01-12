package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

// readCmd represents the read command
var adcReadCmd = &cobra.Command{
	Use:   "readAll",
	Short: "Start Reading from RPI_INTERFACE",
	// Long:  "Start Reading from RPI_INTERFACE",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataFile, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		dev := os.Getenv("RPI_INTERFACE")
		if dev == "" {
			return fmt.Errorf("RPI_INTERFACE not set")
		}

		handle, err := pcap.OpenLive(dev, 256, true, pcap.BlockForever)
		if err != nil {
			return err
		}

		for {
			packet, _, err := handle.ReadPacketData()
			if err != nil {
				return err
			}
			// packet = append(packet, '\n')
			// dataFile.Write(packet)

			adcNum := 1 // TODO: should change!

			ch0Header, ch0Data := packet[15], packet[16:19]
			ch1Header, ch1Data := packet[19], packet[20:23]
			ch2Header, ch2Data := packet[23], packet[24:27]
			ch3Header, ch3Data := packet[27], packet[28:31]

			fmt.Fprintf(dataFile, "%s %d %d %d %d %d %d %d %d %d\n", time.Now().Format(time.StampMilli), adcNum, ch0Header, ch0Data, ch1Header, ch1Data, ch2Header, ch2Data, ch3Header, ch3Data)
		}
		// return nil
	},
}

func init() {
	rootCmd.AddCommand(adcReadCmd)
}
