package driver

import (
	"fmt"
	"log"
	"time"

	flag "github.com/spf13/pflag"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

func (adc *Adc77684) ChStandby(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		//err   error
		//tx    = make([]byte, 2)
		//rx    = make([]byte, 2)
		h, l  uint8
		c     bool
		write bool
		//flags = cmd.Flags()
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)
	write, _ = flags.GetBool("write")
	if !write {
		h = h | 0x80
	}

	//h |= driver.ChannelStandby
	h |= ChannelStandby

	c, err = flags.GetBool("ch3")
	if err != nil {
		return nil, nil, err
	}
	if c {
		l |= 0x08
	}

	c, err = flags.GetBool("ch2")
	if err != nil {
		return nil, nil, err
	}
	if c {
		l |= 0x04
	}

	c, err = flags.GetBool("ch1")
	if err != nil {
		return nil, nil, err
	}
	if c {
		l |= 0x02
	}

	c, err = flags.GetBool("ch0")
	if err != nil {
		return nil, nil, err
	}
	if c {
		l |= 0x01
	}
	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	//if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
	//	log.Println([]byte{h, l}, rx)
	//}
	return rx, tx, nil
}

func (adc *Adc77684) ChModeA(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, _ = flags.GetBool("write")
	if !write {
		h |= 0x80
	}

	h |= ChannelModeA

	ft, err := flags.GetUint8("f-type")
	if err != nil {
		return nil, nil, err
	}
	if ft < 0 || ft > 1 {
		return nil, nil, fmt.Errorf("invalid filter type. expected 0 or 1, got %d", ft)
	}

	dr, err := flags.GetUint16("dec-rate")
	if err != nil {
		return nil, nil, err
	}

	switch dr {
	case 32:
		l |= 0x0
	case 64:
		l |= 0x1
	case 128:
		l |= 0x2
	case 256:
		l |= 0x3
	case 512:
		l |= 0x4
	case 1024:
		l |= 0x5
	default:
		return nil, nil, fmt.Errorf("invalid decimation rate. got %d", dr)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) ChModeB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, _ = flags.GetBool("write")
	if !write {
		h |= 0x80
	}

	h |= ChannelModeB

	ft, err := flags.GetUint8("f-type")
	if err != nil {
		return nil, nil, err
	}
	if ft < 0 || ft > 1 {
		return nil, nil, fmt.Errorf("invalid filter type. expected 0 or 1, got %d", ft)
	}

	dr, err := flags.GetUint16("dec-rate")
	if err != nil {
		return nil, nil, err
	}

	switch dr {
	case 32:
		l |= 0x0
	case 64:
		l |= 0x1
	case 128:
		l |= 0x2
	case 256:
		l |= 0x3
	case 512:
		l |= 0x4
	case 1024:
		l |= 0x5
	default:
		return nil, nil, fmt.Errorf("invalid decimation rate. got %d", dr)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) ChModeSel(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		c     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= ChannelModeSelect

	c, err = flags.GetUint8("ch3")
	if err != nil {
		return nil, nil, err
	}
	if c == 1 {
		l |= 0x20
	}

	c, err = flags.GetUint8("ch2")
	if err != nil {
		return nil, nil, err
	}
	if c == 1 {
		l |= 0x10
	}

	c, err = flags.GetUint8("ch1")
	if err != nil {
		return nil, nil, err
	}
	if c == 1 {
		l |= 0x02
	}

	c, err = flags.GetUint8("ch0")
	if err != nil {
		return nil, nil, err
	}
	if c == 1 {
		l |= 0x01
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) PowerMode(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= PowerMode

	s, err = flags.GetUint8("sleep")
	if err != nil {
		return nil, nil, err
	}
	if s == 1 {
		l |= 0x80
	}

	s, err = flags.GetUint8("power")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 2:
		l |= 0x20
	case 3:
		l |= 0x30
	default:
		return nil, nil, fmt.Errorf("invalid value for power. got %d, expected 0, 2 or 3", s)
	}

	s, err = flags.GetUint8("lvds-clk")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("invalid value for LVDS Clock. got %d, expected 0 or 1", s)
	}

	s, err = flags.GetUint8("mclk-div")
	if err != nil {
		return nil, nil, err
	}

	switch s {
	case 0:
		l |= 0x0
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("invalid value for MCLK division. got %d, expected 0, 2 or 3", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) GeneralConf(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= GeneralConfiguration

	s, err = flags.GetUint8("retime-en")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for retime-en, got %d", s)
	}

	s, err = flags.GetUint8("vcm-pd")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for vcm-pd, got %d", s)
	}

	// reserved bit(bit 3), should be 1
	l |= 0x04

	s, err = flags.GetUint8("vcm-vsel")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("expected 0..3 for vcm-vsel, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) DataControl(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= DataControl

	s, err = flags.GetUint8("spi-sync")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x80
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for spi-sync, got %d", s)
	}

	s, err = flags.GetUint8("single-shot")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for single-shot, got %d", s)
	}

	s, err = flags.GetUint8("spi-reset")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for spi-sync, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) InterfaceConf(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= InterfaceConfiguration

	s, err = flags.GetUint8("crc-sel")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	case 2:
		l |= 0x20
	case 3:
		l |= 0x30
	default:
		return nil, nil, fmt.Errorf("expected 0..3 for crc-sel, got %d", s)
	}

	s, err = flags.GetUint8("dclk-div")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("expected 0..3 for dclk-div, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) BISTControl(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= BISTControl

	s, err = flags.GetUint8("ram-bist-start")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ram-bist-start, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) DeviceStatus(_ *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l uint8
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	h |= 0x80
	h |= DeviceStatus

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) RevisionID(_ *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l uint8
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	h |= 0x80
	h |= RevisionID

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) GPIOControl(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= GPIOControl

	s, err = flags.GetUint8("ugpio-en")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x80
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ugpio-en, got %d", s)
	}

	s, err = flags.GetUint8("gpio4")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio4, got %d", s)
	}

	s, err = flags.GetUint8("gpio3")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio3, got %d", s)
	}

	s, err = flags.GetUint8("gpio2")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x04
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio2, got %d", s)
	}

	s, err = flags.GetUint8("gpio1")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x02
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio1, got %d", s)
	}

	s, err = flags.GetUint8("gpio0")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio0, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) GPIOWriteData(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= GPIOWriteData

	s, err = flags.GetUint8("gpio4")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio4, got %d", s)
	}

	s, err = flags.GetUint8("gpio3")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio3, got %d", s)
	}

	s, err = flags.GetUint8("gpio2")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x04
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio2, got %d", s)
	}

	s, err = flags.GetUint8("gpio1")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x02
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio1, got %d", s)
	}

	s, err = flags.GetUint8("gpio0")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for gpio0, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) GPIOReadData(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l uint8
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	h |= 0x80
	h |= GPIOReadData

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) PrechargeBuffer1(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= PrechargeBuffer1

	s, err = flags.GetUint8("ch1-neg")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch1-neg, got %d", s)
	}

	s, err = flags.GetUint8("ch1-pos")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x04
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch1-pos, got %d", s)
	}

	s, err = flags.GetUint8("ch0-neg")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x02
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch0-neg, got %d", s)
	}

	s, err = flags.GetUint8("ch0-pos")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch0-pos, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) PrechargeBuffer2(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= PrechargeBuffer2

	s, err = flags.GetUint8("ch1-neg")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch1-neg, got %d", s)
	}

	s, err = flags.GetUint8("ch1-pos")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x04
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch1-pos, got %d", s)
	}

	s, err = flags.GetUint8("ch0-neg")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x02
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch0-neg, got %d", s)
	}

	s, err = flags.GetUint8("ch0-pos")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch0-pos, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) PositiveRefPrechargeBuf(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= PositiveReferencePrechargeBuffer

	s, err = flags.GetUint8("ch3")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	s, err = flags.GetUint8("ch2")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	s, err = flags.GetUint8("ch1")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	s, err = flags.GetUint8("ch0")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) NegativeRefPrechargeBuf(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= NegativeReferencePrechargeBuffer

	s, err = flags.GetUint8("ch3")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	s, err = flags.GetUint8("ch2")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	s, err = flags.GetUint8("ch1")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	s, err = flags.GetUint8("ch0")
	if err != nil {
		return nil, nil, err
	}
	switch s {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x20
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
	}

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0OffsetMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch0OffsetMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0OffsetMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch0OffsetMid

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0OffsetLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch0OffsetLSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch1OffsetMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch1OffsetMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch1OffsetMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch1OffsetMid

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch1OffsetLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch1OffsetLSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch2OffsetMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch2OffsetMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch2OffsetMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch2OffsetMid

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch2OffsetLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch2OffsetLSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch3OffsetMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch3OffsetMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch3OffsetMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch3OffsetMid

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch3OffsetLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch3OffsetLSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0GainMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch0GainMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0GainMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch0GainMid

	s, err = flags.GetUint8("Mid")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0GainLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch0GainLSB

	s, err = flags.GetUint8("LSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch1GainMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch1GainMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch1GainMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch1GainMid

	s, err = flags.GetUint8("Mid")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch1GainLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch1GainLSB

	s, err = flags.GetUint8("LSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch2GainMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch2GainMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch2GainMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch2GainMid

	s, err = flags.GetUint8("Mid")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch2GainLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch2GainLSB

	s, err = flags.GetUint8("LSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch3GainMSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch3GainMSB

	s, err = flags.GetUint8("MSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch3GainMid(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch3GainMid

	s, err = flags.GetUint8("Mid")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch3GainLSB(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	var (
		h, l  uint8
		s     uint8
		write bool
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	write, err = flags.GetBool("write")
	if err != nil {
		return nil, nil, err
	}
	if !write {
		h = h | 0x80
	}

	h |= Ch3GainLSB

	s, err = flags.GetUint8("LSB")
	if err != nil {
		return nil, nil, err
	}
	l |= s

	tx = []byte{h, l}
	err = adc.Write(tx)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc77684) Ch0SyncOffset(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) Ch1SyncOffset(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) Ch2SyncOffset(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) Ch3SyncOffset(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) DiagnosticRX(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) DiagnosticMuxControl(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) DiagnosticDelayControl(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) ChopControl(flags *flag.FlagSet) (tx []byte, rx []byte, err error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (adc *Adc77684) HardReset(_ *flag.FlagSet) (_ []byte, _ []byte, err error) {
	r := raspi.NewAdaptor()
	pin := gpio.NewDirectPinDriver(r, "22")

	_ = pin.DigitalWrite(0)
	time.Sleep(3 * time.Second)
	_ = pin.DigitalWrite(1)
	if err := pin.Halt(); err != nil {
		log.Println(err)
		return nil, nil, err
	}
	return nil, nil, nil
}
