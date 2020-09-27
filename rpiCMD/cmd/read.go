package cmd

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
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

var defaultBuilder = strings.Builder{}
var defaultWriter = bufio.NewWriterSize(nil, 52428800)

// adcConvertCmd represents the convert command
var adcConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "read data from 'if', convert the data and write to 'of'",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			i   io.ReadCloser
			o   io.WriteCloser
			err error
		)

		// c := libcmd.NewCmd("sigrok-cli", args...)

		input, _ := cmd.Flags().GetString("if")
		if input == "" {
			i = os.Stdin
		} else {
			i, err = os.Open(input)
			if err != nil {
				log.Fatalln(err)
			}
		}

		output, _ := cmd.Flags().GetString("of")
		if output == "" {
			o = os.Stdout
		} else {
			o, err = os.OpenFile(output, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				log.Fatalln(err)
			}
		}
		convert(i, o)
	},
}

var adcReadCmd = &cobra.Command{
	Use:   "read",
	Short: "read data from logic analyzer and write to 'of'",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			i   *os.File
			o   *os.File
			err error
		)

		input, _ := cmd.Flags().GetString("if")

		output, _ := cmd.Flags().GetString("of")
		if output == "" {
			o = os.Stdout
		} else {
			o, err = os.OpenFile(output, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				log.Fatalln(err)
			}
		}
		c := exec.Command("sigrok-cli", "--driver=fx2lafw", "-O", "binary", "--time", "30s", "-o", input, "--config", "samplerate=24m")
		if err := c.Start(); err != nil {
			log.Fatalf("read command sigrok-cli start failed: %v", err)
		}

		i, err = os.Open(input)
		if err != nil {
			log.Fatalf("failed to open file to read: %v", err)
		}

		convert(i, o)
		//if err := c.Wait(); err != nil {
		//	log.Fatal(err)
		//}
	},
}

func convert(inFile io.ReadCloser, outFile io.WriteCloser) {
	defer outFile.Close()
	defer inFile.Close()

	defaultWriter.Reset(outFile)
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

	data := make([]uint32, 5, 5)
	dclkChan := clockIndex(inFile)
	for b := range dclkChan {
		for c := 0; c < 4; c++ {
			for x := 0; x < 8; x++ { //skip 8 bits of header
				_ = <-dclkChan
			}

			b = <-dclkChan // bit 23 of data

			data[0], data[1], data[2], data[3], data[4] = 0, 0, 0, 0, 0
			if b&logic1DataOut1Mask > 0 {
				data[0] = 255 << 24
			}
			if b&logic1DataOut0Mask > 0 {
				data[1] = 255 << 24
			}
			if b&logic1DataOut3Mask > 0 {
				data[2] = 255 << 24
			}
			if b&logic1DataOut2Mask > 0 {
				data[3] = 255 << 24
			}
			if b&logic1DataOut5Mask > 0 {
				data[4] = 255 << 24
			}

			for counter := 23; counter >= 0; counter-- {
				data[0] |= uint32(b&logic1DataOut1Mask) >> 5 << counter
				data[1] |= uint32(b&logic1DataOut0Mask) >> 4 << counter
				data[2] |= uint32(b&logic1DataOut3Mask) >> 2 << counter
				data[3] |= uint32(b&logic1DataOut2Mask) >> 1 << counter
				data[4] |= uint32(b&logic1DataOut5Mask) >> 0 << counter
				if counter > 0 {
					b = <-dclkChan
				}
			}

			line := make([]byte, 0, 500)

			line = strconv.AppendInt(line, int64(int32(data[0])), 10)
			line = append(line, ',')
			line = strconv.AppendInt(line, int64(int32(data[1])), 10)
			line = append(line, ',')
			line = strconv.AppendInt(line, int64(int32(data[2])), 10)
			line = append(line, ',')
			line = strconv.AppendInt(line, int64(int32(data[3])), 10)
			line = append(line, ',')
			line = strconv.AppendInt(line, int64(int32(data[4])), 10)
			line = append(line, ',')

			outFile.Write(line)
		}
		outFile.Write([]byte{'\n'})
	}
	//defaultWriter.Flush()
}

func clockIndex(inFile io.Reader) (dclkChan <-chan byte) {
	dclk := make(chan byte, 1000)
	buf := make([]byte, 1, 1)
	go func() {
		tempClock := [2]bool{
			buf[0]&logic1DataClockMask > 0,
			// buf[1]&logic1DataClockMask == 0,
			buf[0]&logic1DataClockMask > 0,
		}
		tempDataReady := [2]bool{
			buf[0]&logic1DataReadyMask > 0,
			// buf[1]&logic1DataReadyMask == 0,
			buf[0]&logic1DataReadyMask > 0,
		}

		for {
			if _, err := inFile.Read(buf); err != nil {
				log.Printf("read error: %v", err)
				break
			}
			tempDataReady[0] = !tempDataReady[1]
			tempDataReady[1] = buf[0]&logic1DataReadyMask == 0
			if !(tempDataReady[0] && tempDataReady[1]) {
				continue
			}

			// without this the first for loop inside convert function won't work correctly
			// maybe there is a cleaner way to fix this?
			dclk <- buf[0]

			for clock := 0; clock < 32*4; {
				inFile.Read(buf)
				tempClock[0] = !tempClock[1]
				tempClock[1] = buf[0]&logic1DataClockMask == 0
				if !(tempClock[0] && tempClock[1]) {
					continue
				}
				dclk <- buf[0]
				clock++
			}
		}
		close(dclk)
	}()
	return dclk
}

func init() {
	rootCmd.AddCommand(adcConvertCmd, adcReadCmd)

	adcConvertCmd.Flags().String("if", "", "the file to read and convert")
	adcConvertCmd.Flags().String("of", "", "the file to write the result")

	adcReadCmd.Flags().String("if", "", "the file to read and convert")
	_ = adcReadCmd.MarkFlagRequired("if")
	adcReadCmd.Flags().String("of", "", "the file to write the result")

	defaultBuilder.Grow(256)
	defaultBuilder.Reset()
}
