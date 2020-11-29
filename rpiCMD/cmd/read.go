package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
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
	var d int
	d, _ = strconv.Atoi(driverConnDigits[0])
	c1 := exec.Command(
		"sigrok-cli",
		"--driver=fx2lafw=1."+strconv.Itoa(d), "-O", "binary", "--time", strconv.Itoa(duration), "-o", tempFilePath1, "--config", "samplerate=24m")
	//d, _ = strconv.Atoi(driverConnDigits[1])
	//c2 := exec.Command(
	//	"sigrok-cli",
	//	"--driver=fx2lafw=1."+strconv.Itoa(d), "-O", "binary", "--time", strconv.Itoa(duration), "-o", tempFilePath2, "--config", "samplerate=24m")
	//d, _ = strconv.Atoi(driverConnDigits[2])
	//c3 := exec.Command(
	//	"sigrok-cli",
	//	"--driver=fx2lafw=1."+strconv.Itoa(d), "-O", "binary", "--time", strconv.Itoa(duration), "-o", tempFilePath3, "--config", "samplerate=24m")

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

	if err := json.NewEncoder(dataFile).Encode(enabledChannels); err != nil {
		// this should never happen!
		return fmt.Errorf("error while encoding enabled channels: %v", err)
	}
	//convert(file1, file2, file3, dataFile, stat1.Size())
	convert(file1, nil, nil, dataFile, stat1.Size(), enabledChannels)
	return nil
}

var (
	buffer1 = new(bytes.Buffer)
	buffer2 = new(bytes.Buffer)
	buffer3 = new(bytes.Buffer)
)

func convert(reader1 io.Reader, reader2 io.Reader, reader3 io.Reader, writer io.WriteCloser, size int64, channels [24]bool) {
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

	defer func() {
		err := recover()
		if err != nil {
			log.Printf("ignoring panic: %v", err)
		}
		writer.Close()
	}()

	if reader1 != nil {
		buffer1.Reset()
		buffer1.Grow(int(size))

	}
	if reader2 != nil {
		buffer1.Reset()
		buffer2.Grow(int(size))

	}
	if reader3 != nil {
		buffer1.Reset()
		buffer3.Grow(int(size))
	}

	data := make([]uint32, 6, 6)
	line := make([]byte, 0, 72*4)

	buffer1.ReadFrom(reader1)
	//buffer2.ReadFrom(reader2)
	//buffer3.ReadFrom(reader3)

	bytes1 := buffer1.Bytes()
	//bytes2 := buffer2.Bytes()
	//bytes3 := buffer3.Bytes()

	for i := 0; i < len(bytes1)-1; i++ {
		if bytes1[i]&logic1DataReadyMask == 128 && bytes1[i+1]&logic1DataReadyMask == 0 {
			for dataColumn := 0; dataColumn < 4; dataColumn++ {
				for j := 0; j < 8; {
					if bytes1[i]&logic1DataClockMask == 64 && bytes1[i+1]&logic1DataClockMask == 0 {
						j++
					}
					i++
				}

				data[0], data[1], data[2], data[3], data[4] = 0, 0, 0, 0, 0
				for bytes1[i]&logic1DataClockMask != 64 || bytes1[i+1]&logic1DataClockMask != 0 {
					i++
				}
				if bytes1[i+1]&logic1DataOut0Mask == 16 {
					data[0] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut1Mask == 32 {
					data[1] = 255 << 24
				}
				if bytes1[i]&logic1DataOut2Mask == 2 {
					data[2] = 255 << 24
				}
				if bytes1[i+1]&logic1DataOut3Mask == 4 {
					data[3] = 255 << 24
				}
				if bytes1[i]&logic1DataOut5Mask == 1 {
					data[5] = 255 << 24
				}
				for counter := 23; counter >= 0; counter-- {
					for bytes1[i]&logic1DataClockMask != 64 || bytes1[i+1]&logic1DataClockMask != 0 {
						i++
					}
					data[0] |= uint32(bytes1[i]&logic1DataOut0Mask) >> 4 << counter
					data[1] |= uint32(bytes1[i]&logic1DataOut1Mask) >> 5 << counter
					data[2] |= uint32(bytes1[i]&logic1DataOut2Mask) >> 1 << counter
					data[3] |= uint32(bytes1[i]&logic1DataOut3Mask) >> 2 << counter
					data[5] |= uint32(bytes1[i]&logic1DataOut5Mask) >> 0 << counter
					i++
				}

				value := make([]byte, 4, 4)
				if channels[0+dataColumn] {
					binary.LittleEndian.PutUint32(value, data[0])
					line = append(line, value...)
				}
				if channels[1*4+dataColumn] {
					binary.LittleEndian.PutUint32(value, data[1])
					line = append(line, value...)
				}
				if channels[2*4+dataColumn] {
					binary.LittleEndian.PutUint32(value, data[2])
					line = append(line, value...)
				}
				if channels[3*4+dataColumn] {
					binary.LittleEndian.PutUint32(value, data[3])
					line = append(line, value...)
				}
				if channels[5*4+dataColumn] {
					binary.LittleEndian.PutUint32(value, data[5])
					line = append(line, value...)
				}
			}
			writer.Write(line)
			line = make([]byte, 0, 72*4)
		}
	}
}

func init() {
}
