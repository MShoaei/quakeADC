package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"time"

	"github.com/MShoaei/quakeADC/driver/usb"
)

const (
	cmdStartFlagsCLK48MHZ   uint8 = 1 << 6
	cmdStartFlagsSample8Bit uint8 = 0 << 5
	delay                   uint8 = 1
)

type cmdStartAcquisition struct {
	Flags        uint8
	SampleDelayH uint8
	SampleDelayL uint8
}

const (
	logic1DataReadyMask uint8 = 0x80 >> iota
	logic1DataClockMask
	logic1DataOut1Mask
	logic1DataOut0Mask
	logic1DataOut4Mask
	logic1DataOut3Mask
	logic1DataOut2Mask
	logic1DataOut5Mask
)

func execSigrokCLI(dstPath string, logicConnDigit string, duration int) error {
	var d int
	d, _ = strconv.Atoi(logicConnDigit)
	c1 := exec.Command(
		"sigrok-cli",
		"--driver=fx2lafw:conn=1."+strconv.Itoa(d), "-O", "binary", "-D", "--time", strconv.Itoa(duration), "-o", dstPath, "--config", "samplerate=24m")

	log.Println(c1.String())

	if err := c1.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}
	return nil
}

func ExecSigrokCLI(logicConnDigit string, duration int) (io.ReadCloser, int64, error) {
	wd, _ := os.Getwd()
	tempFilePath1 := path.Join(wd, "temp", "data1.raw")

	err := execSigrokCLI(tempFilePath1, logicConnDigit, duration)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(tempFilePath1)
	if err != nil {
		return nil, 0, err
	}

	stat, _ := f.Stat()

	return f, stat.Size(), err
}

var (
	buffer1 = new(bytes.Buffer)
)

const k float32 = 0.00000048828125 * 1e6 // (4.096/2^23)*1e6

func Convert(reader1 io.Reader, writer io.Writer, size int, channels [24]bool) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		for sig := range interruptChan {
			log.Println(sig)
			os.Exit(1)
		}
	}()

	defer func() {
		err := recover()
		if err != nil {
			log.Printf("ignoring panic: %v", err)
		}
	}()

	if reader1 != nil {
		buffer1.Reset()
		buffer1.Grow(int(size))
	}

	var line []byte

	buffer1.ReadFrom(reader1)

	bytes1 := buffer1.Bytes()

	enChannels := onlyEnabledChannels(channels)
	line = make([]byte, len(enChannels)*4, len(enChannels)*4)

	for i := 0; i < len(bytes1)-1; i++ {
		if bytes1[i]&logic1DataReadyMask == 128 && bytes1[i+1]&logic1DataReadyMask == 0 {

			data := make([]uint32, 24, 24)

			for dataColumn := 0; dataColumn < 4; dataColumn++ {
				for j := 0; j < 8; {
					if bytes1[i]&logic1DataClockMask == 64 && bytes1[i+1]&logic1DataClockMask == 0 {
						j++
					}
					i++
				}

				for bytes1[i]&logic1DataClockMask != 64 || bytes1[i+1]&logic1DataClockMask != 0 {
					i++
				}
				offset := dataColumn * 6
				if bytes1[i+1]&logic1DataOut0Mask == 16 {
					data[0+offset] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut1Mask == 32 {
					data[1+offset] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut2Mask == 2 {
					data[2+offset] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut3Mask == 4 {
					data[3+offset] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut4Mask == 8 {
					data[4+offset] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut5Mask == 1 {
					data[5+offset] = 255 << 24
				}
				for counter := 23; counter >= 0; counter-- {
					for bytes1[i]&logic1DataClockMask != 64 || bytes1[i+1]&logic1DataClockMask != 0 {
						i++
					}
					data[0+offset] |= uint32(bytes1[i]&logic1DataOut0Mask) >> 4 << counter
					data[1+offset] |= uint32(bytes1[i]&logic1DataOut1Mask) >> 5 << counter
					data[2+offset] |= uint32(bytes1[i]&logic1DataOut2Mask) >> 1 << counter
					data[3+offset] |= uint32(bytes1[i]&logic1DataOut3Mask) >> 2 << counter
					data[4+offset] |= uint32(bytes1[i]&logic1DataOut4Mask) >> 3 << counter
					data[5+offset] |= uint32(bytes1[i]&logic1DataOut5Mask) >> 0 << counter
					i++
				}
			}
			for index, value := range enChannels {
				// binary.LittleEndian.PutUint32(line[index*4:], uint32(int32(float32(int32(data[(value*6)%24+(value/4)]))*k)))
				binary.LittleEndian.PutUint32(line[index*4:], uint32(int32(int32(data[(value*6)%24+(value/4)]))))
			}
			writer.Write(line)
		}
	}
}

func onlyEnabledChannels(channels [24]bool) []int {
	res := make([]int, 0, 24)
	for i, enabled := range channels {
		if enabled {
			res = append(res, i)
		}
	}
	return res
}

const maxPacketSize int = 512

func MonitorLive(w io.WriteCloser, samples int) {
	streamConnection, err := usb.NewReadStream()
	if err != nil {
		log.Fatalf("failed to create ReadStream: %v", err)
	}
	defer streamConnection.Close()

	stream := streamConnection.Stream

	size := samples / 512
	buf := make([]byte, size*maxPacketSize, size*maxPacketSize)

	i := 0
	start := time.Now()
	for i < size {
		_, err := stream.Read(buf[i*maxPacketSize : (i+1)*maxPacketSize])
		if err != nil {
			log.Fatal(err)
		}
		if i == 0 {
			liveConvert(w, buf[i*maxPacketSize:(i+1)*maxPacketSize])
		} else {
			liveConvert(w, buf[i*maxPacketSize-1:(i+1)*maxPacketSize])
		}
		i++
	}
	f, _ := os.Create("../testStream.raw")
	f.Write(buf)
	log.Println(time.Since(start))
}

var clkCounter = -1
var dataBuf = make([]uint32, 6, 6)

func liveConvert(w io.WriteCloser, b []byte) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("ignoring panic: %v", err)
		}
	}()
	for i := 0; i < len(b)-1; i++ {
		if clkCounter != -1 || (b[i]&logic1DataReadyMask == 128 && b[i+1]&logic1DataReadyMask == 0) {
			if clkCounter == -1 {
				clkCounter = 0
			}
			if clkCounter < 9 {
				for j := clkCounter; j < 8; {
					if b[i]&logic1DataClockMask == 64 && b[i+1]&logic1DataClockMask == 0 {
						j++
						clkCounter++
					}
					i++
				}
				for b[i]&logic1DataClockMask != 64 || b[i+1]&logic1DataClockMask != 0 {
					i++
				}
				clkCounter++
				if b[i+1]&logic1DataOut2Mask == 2 {
					dataBuf[2] = 255 << 24
				}
			}

			for counter := 32 - clkCounter; counter >= 0; counter-- {
				for b[i]&logic1DataClockMask != 64 || b[i+1]&logic1DataClockMask != 0 {
					i++
				}
				clkCounter++
				dataBuf[2] |= uint32(b[i]&logic1DataOut2Mask) >> 1 << counter
				i++
			}
			clkCounter = -1
			line := make([]byte, 24, 24)
			binary.LittleEndian.PutUint32(line[8:], uint32(int32(float32(int32(dataBuf[2])))))
			w.Write(line)
			dataBuf[0], dataBuf[1], dataBuf[2], dataBuf[3], dataBuf[4], dataBuf[5] = 0, 0, 0, 0, 0, 0
		}
	}
}
