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