package xmega

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/go-cmd/cmd"
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
	tx = []byte{uint8(0x01), uint8(0x0f), 0}
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
	var tx []byte
	rx := make([]byte, 3)
	connDigitRegex := regexp.MustCompile(`\d*\.\d*`)

	tx = []byte{uint8(0x01), uint8(0x00), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("failed to reset all logic analyzers: %v", err)
	}
	_ = driver.DisableChipSelect(0)

	tx = []byte{uint8(0x01), uint8(0x01), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("detecting logic 1 failed: %v", err)
	}
	_ = driver.DisableChipSelect(0)

	{
		<-cmd.NewCmd("sigrok-cli", "--scan").Start()
		time.Sleep(2 * time.Second)
		status := <-cmd.NewCmd("sigrok-cli", "--scan").Start()
		if len(status.Stdout) < 3 {
			return nil, fmt.Errorf("looks like logic 2 is not connected")
		}
		list = append(list, connDigitRegex.FindString(status.Stdout[2]))
	}

	tx = []byte{uint8(0x01), uint8(0x02), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("detecting logic 2 failed: %v", err)
	}
	_ = driver.DisableChipSelect(0)

	{
		<-cmd.NewCmd("sigrok-cli", "--scan").Start()
		time.Sleep(2 * time.Second)
		status := <-cmd.NewCmd("sigrok-cli", "--scan").Start()
		if len(status.Stdout) < 3 {
			return nil, fmt.Errorf("looks like logic 3 is not connected")
		}
		list = append(list, connDigitRegex.FindString(status.Stdout[2]))
	}

	tx = []byte{uint8(0x01), uint8(0x04), 0}
	_ = driver.EnableChipSelect(0)
	if err = conn.Tx(tx, rx); err != nil {
		return nil, fmt.Errorf("detecting logic 3 failed: %v", err)
	}
	_ = driver.DisableChipSelect(0)

	{
		<-cmd.NewCmd("sigrok-cli", "--scan").Start()
		time.Sleep(2 * time.Second)
		status := <-cmd.NewCmd("sigrok-cli", "--scan").Start()
		if len(status.Stdout) < 3 {
			return nil, fmt.Errorf("looks like logic 3 is not connected")
		}
		list = append(list, connDigitRegex.FindString(status.Stdout[2]))
	}

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
