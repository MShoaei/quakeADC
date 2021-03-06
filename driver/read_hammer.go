package driver

import (
	"fmt"
	"log"
	"time"

	"github.com/MShoaei/quakeADC/driver/usb"
)

var (
	tempSize = 468750 * 3 // 30 seconds
	tempBuf  = make([]byte, tempSize*maxPacketSize, tempSize*maxPacketSize)
)

func ReadWithThreshold(threshold int, duration int, channel int) []byte {
	streamConnection, err := usb.NewReadStream()
	if err != nil {
		log.Fatalf("failed to create ReadStream: %v", err)
	}
	defer streamConnection.Close()

	stream := streamConnection.Stream

	size := duration * 24000 / 512
	buf := make([]byte, size*maxPacketSize, size*maxPacketSize)
	fmt.Println("Channel: ", channel)
	clockSkip = (channel%4)*32 + 8
	mask := uint8(0)
	switch channel / 4 {
	case 0:
		mask = logic1DataOut0Mask
		shift = 4
	case 1:
		mask = logic1DataOut1Mask
		shift = 5
	case 2:
		mask = logic1DataOut2Mask
		shift = 1
	case 3:
		mask = logic1DataOut3Mask
		shift = 2
	case 4:
		mask = logic1DataOut4Mask
		shift = 3
	case 5:
		mask = logic1DataOut5Mask
		shift = 0
	}

	i := 0
	thresholdReached := false

	// threshold = int(int32(float32(threshold) / k))
	log.Println(int(int32(threshold)))
	for i < tempSize-1 {
		_, err := stream.Read(tempBuf[i*maxPacketSize : (i+1)*maxPacketSize])
		if err != nil {
			log.Fatal(err)
		}
		if i == 0 {
			if checkThreshold(tempBuf[i*maxPacketSize:(i+1)*maxPacketSize], threshold, mask) {
				thresholdReached = true
				break
			}
		} else {
			if checkThreshold(tempBuf[i*maxPacketSize-1:(i+1)*maxPacketSize], threshold, mask) {
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
	fmt.Println(start)
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

var hammerClockCounter = -1
var clockSkip int
var shift int

func checkThreshold(b []byte, threshold int, mask uint8) bool {
	defer func() {
		_ = recover()
		// err := recover()
		// if err != nil {
		// 	log.Printf("ignoring panic: %v", err)
		// }
	}()
	for i := 0; i < len(b)-1; i++ {
		if hammerClockCounter != -1 || ((b[i]&logic1DataReadyMask == 128) && (b[i+1]&logic1DataReadyMask == 0)) {
			if hammerClockCounter == -1 {
				hammerClockCounter = 0
			}
			if hammerClockCounter < clockSkip+1 {
				for j := hammerClockCounter; j < clockSkip; {
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
				if b[i+1]&mask == mask {
					monitoredChannelVal = 255 << 24
				}
			}

			for counter := clockSkip + 24 - hammerClockCounter; counter >= 0; counter-- {
				for b[i]&logic1DataClockMask != 64 || b[i+1]&logic1DataClockMask != 0 {
					i++
				}
				hammerClockCounter++
				monitoredChannelVal |= uint32(b[i]&mask) >> shift << counter
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
