package cmd

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
)

type FlushWriter interface {
	io.Writer
	Flush() error
}

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

//var input io.ReadCloser
//var buffer []byte
var driverConnDigits []string

func execSigrokCLI(duration int) error {
	sigrokRunning = true
	defer func() { sigrokRunning = false }()
	homePath, _ := os.UserHomeDir()
	tempFilePath1 := path.Join(homePath, "quakeWorkingDir", "temp", "data1.raw")
	//tempFilePath2 := path.Join(homePath, "quakeWorkingDir", "temp", "data2.raw")
	//tempFilePath3 := path.Join(homePath, "quakeWorkingDir", "temp", "data3.raw")
	c1 := exec.Command(
		"sigrok-cli",
		"--driver=fx2lafw="+driverConnDigits[0], "-O", "binary", "--time", fmt.Sprintf("%d", duration), "-o", tempFilePath1, "--config", "samplerate=24m")
	//c2 := exec.Command(
	//	"sigrok-cli",
	//	"--driver=fx2lafw="+driverConnDigits[0], "-O", "binary", "--time", fmt.Sprintf("%d", duration), "-o", tempFilePath2, "--config", "samplerate=24m")
	//c3 := exec.Command(
	//	"sigrok-cli",
	//	"--driver=fx2lafw="+driverConnDigits[0], "-O", "binary", "--time", fmt.Sprintf("%d", duration), "-o", tempFilePath3, "--config", "samplerate=24m")

	if err := c1.Start(); err != nil {
		return fmt.Errorf("start 'c1' failed with error: %v", err)
	}
	//if err := c2.Start(); err != nil {
	//	return fmt.Errorf("start 'c2' failed with error: %v", err)
	//}
	//if err := c3.Start(); err != nil {
	//	return fmt.Errorf("start 'c3' failed with error: %v", err)
	//}

	if err := c1.Wait(); err != nil {
		return fmt.Errorf("c1 completed with error: %v", err)
	}
	//if err := c2.Wait(); err != nil {
	//	return fmt.Errorf("c2 completed with error: %v", err)
	//}
	//if err := c3.Wait(); err != nil {
	//	return fmt.Errorf("c3 completed with error: %v", err)
	//}

	stat1, _ := os.Stat(tempFilePath1)
	//stat2, _ := os.Stat(tempFilePath2)
	//stat3, _ := os.Stat(tempFilePath3)
	//if stat1.Size() != stat2.Size() || stat2.Size() != stat3.Size() {
	//	return fmt.Errorf("data size does not match")
	//}
	file1, _ := os.Open(tempFilePath1)
	//file2, _ := os.Open(tempFilePath2)
	//file3, _ := os.Open(tempFilePath3)

	//convert(file1, file2, file3, dataFile, stat1.Size())
	convert(file1, nil, nil, dataFile, stat1.Size())
	return nil
}

var (
	buffer1 = new(bytes.Buffer)
	buffer2 = new(bytes.Buffer)
	buffer3 = new(bytes.Buffer)
)

func convert(reader1 io.Reader, reader2 io.Reader, reader3 io.Reader, writer io.WriteCloser, size int64) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		for sig := range interruptChan {
			log.Println(sig)
			err := writer.Close()
			if err != nil {
				log.Println("interrupt received and close failed with error: ", err)
			}
			os.Exit(1)
		}
	}()

	if reader1 != nil {
		buffer1.Grow(int(size))

	}
	if reader2 != nil {
		buffer2.Grow(int(size))

	}
	if reader3 != nil {
		buffer3.Grow(int(size))
	}

	data := make([]uint32, 5, 5)
	line := make([]byte, 20, 20)

	buffer1.ReadFrom(reader1)
	//buffer2.ReadFrom(reader2)
	//buffer3.ReadFrom(reader3)

	bytes1 := buffer1.Bytes()
	//bytes2 := buffer2.Bytes()
	//bytes3 := buffer3.Bytes()

	for i := 0; i < len(bytes1)-1; i++ {
		if bytes1[i]&logic1DataReadyMask == 1 && bytes1[i+1]&logic1DataReadyMask == 0 {
			for j := 0; j < 8; {
				if bytes1[i]&logic1DataClockMask == 1 && bytes1[i+1]&logic1DataClockMask == 0 {
					j++
				}
				i++
			}
			data[0], data[1], data[2], data[3], data[4] = 0, 0, 0, 0, 0
			if bytes1[i]&logic1DataOut1Mask > 0 {
				data[0] = 255 << 24
			}
			if bytes1[i]&logic1DataOut0Mask > 0 {
				data[1] = 255 << 24
			}
			if bytes1[i]&logic1DataOut3Mask > 0 {
				data[2] = 255 << 24
			}
			if bytes1[i]&logic1DataOut2Mask > 0 {
				data[3] = 255 << 24
			}
			if bytes1[i]&logic1DataOut5Mask > 0 {
				data[4] = 255 << 24
			}
			for counter := 23; counter >= 0; counter-- {
				for bytes1[i]&logic1DataClockMask == 1 && bytes1[i+1]&logic1DataClockMask == 0 {
					i++
				}
				data[0] |= uint32(bytes1[i]&logic1DataOut0Mask) >> 4 << counter
				data[1] |= uint32(bytes1[i]&logic1DataOut1Mask) >> 5 << counter
				data[2] |= uint32(bytes1[i]&logic1DataOut2Mask) >> 1 << counter
				data[3] |= uint32(bytes1[i]&logic1DataOut3Mask) >> 2 << counter
				data[4] |= uint32(bytes1[i]&logic1DataOut5Mask) >> 0 << counter
			}
			binary.LittleEndian.PutUint32(line[0:4], data[0])
			binary.LittleEndian.PutUint32(line[4:8], data[1])
			binary.LittleEndian.PutUint32(line[8:12], data[2])
			binary.LittleEndian.PutUint32(line[12:16], data[3])
			binary.LittleEndian.PutUint32(line[16:20], data[4])

			writer.Write(line)
		}
	}
}

func init() {
}
