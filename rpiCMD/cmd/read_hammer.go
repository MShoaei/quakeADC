package cmd

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/google/gousb"
)

var (
	tempSize = 468750 * 3 // 30 seconds
	tempBuf  = make([]byte, tempSize*maxPacketSize, tempSize*maxPacketSize)
)

func readWithThreshold(threshold int, duration int) []byte {
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

	var stream *gousb.ReadStream
	stream, err = epIn.NewStream(512*10, 1000)
	if err != nil {
		log.Fatalf("failed to create ReadStream: %v", err)
	}

	size := duration * 24000 / 512
	buf := make([]byte, size*maxPacketSize, size*maxPacketSize)

	i := 0
	thresholdReached := false
	for i < tempSize-1 {
		_, err := stream.Read(tempBuf[i*maxPacketSize : (i+1)*maxPacketSize])
		if err != nil {
			log.Fatal(err)
		}
		if i == 0 {
			if checkThreshold(tempBuf[i*maxPacketSize:(i+1)*maxPacketSize], threshold) {
				thresholdReached = true
				break
			}
		} else {
			if checkThreshold(tempBuf[i*maxPacketSize-1:(i+1)*maxPacketSize], threshold) {
				thresholdReached = true
				break
			}
		}
		i++
	}

	if !thresholdReached {
		// f.Write(tempBuf)
		return nil
	}

	// stream, err = epIn.NewStream(512*10, 1000)
	if err != nil {
		log.Fatalf("failed to create ReadStream: %v", err)
	}

	i = 0
	start := time.Now()
	fmt.Println("data ready:", drdyCount, start)
	for i < size {
		_, err := stream.Read(buf[i*maxPacketSize : (i+1)*maxPacketSize])
		if err != nil {
			log.Fatal(err)
		}
		// if i == 0 {
		// 	liveConvert(w, buf[i*maxPacketSize:(i+1)*maxPacketSize])
		// } else {
		// 	liveConvert(w, buf[i*maxPacketSize-1:(i+1)*maxPacketSize])
		// }
		i++
	}
	fmt.Println(time.Since(start))
	stream.Close()
	return buf
}

// var f, _ = os.Create("/home/pi/Desktop/threshold2.raw")
var monitoredChannelVal uint32
var drdyCount int = 0

var hammerClockCounter = -1

func checkThreshold(b []byte, threshold int) bool {
	defer func() {
		_ = recover()
		// err := recover()
		// if err != nil {
		// 	log.Printf("ignoring panic: %v", err)
		// }
	}()
	tempVar := 0
	for i := 0; i < len(b)-1; i++ {
		tempVar++
		if hammerClockCounter != -1 || ((b[i]&logic1DataReadyMask == 128) && (b[i+1]&logic1DataReadyMask == 0)) {
			// fmt.Println(drdyCount, hammerClockCounter, i)
			if hammerClockCounter == -1 {
				hammerClockCounter = 0
				drdyCount++
			}
			if hammerClockCounter < 9 {
				for j := hammerClockCounter; j < 8; {
					if b[i]&logic1DataClockMask == 64 && b[i+1]&logic1DataClockMask == 0 {
						j++
						hammerClockCounter++
					}
					i++
				}
				for b[i]&logic1DataClockMask != 64 || b[i+1]&logic1DataClockMask != 0 {
					i++
				}
				hammerClockCounter++
				if b[i+1]&logic1DataOut0Mask == 16 {
					monitoredChannelVal = 255 << 24
				}
			}

			for counter := 32 - hammerClockCounter; counter >= 0; counter-- {
				for b[i]&logic1DataClockMask != 64 || b[i+1]&logic1DataClockMask != 0 {
					i++
				}
				hammerClockCounter++
				monitoredChannelVal |= uint32(b[i]&logic1DataOut0Mask) >> 4 << counter
				i++
			}
			hammerClockCounter = -1
			if int(int32(monitoredChannelVal)) < threshold {
				monitoredChannelVal = 0
				continue
			}
			monitoredChannelVal = 0
			return true
		}
	}
	return false
}