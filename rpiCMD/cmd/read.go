package cmd

import (
	"bufio"
	"encoding/binary"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

var defaultBuilder = strings.Builder{}
var defaultWriter = bufio.NewWriterSize(nil, 524288000)

//var defaultChannel = make(chan []byte, 1000000)

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
		f, err := os.OpenFile("access", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		return read(readOptions{file: f, ch: nil, skip: 0})
	},
}

func read(opt readOptions) error {
	ch := make(chan []byte, 1000000)

	defaultWriter.Reset(opt.file)
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		for sig := range interruptChan {
			log.Println(sig)
			err := defaultWriter.Flush()
			if err != nil {
				log.Println("interrupt and flush failed with error: ", err)
			}
			os.Exit(1)
		}
	}()

	counter := 0
	go getPackets(ch, time.Duration(opt.duration))
	for {
		var packet []byte
		var ok bool

		packet, ok = <-ch
		if !ok {
			close(opt.ch)
			err := defaultWriter.Flush()
			if err != nil {
				log.Println("flush failed with error: ", err)
			}
			dataFile.Close()
			return nil
		}

		err := getPacketData(packet)
		if err != nil {
			log.Println("failed to read data from packet: ", err)
			return err
		}
		defaultWriter.WriteString(defaultBuilder.String())
		if counter%opt.skip == 0 {
			opt.ch <- defaultBuilder.String()
			defaultBuilder.Reset()
			counter = 0
		}
		counter++
		//if skip != nil {
		//	select {
		//	case <-skip.C:
		//		opt.ch <- defaultBuilder.String()
		//		defaultBuilder.Reset()
		//	default:
		//		continue
		//	}
		//}
	}
}

func getPackets(ch chan<- []byte, d time.Duration) {
	if d != 0 {
		getWithTicker(ch, d)
	} else {
		getNonStop(ch)
	}
}

// getWithTicker reads packets from handle for certain Duration d.
func getWithTicker(ch chan<- []byte, d time.Duration) {
	dev := os.Getenv("RPI_INTERFACE")
	if dev == "" {
		panic("RPI_INTERFACE not set")
	}

	p, err := pcap.NewInactiveHandle(dev)
	if err != nil {
		panic(err)
	}
	err = p.SetPromisc(true)
	if err != nil {
		panic(err)
	}
	err = p.SetSnapLen(256)
	if err != nil {
		panic(err)
	}
	//err = p.SetTimeout(d * time.Millisecond)
	//if err != nil {
	//	panic(err)
	//}

	handle, err := p.Activate()
	if err != nil {
		panic(err)
	}

	epoch := time.Now()
	t := time.NewTimer(d * time.Millisecond)
	for {
		select {
		case <-t.C:
			close(ch)
			t.Stop()
			handle.Close()
			return
		default:
			packet, info, err := handle.ReadPacketData()
			if err != nil {
				log.Println("read packet data failed with error: ", err)
				handle.Close()
			}

			// 8 byte time in milliseconds
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, uint64(info.Timestamp.Sub(epoch).Milliseconds()))
			packet = append(buf, packet...)
			ch <- packet
		}
	}
}

func getNonStop(ch chan<- []byte) {
	dev := os.Getenv("RPI_INTERFACE")
	if dev == "" {
		panic("RPI_INTERFACE not set")
	}

	handle, err := pcap.OpenLive(dev, 256, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	epoch := time.Now()
	for {
		packet, info, err := handle.ReadPacketData()
		if err != nil {
			panic(err)
		}

		//8 byte time in milliseconds
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(info.Timestamp.Sub(epoch).Milliseconds()))
		packet = append(buf, packet...)
		ch <- packet
	}
}

func getPacketData(packet []byte) error {
	var signCH0 uint32 = 0
	var signCH1 uint32 = 0
	var signCH2 uint32 = 0
	var signCH3 uint32 = 0
	var str string

	defaultBuilder.Reset()

	_ = 1 // adcNum // TODO: should change!

	_ = packet[:8] // time
	if int8(packet[24]) < 0 {
		signCH0 = 255 << 24
	}
	str = strconv.FormatInt(int64(int32(signCH0+(uint32(packet[24])<<16)+(uint32(packet[25])<<8)+(uint32(packet[26])))), 10)
	defaultBuilder.WriteString(str)
	defaultBuilder.WriteString(",")

	if int8(packet[28]) < 0 {
		signCH1 = 255 << 24
	}
	str = strconv.FormatInt(int64(int32(signCH1+(uint32(packet[28])<<16)+(uint32(packet[29])<<8)+(uint32(packet[30])))), 10)
	defaultBuilder.WriteString(str)
	defaultBuilder.WriteString(",")

	if int8(packet[32]) < 0 {
		signCH2 = 255 << 24
	}
	str = strconv.FormatInt(int64(int32(signCH2+(uint32(packet[32])<<16)+(uint32(packet[33])<<8)+(uint32(packet[34])))), 10)
	defaultBuilder.WriteString(str)
	defaultBuilder.WriteString(",")

	if int8(packet[36]) < 0 {
		signCH3 = 255 << 24
	}
	str = strconv.FormatInt(int64(int32(signCH3+(uint32(packet[36])<<16)+(uint32(packet[37])<<8)+(uint32(packet[38])))), 10)
	defaultBuilder.WriteString(str)
	defaultBuilder.WriteString("\r\n")
	return nil
}

func init() {
	rootCmd.AddCommand(adcReadCmd)
	defaultBuilder.Grow(256)
	defaultBuilder.Reset()
}
