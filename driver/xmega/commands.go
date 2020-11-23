package xmega

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/spf13/afero"
	"gobot.io/x/gobot/drivers/spi"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/bcm283x"
)

func Reset() error {
	if err := bcm283x.GPIO26.SetFunc(gpio.OUT_LOW); err != nil {
		return fmt.Errorf("reset failed: %v", err)
	}
	bcm283x.GPIO26.FastOut(gpio.Low)
	bcm283x.GPIO26.FastOut(gpio.High)
	if err := bcm283x.GPIO26.SetFunc(gpio.IN); err != nil {
		return fmt.Errorf("reset failed: %v", err)
	}
	return nil
}

func Shutdown(conn spi.Connection) {
	var tx []byte
	tx = []byte{uint8(0x02), uint8(0x02), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	_ = conn.Tx(tx, rx)
	_ = driver.DisableChipSelect(0)
}

type status int

const (
	Off status = 0
	On  status = 1
)

func StatusLED(conn spi.Connection, s status) error {
	var tx []byte
	switch s {
	case On:
		tx = []byte{uint8(4), uint8(1), 0}
	case Off:
		tx = []byte{uint8(4), uint8(0), 0}
	}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	if err := conn.Tx(tx, rx); err != nil {
		return fmt.Errorf("failed to set led status: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	return nil
}

func TurnOnAllADC(conn spi.Connection) error {
	var tx []byte
	tx = []byte{uint8(0x05), uint8(0x07), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	if err := conn.Tx(tx, rx); err != nil {
		return fmt.Errorf("reset all adcs failed: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	return nil
}

func ResetAllADC(conn spi.Connection) error {
	var tx []byte
	tx = []byte{uint8(0x01), uint8(0x01), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	if err := conn.Tx(tx, rx); err != nil {
		return fmt.Errorf("reset all adcs failed: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	return nil
}

func EnableMCLK(conn spi.Connection) error {
	var tx []byte
	tx = []byte{uint8(0x02), uint8(0x01), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	if err := conn.Tx(tx, rx); err != nil {
		return fmt.Errorf("enabling MCLK failed: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	return nil
}

func DetectLogicConnString(conn spi.Connection) (list []string, err error) {
	var devices []os.FileInfo
	var tx []byte
	rx := make([]byte, 3)

	tx = []byte{uint8(0x01), uint8(0x00), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("failed to reset all logic analyzers: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	time.Sleep(1 * time.Second)

	devices, err = afero.ReadDir(afero.NewOsFs(), "/dev/bus/usb/001/")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory at /dev/bus/usb/001/: %v", err)
	}

	tx = []byte{uint8(0x01), uint8(0x02), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("failed to enable logic 1: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	time.Sleep(1 * time.Second)
	devices, err = afero.ReadDir(afero.NewOsFs(), "/dev/bus/usb/001/")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory at /dev/bus/usb/001/: %v", err)
	}
	list = append(list, devices[len(devices)-1].Name())

	tx = []byte{uint8(0x01), uint8(0x06), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("failed to enable logic 2: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	time.Sleep(1 * time.Second)
	devices, err = afero.ReadDir(afero.NewOsFs(), "/dev/bus/usb/001/")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory at /dev/bus/usb/001/: %v", err)
	}
	list = append(list, devices[len(devices)-1].Name())

	tx = []byte{uint8(0x01), uint8(0x0e), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("failed to enable logic 3: %v", err)
	}
	_ = driver.DisableChipSelect(0)
	time.Sleep(1 * time.Second)
	devices, err = afero.ReadDir(afero.NewOsFs(), "/dev/bus/usb/001/")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory at /dev/bus/usb/001/: %v", err)
	}
	list = append(list, devices[len(devices)-1].Name())

	log.Printf("%#v", list)

	return list, err
}

func ReadID(conn spi.Connection) {
	tx := []byte{uint8(0x88), uint8(0), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	conn.Tx(tx, rx)
	_ = driver.DisableChipSelect(0)
	log.Println(rx)
}

func SamplingStart(conn spi.Connection) {
	tx := []byte{uint8(0x10), uint8(0x42), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	conn.Tx(tx, rx)
	_ = driver.DisableChipSelect(0)
	log.Println(rx)
}

func SamplingEnd(conn spi.Connection) {
	tx := []byte{uint8(0x10), uint8(0x02), 0}
	rx := make([]byte, 3)

	_ = driver.EnableChipSelect(0)
	conn.Tx(tx, rx)
	_ = driver.DisableChipSelect(0)
	log.Println(rx)
}

func GetVoltage(conn spi.Connection) []int16 {
	tx := make([]byte, 3, 3)
	rx := make([]byte, 3, 3)
	res := make([]int16, 0, 12)
	for i := uint8(0); i < 4; i++ {
		if i == 1 || i == 2 {
			res = append(res, 0, 0, 0)
			continue
		}
		tx = []byte{uint8(0x0b), 0x04 | i, 0}
		_ = driver.EnableChipSelect(0)
		_ = conn.Tx(tx, rx)
		_ = driver.DisableChipSelect(0)
		time.Sleep(100 * time.Millisecond)

		for j := uint8(2); j < 8; j += 2 {
			tx = []byte{uint8(0x0c), j, 0}
			_ = driver.EnableChipSelect(0)
			_ = conn.Tx(tx, rx)
			_ = driver.DisableChipSelect(0)
			time.Sleep(100 * time.Millisecond)

			var value int16
			tx = []byte{uint8(0x8d), 0, 0}
			_ = driver.EnableChipSelect(0)
			_ = conn.Tx(tx, rx)
			_ = driver.DisableChipSelect(0)
			time.Sleep(100 * time.Millisecond)
			value = int16(rx[2]) << 8

			tx = []byte{uint8(0x8e), 0, 0}
			_ = driver.EnableChipSelect(0)
			_ = conn.Tx(tx, rx)
			_ = driver.DisableChipSelect(0)
			time.Sleep(100 * time.Millisecond)
			value |= int16(rx[2])

			res = append(res, value)
		}
	}
	return res
}

func GetCurrent(conn spi.Connection) []int16 {
	tx := make([]byte, 3, 3)
	rx := make([]byte, 3, 3)
	res := make([]int16, 0, 12)
	for i := uint8(0); i < 4; i++ {
		tx = []byte{uint8(0x0b), 0x04 | i, 0}
		_ = driver.EnableChipSelect(0)
		_ = conn.Tx(tx, rx)
		_ = driver.DisableChipSelect(0)
		time.Sleep(100 * time.Millisecond)

		for j := uint8(1); j < 7; j += 2 {
			tx = []byte{uint8(0x0c), j, 0}
			_ = driver.EnableChipSelect(0)
			_ = conn.Tx(tx, rx)
			_ = driver.DisableChipSelect(0)
			time.Sleep(100 * time.Millisecond)

			var value int16
			tx = []byte{uint8(0x8d), 0, 0}
			_ = driver.EnableChipSelect(0)
			_ = conn.Tx(tx, rx)
			_ = driver.DisableChipSelect(0)
			time.Sleep(100 * time.Millisecond)
			value = int16(rx[2]) << 8

			tx = []byte{uint8(0x8e), 0, 0}
			_ = driver.EnableChipSelect(0)
			_ = conn.Tx(tx, rx)
			_ = driver.DisableChipSelect(0)
			time.Sleep(100 * time.Millisecond)
			value |= int16(rx[2])

			res = append(res, value)
		}
	}
	return res
}
