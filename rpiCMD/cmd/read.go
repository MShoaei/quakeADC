package cmd

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
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

// var debug

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
	defer opt.file.Close()
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
			return nil
		}

		err := getPacketData(packet)
		if err != nil {
			log.Println("failed to read data from packet: ", err)
			return err
		}
		_, _ = defaultWriter.WriteString(defaultBuilder.String())
		if counter%opt.skip == 0 {
			opt.ch <- defaultBuilder.String()
			defaultBuilder.Reset()
			counter = 0
		}
		counter++
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

var adcConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert file created by sigrok-cli",
	// Long:  "Start Reading from RPI_INTERFACE",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		inFile, err := os.Open(input)
		if err != nil {
			log.Fatalf("could not open file: %v", err)
		}

		output, _ := cmd.Flags().GetString("output")
		outFile, err := os.Create(output)
		if err != nil {
			log.Fatalf("create file failed: %v", err)
		}
		defaultWriter.Reset(outFile)
		defer defaultWriter.Flush()
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			return testConvert(inFile, outFile)
		}

		return convert(inFile, defaultWriter)
	},
}

func testConvert(inFile io.Reader, out io.Writer) error {
	fmt.Println("testing")
	var (
		n   int64
		err error
	)

	n, err = io.CopyN(out, inFile, 100)
	for n == 100 && err == nil {
		n, err = io.CopyN(out, inFile, 100)
		out.Write([]byte{0x1, 0x2})
	}
	// ioutil.ReadAll()
	return err
}

func convert(inFile io.Reader, outFile io.StringWriter) (err error) {
	const (
		drdyMask uint8 = 1 << iota
		d0Mask
		d1Mask
		d2Mask
		d3Mask
	)
	data := make([]uint32, 4)

	buf, err := ioutil.ReadAll(inFile)
	if err != nil {
		log.Fatalf("read failed: %v", err)
	}

	dclkChan := index(buf)
	for i := range dclkChan {
		if buf[i]&drdyMask == 0 {
			continue
		}
		for buf[i]&drdyMask > 0 {
			for x := 0; x < 8; x++ { //skip 8 bits of header
				_ = <-dclkChan
			}
			i = <-dclkChan // bit 23 of data
			data[0], data[1], data[2], data[3] = 0, 0, 0, 0
			if buf[i]&d0Mask > 0 {
				data[0] = 255 << 24
			}
			if buf[i]&d1Mask > 0 {
				data[1] = 255 << 24
			}
			if buf[i]&d2Mask > 0 {
				data[2] = 255 << 24
			}
			if buf[i]&d3Mask > 0 {
				data[3] = 255 << 24
			}
			for counter := 23; counter >= 0; counter-- {
				data[0] |= uint32(buf[i]&d0Mask) >> 1 << counter
				data[1] |= uint32(buf[i]&d1Mask) >> 2 << counter
				data[2] |= uint32(buf[i]&d2Mask) >> 3 << counter
				data[3] |= uint32(buf[i]&d3Mask) >> 4 << counter
				if counter > 0 {
					i = <-dclkChan
				}
			}

			outFile.WriteString(
				strconv.FormatInt(int64(int32(data[0])), 10) + "," +
					strconv.FormatInt(int64(int32(data[1])), 10) + "," +
					strconv.FormatInt(int64(int32(data[2])), 10) + "," +
					strconv.FormatInt(int64(int32(data[3])), 10) + "\n")
		}
	}
	return nil
}

func index(buf []byte) (dclkChan <-chan int) {
	const dclkMask = 0x40
	dclk := make(chan int, 200)
	go func() {
		temp := [2]bool{
			buf[0]&dclkMask > 0,
			buf[1]&dclkMask == 0,
		}
		//_ = buf[len(buf)-1]
		for i := 0; i < len(buf)-1; i++ {
			temp[0] = !temp[1]
			temp[1] = buf[i+1]&dclkMask == 0
			if !(temp[0] && temp[1]) {
				continue
			}
			dclk <- i
		}
		close(dclk)
	}()
	return dclk
}

func init() {
	rootCmd.AddCommand(adcReadCmd)

	rootCmd.AddCommand(adcConvertCmd)
	adcConvertCmd.Flags().String("input", "", "the file to read and convert")
	_ = adcConvertCmd.MarkFlagRequired("input")
	adcConvertCmd.Flags().String("output", "", "the file to write the result")
	_ = adcConvertCmd.MarkFlagRequired("output")

	defaultBuilder.Grow(256)
	defaultBuilder.Reset()
}

