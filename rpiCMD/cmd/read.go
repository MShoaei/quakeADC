package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

type readOptions struct {
	file     *os.File
	ch       chan string
	skip     int
	duration int
}

// readCmd represents the read command
var adcReadCmd = &cobra.Command{
	Use:   "readAll",
	Short: "Start Reading from RPI_INTERFACE",
	// Long:  "Start Reading from RPI_INTERFACE",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		return read(readOptions{file: f, ch: nil, skip: 0})
	},
}

func read(opt readOptions) error {
	// dataFile := opt.file

	dev := os.Getenv("RPI_INTERFACE")
	if dev == "" {
		return fmt.Errorf("RPI_INTERFACE not set")
	}

	handle, err := pcap.OpenLive(dev, 256, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	ch := make(chan []byte, 100000)
	go func(ch chan<- []byte, d time.Duration) error {
		if opt.duration != 0 {
			t := time.NewTimer(d * time.Second)
			for {
				select {
				case <-t.C:
					close(ch)
					t.Stop()
					return nil
				default:
					var timeBuffer []byte
					packet, info, err := handle.ReadPacketData()
					if err != nil {
						return err
					}
					packet = append(info.Timestamp.AppendFormat(timeBuffer, "06-01-02,15:04:05.000000000"), packet...)
					ch <- packet
				}
			}
		} else {
			for {
				var timeBuffer []byte
				packet, info, err := handle.ReadPacketData()
				if err != nil {
					return err
				}
				packet = append(info.Timestamp.AppendFormat(timeBuffer, "06-01-02,15:04:05.000000000"), packet...)
				ch <- packet
			}
		}
	}(ch, time.Duration(opt.duration))

	if opt.ch != nil {
		skip := time.NewTicker(time.Duration(opt.skip) * time.Millisecond)
		b := strings.Builder{}
		b.Grow(100)
		for {
			var packet []byte
			var ok bool
			packet, ok = <-ch
			if !ok {
				close(opt.ch)
				dataFile.Close()
				return nil
			}
			adcNum := 1 // TODO: should change!

			t := packet[:27]

			ch0Value := (int32(packet[43]) << 16) + (int32(packet[44]) << 8) + (int32(packet[45]))

			ch1Value := (int32(packet[47]) << 16) + (int32(packet[48]) << 8) + int32((packet[49]))

			ch2Value := (int32(packet[51]) << 16) + (int32(packet[52]) << 8) + (int32(packet[53]))

			ch3Value := (int32(packet[55]) << 16) + (int32(packet[56]) << 8) + (int32(packet[57]))

			b.WriteString(fmt.Sprintf("%s %d %d %d %d %d\n", t, adcNum, ch0Value, ch1Value, ch2Value, ch3Value))
			fmt.Fprintf(dataFile, b.String())
			select {
			case <-skip.C:
				opt.ch <- b.String()
				b.Reset()
			default:
				b.Reset()

			}
		}
	} else {
		for {
			packet := <-ch

			adcNum := 1 // TODO: should change!

			t := packet[:27]

			// _, ch0Data := packet[42], packet[43:46]
			// ch0Header, ch0Data := packet[15], packet[16:19]

			// ch0Value := (packet[43] << 16) | (packet[44] << 8) | (packet[45])
			ch0Value := (int32(packet[43]) << 16) + (int32(packet[44]) << 8) + (int32(packet[45]))

			// _, ch1Data := packet[46], packet[47:50]
			ch1Value := (int32(packet[47]) << 16) + (int32(packet[48]) << 8) + int32((packet[49]))

			// _, ch2Data := packet[50], packet[51:54]
			ch2Value := (int32(packet[51]) << 16) + (int32(packet[52]) << 8) + (int32(packet[53]))

			// _, ch3Data := packet[54], packet[55:58]
			ch3Value := (int32(packet[55]) << 16) + (int32(packet[56]) << 8) + (int32(packet[57]))

			fmt.Fprintf(dataFile, "%s %d %d %d %d %d\n", t, adcNum, ch0Value, ch1Value, ch2Value, ch3Value)
		}
	}
}

func init() {
	rootCmd.AddCommand(adcReadCmd)
}
