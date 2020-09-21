package cmd

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
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

// readCmd represents the read command
var adcReadCmd = &cobra.Command{
	Use:   "readAll",
	Short: "Start Reading from data file",
	Run: func(cmd *cobra.Command, args []string) {
		input, _ := cmd.Flags().GetString("if")
		i, err := os.Open(input)
		if err != nil {
			log.Fatalln(err)
		}

		output, _ := cmd.Flags().GetString("of")
		if output == "" {
			read(i, os.Stdout)
			return
		}

		o, err := os.OpenFile(output, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		read(i, o)
	},
}

func read(inFile io.ReadCloser, outFile io.WriteCloser) {
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
	buf, _ := ioutil.ReadAll(inFile)
	dclkChan := clockIndex(buf)
	for i := range dclkChan {
		for c := 0; c < 4; c++ {
			for x := 0; x < 8; x++ { //skip 8 bits of header
				_ = <-dclkChan
			}

			i = <-dclkChan // bit 23 of data

			data[0], data[1], data[2], data[3], data[4] = 0, 0, 0, 0, 0
			if buf[i]&logic1DataOut1Mask > 0 {
				data[0] = 255 << 24
			}
			if buf[i]&logic1DataOut0Mask > 0 {
				data[1] = 255 << 24
			}
			if buf[i]&logic1DataOut3Mask > 0 {
				data[2] = 255 << 24
			}
			if buf[i]&logic1DataOut2Mask > 0 {
				data[3] = 255 << 24
			}
			if buf[i]&logic1DataOut5Mask > 0 {
				data[4] = 255 << 24
			}

			for counter := 23; counter >= 0; counter-- {
				data[0] |= uint32(buf[i]&logic1DataOut1Mask) >> 5 << counter
				data[1] |= uint32(buf[i]&logic1DataOut0Mask) >> 4 << counter
				data[2] |= uint32(buf[i]&logic1DataOut3Mask) >> 2 << counter
				data[3] |= uint32(buf[i]&logic1DataOut2Mask) >> 1 << counter
				data[4] |= uint32(buf[i]&logic1DataOut5Mask) >> 0 << counter
				if counter > 0 {
					i = <-dclkChan
				}
			}

			defaultWriter.WriteString(
				strconv.FormatInt(int64(int32(data[0])), 10) + "," +
					strconv.FormatInt(int64(int32(data[1])), 10) + "," +
					strconv.FormatInt(int64(int32(data[2])), 10) + "," +
					strconv.FormatInt(int64(int32(data[3])), 10) + "," +
					strconv.FormatInt(int64(int32(data[4])), 10) + ",")
			//outFile.Write([]byte(fmt.Sprintf("%024b\n",int64(int32(data[0])))))
		}
		defaultWriter.WriteByte('\n')
	}
	defaultWriter.Flush()
}

func clockIndex(buf []byte) (dclkChan <-chan int) {
	dclk := make(chan int, 1000)
	drdyChan := dataReadyIndex(buf)
	go func(drdyChan <-chan int) {
		tempClock := [2]bool{
			buf[0]&logic1DataClockMask > 0,
			buf[1]&logic1DataClockMask == 0,
		}

		for i := range drdyChan {

			// without this the first for loop inside read function won't work correctly
			// maybe there is a cleaner way to fix this?
			dclk <- i + 1

			for clock := 0; clock < 32*4; {
				tempClock[0] = !tempClock[1]
				tempClock[1] = buf[i+1]&logic1DataClockMask == 0
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

func dataReadyIndex(buf []byte) (drdyChan <-chan int) {
	drdy := make(chan int, 1000)
	go func() {
		tempDataReady := []bool{
			buf[0]&logic1DataReadyMask > 0,
			buf[1]&logic1DataReadyMask == 0,
		}

		tempSync := []uint8{
			buf[0]&logic1SyncMask,
			buf[1]&logic1SyncMask,
		}
		i := 0
		for ; i < len(buf)-1; i++ {
			tempSync[0] = tempSync[1]
			tempSync[1] = buf[i+1]&logic1SyncMask
			if tempSync[0] == 0 && tempSync[1] == 8 {
				break
			}
		}
		for ; i < len(buf)-1; i++ {
			tempDataReady[0] = !tempDataReady[1]
			tempDataReady[1] = buf[i+1]&logic1DataReadyMask == 0
			if !(tempDataReady[0] && tempDataReady[1]) {
				continue
			}
			drdy <- i
		}
		close(drdy)
	}()
	return drdy
}

func init() {
	rootCmd.AddCommand(adcReadCmd)

	adcReadCmd.Flags().String("if", "", "the file to read and convert")
	_ = adcReadCmd.MarkFlagRequired("input")
	adcReadCmd.Flags().String("of", "", "the file to write the result")
	//_ = adcReadCmd.MarkFlagRequired("output")

	defaultBuilder.Grow(256)
	defaultBuilder.Reset()
}
