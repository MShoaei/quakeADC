package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/MShoaei/quakeADC/driver/xmega"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var dataFS afero.Fs
var memFS afero.Fs

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "launch a server to execute command",
	Long:  `launch a server which listens on port 9090 and executes commands.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOARCH == "arm" {
			if err := HardwareInitSeq(); err != nil {
				log.Fatalf("hardware init failed: %v", err)
			}
			log.Println("hardware init SUCCESSFUL")
		}

		wd, _ := os.Getwd()
		if err := os.MkdirAll(path.Join(wd, "data"), os.ModeDir|0755); err != nil {
			log.Fatalf("failed to create directory: %v", err)
		}
		dataFS = afero.NewBasePathFs(afero.NewOsFs(), path.Join(wd, "data"))
		memFS = afero.NewMemMapFs()
		_, _ = memFS.Create("/data.raw")

		api := NewAPI()
		port := "9090"
		if os.Getenv("PORT") != "" {
			port = os.Getenv("PORT")
		}
		_ = api.Run(":" + port)
	},
}

func HardwareInitSeq() error {
	xmega.ReadID(adcConnection.Connection())
	time.Sleep(100 * time.Millisecond)

	if err := xmega.TurnOnAllADC(adcConnection.Connection()); err != nil {
		return fmt.Errorf("failed to reset ADCs: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := xmega.ResetAllADC(adcConnection.Connection()); err != nil {
		return fmt.Errorf("failed to reset ADCs: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	//TODO: Multi logic analyzer is not implemented. The function below SHOULD be implemented.
	list, err := xmega.DetectLogicConnString(adcConnection.Connection())
	if err != nil {
		return fmt.Errorf("failed to detect logic analyzers conn string: %v", err)
	}
	driverConnDigits = list
	time.Sleep(100 * time.Millisecond)

	if err := xmega.EnableMCLK(adcConnection.Connection()); err != nil {
		return fmt.Errorf("failed to enable MCLK: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := xmega.StatusLED(adcConnection.Connection(), xmega.On); err != nil {
		return fmt.Errorf("failed to turn on LED: %v", err)
	}
	time.Sleep(5000 * time.Millisecond)

	SendSyncSignal()
	cilabrateChOffset()

	return nil
}

func cilabrateChOffset() {
	const MSBMask uint32 = 0x00ff0000
	const MidMask uint32 = 0x0000ff00
	const LSBMask uint32 = 0x000000ff
	gainOpts := driver.ChannelGainOpts{Write: true}
	for i := 0; i < 24; i++ {
		gainOpts.Channel = uint8(i) % 8
		val := 1000 * gainMultiply
		gainOpts.Offset[0] = uint8((val & MSBMask) >> 16)
		gainOpts.Offset[1] = uint8((val & MidMask) >> 8)
		gainOpts.Offset[2] = uint8(val & LSBMask)
		adcConnection.ChannelGain(gainOpts, uint8(i/8)+1, debug)
		time.Sleep(100 * time.Millisecond)
	}

	offsetOpts := driver.ChannelOffsetOpts{Write: true}
	for i := 0; i < 24; i++ {
		offsetOpts.Channel = uint8(i) % 8
		adcConnection.ChannelOffset(offsetOpts, uint8(i/8)+1, debug)
		time.Sleep(100 * time.Millisecond)
	}

	ChOpts := driver.ChStandbyOpts{
		Write:    true,
		Channels: [8]bool{},
	}
	enabledCh := [24]bool{}
	for i := 0; i < 24; i++ {
		enabledCh[i] = true
		ChOpts.Channels[i%8] = false
		if i%8 == 7 {
			if i < 8 {
				ChOpts.Channels[0] = false
			}
			log.Println(ChOpts, uint8(i/8)+1)
			adcConnection.ChStandby(ChOpts, uint8(i/8)+1)
			ChOpts.Channels = [8]bool{}
			time.Sleep(100 * time.Millisecond)
		}
	}

	xmega.SamplingStart(adcConnection.Connection())

	homePath, _ := os.UserHomeDir()
	tempFilePath1 := path.Join(homePath, "quakeWorkingDir", "temp", "data1.raw")
	d, _ := strconv.Atoi(driverConnDigits[0])
	c1 := exec.Command(
		"sigrok-cli",
		"--driver=fx2lafw:conn=1."+strconv.Itoa(d), "-O", "binary", "-D", "--time", strconv.Itoa(1024), "-o", tempFilePath1, "--config", "samplerate=24m")
	log.Println(c1.String())
	if err := c1.Start(); err != nil {
		log.Fatalf("start 'c1' in calibrate failed with error: %v", err)
	}
	if err := c1.Wait(); err != nil {
		log.Fatalf("c1 completed with error: %v", err)
	}
	stat1, _ := os.Stat(tempFilePath1)
	file1, _ := os.Open(tempFilePath1)

	buf := bytes.NewBuffer(make([]byte, 0, 1024*24*4))
	convert(file1, nil, nil, buf, int(stat1.Size()), enabledCh)
	file1.Close()

	xmega.SamplingEnd(adcConnection.Connection())

	total := make([]int, 24)
	data := buf.Bytes()
	for i := 0; i < len(data); i += 4 * 24 {
		for j := 0; j < 24 && i+4+j*4 < len(data); j++ {
			offset := i + j*4
			total[j] += int(int32(binary.LittleEndian.Uint32([]byte{data[offset], data[offset+1], data[offset+2], data[offset+3]})))
		}
	}
	for i := 0; i < len(total); i++ {
		total[i] = total[i] / 1024
	}

	log.Println(total)

	offsetOpts = driver.ChannelOffsetOpts{Write: true}
	for i := 0; i < 24; i++ {
		offsetOpts.Channel = uint8(i) % 8
		val := uint32(int32(float32(total[i]) * 4.2))
		offsetOpts.Offset[0] = uint8((val & MSBMask) >> 16)
		offsetOpts.Offset[1] = uint8((val & MidMask) >> 8)
		offsetOpts.Offset[2] = uint8(val & LSBMask)
		log.Println(offsetOpts.Offset)
		adcConnection.ChannelOffset(offsetOpts, uint8(i/8)+1, debug)
		time.Sleep(100 * time.Millisecond)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
