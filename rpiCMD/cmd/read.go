package cmd

import (
	"encoding/binary"
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
	ch := make(chan []byte, 1000000)
	epoch := time.Now()
	go func(ch chan<- []byte, d time.Duration) {
		if opt.duration != 0 {
			t := time.NewTimer(d * time.Millisecond)
			for {
				select {
				case <-t.C:
					close(ch)
					t.Stop()
					return
				default:
					packet, info, err := handle.ReadPacketData()
					if err != nil {
						return
					}

					// 8 byte time in milliseconds
					buf := make([]byte, 8)
					binary.LittleEndian.PutUint64(buf, uint64(info.Timestamp.Sub(epoch).Milliseconds()))
					packet = append(buf, packet...)
					ch <- packet
				}
			}
		} else {
			for {
				packet, info, err := handle.ReadPacketData()
				if err != nil {
					return
				}

				//8 byte time in milliseconds
				buf := make([]byte, 8)
				binary.LittleEndian.PutUint64(buf, uint64(info.Timestamp.Sub(epoch).Milliseconds()))
				packet = append(buf, packet...)
				ch <- packet
			}
		}
	}(ch, time.Duration(opt.duration))

	if opt.ch != nil {
		skip := time.NewTicker(time.Duration(opt.skip) * time.Millisecond)
		b := strings.Builder{}
		b.Grow(100)
		for {
			var signCH0 uint32 = 0
			var signCH1 uint32 = 0
			var signCH2 uint32 = 0
			var signCH3 uint32 = 0
			var packet []byte
			var ok bool

			packet, ok = <-ch
			if !ok {
				close(opt.ch)
				dataFile.Close()
				return nil
			}
			//adcNum := 1 // TODO: should change!

			_ = packet[:8]
			if int8(packet[24]) < 0 {
				signCH0 = 255 << 24
			}
			ch0Value := (uint32(packet[24]) << 16) + (uint32(packet[25]) << 8) + (uint32(packet[26]))

			if int8(packet[28]) < 0 {
				signCH1 = 255 << 24
			}
			ch1Value := (uint32(packet[28]) << 16) + (uint32(packet[29]) << 8) + uint32(packet[30])

			if int8(packet[32]) < 0 {
				signCH2 = 255 << 24
			}
			ch2Value := (uint32(packet[32]) << 16) + (uint32(packet[33]) << 8) + (uint32(packet[34]))

			if int8(packet[36]) < 0 {
				signCH3 = 255 << 24
			}
			ch3Value := (uint32(packet[36]) << 16) + (uint32(packet[37]) << 8) + (uint32(packet[38]))

			//b.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d\n", t, adcNum, ch0Value, ch1Value, ch2Value, ch3Value))
			b.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d\n",
				ch0Value, int32(ch0Value+signCH0),
				ch1Value, int32(ch1Value+signCH1),
				ch2Value, int32(ch2Value+signCH2),
				ch3Value, int32(ch3Value+signCH3)))

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
			var signCH0 uint32 = 0
			var signCH1 int32 = 0
			var signCH2 int32 = 0
			var signCH3 int32 = 0

			packet := <-ch

			_ = 1 // TODO: should change!

			_ = packet[:8]
			if int8(packet[24]) < 0 {
				signCH0 = 255 << 24
			}
			ch0Value := signCH0 + (uint32(packet[24]) << 16) + (uint32(packet[25]) << 8) + (uint32(packet[26]))

			if int8(packet[28]) < 0 {
				signCH1 = -1 << 24
			}
			_ = signCH1 + (int32(packet[28]) << 16) + (int32(packet[29]) << 8) + int32(packet[30])

			if int8(packet[32]) < 0 {
				signCH2 = -1 << 24
			}
			_ = signCH2 + (int32(packet[32]) << 16) + (int32(packet[33]) << 8) + (int32(packet[34]))

			if int8(packet[36]) < 0 {
				signCH3 = -1 << 24
			}
			_ = signCH3 + (int32(packet[36]) << 16) + (int32(packet[37]) << 8) + (int32(packet[38]))

			data := fmt.Sprintf("%+d\n", int32(ch0Value))
			fmt.Fprintf(dataFile, data)

			fmt.Println(data)
		}
	}
}

func init() {
	rootCmd.AddCommand(adcReadCmd)
}
