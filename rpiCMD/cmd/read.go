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
	"time"
	"unsafe"

	"github.com/google/gousb"
	"github.com/spf13/cobra"
)

const (
	cmdStartFlagsCLK48MHZ   uint8 = 1 << 6
	cmdSTartFlagsSample8Bit uint8 = 0 << 5
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
	logic1SyncMask
	logic1DataOut3Mask
	logic1DataOut2Mask
	logic1DataOut5Mask
)

//var input io.ReadCloser
//var buffer []byte
var driverConnDigits []string

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "command for operation",
}

func newMonitorLiveCommand() *cobra.Command {
	options := struct {
		sample int
	}{}
	cmd := &cobra.Command{
		Use: "monitor",
		Run: func(cmd *cobra.Command, args []string) {
			f, _ := os.Create("test.raw")
			monitorLive(f, options.sample)
		},
	}
	f := cmd.Flags()
	f.SortFlags = false
	f.IntVar(&options.sample, "sample", 0, "")
	_ = cmd.MarkFlagRequired("sample")

	return cmd
}

func execSigrokCLI(duration int) error {
	sigrokRunning = true
	defer func() { sigrokRunning = false }()
	homePath, _ := os.UserHomeDir()
	tempFilePath1 := path.Join(homePath, "quakeWorkingDir", "temp", "data1.raw")
	//tempFilePath2 := path.Join(homePath, "quakeWorkingDir", "temp", "data2.raw")
	//tempFilePath3 := path.Join(homePath, "quakeWorkingDir", "temp", "data3.raw")
	var d int
	d, _ = strconv.Atoi(driverConnDigits[1])
	c1 := exec.Command(
		"sigrok-cli",
		"--driver=fx2lafw:conn=1."+strconv.Itoa(d), "-O", "binary", "-D", "--time", strconv.Itoa(duration), "-o", tempFilePath1, "--config", "samplerate=24m")
	//d, _ = strconv.Atoi(driverConnDigits[0])
	//c2 := exec.Command(
	//	"sigrok-cli",
	//	"--driver=fx2lafw:conn=1."+strconv.Itoa(d), "-O", "binary", "-D", "--time", strconv.Itoa(duration), "-o", tempFilePath2, "--config", "samplerate=24m")
	//d, _ = strconv.Atoi(driverConnDigits[2])
	//c3 := exec.Command(
	//	"sigrok-cli",
	//	"--driver=fx2lafw:conn=1."+strconv.Itoa(d), "-O", "binary", "-D", "--time", strconv.Itoa(duration), "-o", tempFilePath3, "--config", "samplerate=24m")

	log.Println(c1.String())
	// log.Println(c2.String())
	// log.Println(c3.String())

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

	if err := json.NewEncoder(dataFile).Encode(hd); err != nil {
		// this should never happen!
		return fmt.Errorf("error while encoding enabled channels: %v", err)
	}
	//convert(file1, file2, file3, dataFile, stat1.Size())
	convert(file1, nil, nil, dataFile, int(stat1.Size()), hd.EnabledChannels)
	return nil
}

var (
	buffer1 = new(bytes.Buffer)
	buffer2 = new(bytes.Buffer)
	buffer3 = new(bytes.Buffer)
)

//TODO: unsure about type
const k float32 = 0.00000048828125 * 1e6 // (4.096/2^23)*1e6

func convert(reader1 io.Reader, reader2 io.Reader, reader3 io.Reader, writer io.WriteCloser, size int, channels [24]bool) {
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

	var line []byte

	buffer1.ReadFrom(reader1)
	//buffer2.ReadFrom(reader2)
	//buffer3.ReadFrom(reader3)

	bytes1 := buffer1.Bytes()
	//bytes2 := buffer2.Bytes()
	//bytes3 := buffer3.Bytes()

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
					data[5+offset] |= uint32(bytes1[i]&logic1DataOut5Mask) >> 0 << counter
					i++
				}
			}
			for index, value := range enChannels {
				binary.LittleEndian.PutUint32(line[index*4:], uint32(int32(float32(int32(data[(value*6)%24+(value/4)]))*k)))
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

func monitorLive(w io.WriteCloser, samples int) {
	ctx := gousb.NewContext()
	defer ctx.Close()
	vid, pid := gousb.ID(0x0925), gousb.ID(0x3881)
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return desc.Vendor == vid && desc.Product == pid
	})
	// All returned devices are now open and will need to be closed.
	for _, d := range devs {
		defer d.Close()
	}
	if err != nil {
		log.Fatalf("OpenDevices(): %v", err)
	}
	if len(devs) == 0 {
		log.Fatalf("no devices found matching VID %s and PID %s", vid, pid)
	}

	dev := devs[0]

	cfg, err := dev.Config(1)
	if err != nil {
		log.Fatalf("%s.Config(1): %v", dev, err)
	}
	defer cfg.Close()

	cmd := cmdStartAcquisition{}
	cmd.Flags = cmdStartFlagsCLK48MHZ
	cmd.Flags |= cmdSTartFlagsSample8Bit
	cmd.Flags |= 0 // not using analog channels
	cmd.SampleDelayH = (delay >> 8) & 0xff
	cmd.SampleDelayL = delay & 0xff

	const sz = int(unsafe.Sizeof(cmdStartAcquisition{}))
	var asByteSlice []byte = (*(*[sz]byte)(unsafe.Pointer(&cmd)))[:]

	num, err := dev.Control(gousb.ControlVendor|gousb.ControlOut, 0xb1, 0, 0, asByteSlice)
	if num != 3 || err != nil {
		log.Fatalln(err)
	}

	intf, err := cfg.Interface(0, 0)
	if err != nil {
		log.Fatalf("%s.Interface(0, 0): %v", cfg, err)
	}
	defer intf.Close()

	epIn, err := intf.InEndpoint(2)
	if err != nil {
		log.Fatalf("%s.InEndpoint(2): %v", intf, err)
	}
	log.Println(epIn.Desc.MaxPacketSize)

	stream, err := epIn.NewStream(512*10, 1000)
	if err != nil {
		log.Fatalf("failed to create ReadStream: %v", err)
	}

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
				// if b[i+1]&logic1DataOut0Mask == 16 {
				// 	dataBuf[0] = 255 << 24
				// }
				// if b[i+1]&logic1DataOut1Mask == 32 {
				// 	dataBuf[1] = 255 << 24
				// }
				if b[i+1]&logic1DataOut2Mask == 2 {
					dataBuf[2] = 255 << 24
				}
				// if b[i+1]&logic1DataOut3Mask == 4 {
				// 	dataBuf[3] = 255 << 24
				// }
				// if b[i+1]&logic1DataOut5Mask == 1 {
				// 	dataBuf[5] = 255 << 24
				// }
			}

			for counter := 32 - clkCounter; counter >= 0; counter-- {
				for b[i]&logic1DataClockMask != 64 || b[i+1]&logic1DataClockMask != 0 {
					i++
				}
				clkCounter++
				// dataBuf[0] |= uint32(b[i]&logic1DataOut0Mask) >> 4 << counter
				// dataBuf[1] |= uint32(b[i]&logic1DataOut1Mask) >> 5 << counter
				dataBuf[2] |= uint32(b[i]&logic1DataOut2Mask) >> 1 << counter
				// dataBuf[3] |= uint32(b[i]&logic1DataOut3Mask) >> 2 << counter
				// dataBuf[5] |= uint32(b[i]&logic1DataOut5Mask) >> 0 << counter
				i++
			}
			clkCounter = -1
			line := make([]byte, 24, 24)
			// binary.LittleEndian.PutUint32(line[0:], uint32(int32(float32(int32(dataBuf[0])))))
			// binary.LittleEndian.PutUint32(line[4:], uint32(int32(float32(int32(dataBuf[1])))))
			binary.LittleEndian.PutUint32(line[8:], uint32(int32(float32(int32(dataBuf[2])))))
			// binary.LittleEndian.PutUint32(line[12:], uint32(int32(float32(int32(dataBuf[3])))))
			// binary.LittleEndian.PutUint32(line[16:], uint32(int32(float32(int32(dataBuf[4])))))
			// binary.LittleEndian.PutUint32(line[20:], uint32(int32(float32(int32(dataBuf[5])))))
			w.Write(line)
			dataBuf[0], dataBuf[1], dataBuf[2], dataBuf[3], dataBuf[4], dataBuf[5] = 0, 0, 0, 0, 0, 0
		}
	}
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.AddCommand(newMonitorLiveCommand())
}
