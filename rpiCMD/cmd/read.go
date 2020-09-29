package cmd

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
)

const (
	logic1DataReadyMask uint8 = 0x80 >> iota
	logic1DataClockMask
	logic1DataOut1Mask
	logic1DataOut0Mask
	logic1SyncMask
	logic1DataOut3Mask
	logic1DataOut2Mask
	logic1DataOut5Mask
)

const second = 60

var defaultBuilder = strings.Builder{}
var inputPipe io.ReadCloser
var buffer = make([]byte, second*24_000_000, second*24_000_000)

// adcConvertCmd represents the convert command
var adcConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "read data from 'if', convert the data and write to 'of'",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			o   io.WriteCloser
			err error
		)

		input, _ := cmd.Flags().GetString("if")
		if input == "" {
			inputPipe = os.Stdin
		} else {
			inputPipe, err = os.Open(input)
			if err != nil {
				log.Fatalln(err)
			}
		}

		output, _ := cmd.Flags().GetString("of")
		if output == "" {
			o = os.Stdout
		} else {
			o, err = os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				log.Fatalln(err)
			}
		}
		w := bufio.NewWriter(o)
		convert(w)
	},
}

var adcReadCmd = &cobra.Command{
	Use:   "read",
	Short: "read data from logic analyzer and write to 'of'",
	Run: func(cmd *cobra.Command, args []string) {
		w := bufio.NewWriterSize(nil, second*1024*1024)
		output, _ := cmd.Flags().GetString("of")
		if output == "" {
			w.Reset(os.Stdout)
		} else {
			o, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				log.Fatalln(err)
			}
			w.Reset(o)
		}

		c := exec.Command(
			"sigrok-cli",
			"--driver=fx2lafw", "-O", "binary", "--time", fmt.Sprintf("%ds", second), "--config", "samplerate=24m")

		var err error
		inputPipe, err = c.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := c.Start(); err != nil {
			log.Fatalf("read command sigrok-cli start failed: %v", err)
		}
		convert(w)
		if err := c.Wait(); err != nil {
			log.Fatal(err)
		}
	},
}

func convert(w *bufio.Writer) {
	defer w.Flush()
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		for sig := range interruptChan {
			log.Println(sig)
			err := w.Flush()
			if err != nil {
				log.Println("interrupt and flush failed with error: ", err)
			}
			os.Exit(1)
		}
	}()

	data := make([]uint32, 5, 5)
	dclkChan := clockIndex()

	line := make([]byte, 20, 20)
	for i := range dclkChan {
		for c := 0; c < 4; c++ {
			for x := 0; x < 8; x++ { //skip 8 bits of header
				_ = <-dclkChan
			}

			i = <-dclkChan // bit 23 of data

			data[0], data[1], data[2], data[3], data[4] = 0, 0, 0, 0, 0
			if buffer[i]&logic1DataOut1Mask > 0 {
				data[0] = 255 << 24
			}
			if buffer[i]&logic1DataOut0Mask > 0 {
				data[1] = 255 << 24
			}
			if buffer[i]&logic1DataOut3Mask > 0 {
				data[2] = 255 << 24
			}
			if buffer[i]&logic1DataOut2Mask > 0 {
				data[3] = 255 << 24
			}
			if buffer[i]&logic1DataOut5Mask > 0 {
				data[4] = 255 << 24
			}

			for counter := 23; counter >= 0; counter-- {
				data[0] |= uint32(buffer[i]&logic1DataOut1Mask) >> 5 << counter
				data[1] |= uint32(buffer[i]&logic1DataOut0Mask) >> 4 << counter
				data[2] |= uint32(buffer[i]&logic1DataOut3Mask) >> 2 << counter
				data[3] |= uint32(buffer[i]&logic1DataOut2Mask) >> 1 << counter
				data[4] |= uint32(buffer[i]&logic1DataOut5Mask) >> 0 << counter
				if counter > 0 {
					i = <-dclkChan
				}
			}
			binary.LittleEndian.PutUint32(line[0:4], data[0])
			binary.LittleEndian.PutUint32(line[4:8], data[1])
			binary.LittleEndian.PutUint32(line[8:12], data[2])
			binary.LittleEndian.PutUint32(line[12:16], data[3])
			binary.LittleEndian.PutUint32(line[16:20], data[4])

			w.Write(line)
		}
	}
}

func clockIndex() (dclkChan <-chan int) {
	dclk := make(chan int, 1000)
	drdyChan := dataReadyIndex()
	go func(drdyChan <-chan int) {

		tempClock := [2]bool{
			false,
			false,
		}

		for i := range drdyChan {

			// without this the first for loop inside read function won't work correctly
			// maybe there is a cleaner way to fix this?
			dclk <- i + 1

			for clock := 0; clock < 32*4; {
				tempClock[0] = !tempClock[1]
				tempClock[1] = buffer[i+1]&logic1DataClockMask == 0
				if !(tempClock[0] && tempClock[1]) {
					i++
					continue
				}
				dclk <- i + 1
				clock++
			}
		}
		close(dclk)
	}(drdyChan)
	return dclk
}

func dataReadyIndex() (drdyChan <-chan int) {
	drdy := make(chan int, 1000)
	counter := 0
	go func() {
		var (
			err error
			n   = 0
			min = second * 24_000_000
		)
		tempDataReady := []bool{
			buffer[0]&logic1DataReadyMask > 0,
			buffer[1]&logic1DataReadyMask == 0,
		}

		for n < min && err == nil {
			var nn int
			nn, err = inputPipe.Read(buffer[n:])
			for i := n; i < n+nn-1; i++ {
				tempDataReady[0] = !tempDataReady[1]
				tempDataReady[1] = buffer[i+1]&logic1DataReadyMask == 0
				if !(tempDataReady[0] && tempDataReady[1]) {
					continue
				}
				counter++
				drdy <- i
			}
			n += nn
		}
		if err != nil {
			log.Fatal(err)
		}
		close(drdy)
		fmt.Printf("%d\n", counter)
	}()
	return drdy
}

func init() {
	rootCmd.AddCommand(adcConvertCmd, adcReadCmd)

	adcConvertCmd.Flags().String("if", "", "the file to read and convert")
	adcConvertCmd.Flags().String("of", "", "the file to write the result")

	adcReadCmd.Flags().String("if", "", "the file to read and convert")
	adcReadCmd.Flags().String("of", "", "the file to write the result")

	defaultBuilder.Grow(256)
	defaultBuilder.Reset()
}
