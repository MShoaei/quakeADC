package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	flag "github.com/spf13/pflag"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/bcm283x"
)

type ChStandbyOpts struct {
	Write    bool    `json:"write"`
	Channels [8]bool `json:"channels"`
}

func (adc *Adc7768) ChStandby(opts ChStandbyOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8

	tx = make([]byte, 2)
	rx = make([]byte, 2)
	if !opts.Write {
		h |= 0x80
	}

	h |= ChannelStandby

	for i, standby := range opts.Channels {
		if standby {
			l |= 0x01 << i
		}
	}
	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, nil
}

type ChModeOpts struct {
	Write   bool
	FType   uint8  `json:"f-type"`
	DecRate uint16 `json:"dec-rate"`
}

func (adc *Adc7768) ChModeA(opts ChModeOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8

	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h |= 0x80
	}

	h |= ChannelModeA

	if opts.FType < 0 || opts.FType > 1 {
		return nil, nil, fmt.Errorf("invalid filter type. expected 0 or 1, got %d", opts.FType)
	}
	switch opts.FType {
	case 0:
		l |= 0x0
	case 1:
		l |= 0x8
	}

	switch opts.DecRate {
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
		return nil, nil, fmt.Errorf("invalid decimation rate. got %d", opts.DecRate)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}
	return tx, rx, err
}

func (adc *Adc7768) ChModeB(opts ChModeOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8

	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h |= 0x80
	}

	h |= ChannelModeB

	if opts.FType < 0 || opts.FType > 1 {
		return nil, nil, fmt.Errorf("invalid filter type. expected 0 or 1, got %d", opts.FType)
	}
	switch opts.FType {
	case 0:
		l |= 0x0
	case 1:
		l |= 0x8
	}

	switch opts.DecRate {
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
		return nil, nil, fmt.Errorf("invalid decimation rate. got %d", opts.DecRate)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}
	return tx, rx, err
}

type ChModeSelectOpts struct {
	Write    bool
	Channels [8]uint8
}

func (adc *Adc7768) ChModeSel(opts ChModeSelectOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8

	tx = make([]byte, 2)
	rx = make([]byte, 2)
	if !opts.Write {
		h |= 0x80
	}

	h |= ChannelModeSelect

	for i, mode := range opts.Channels {
		switch mode {
		case 0:
			l |= 0x00 << i
		case 1:
			l |= 0x01 << i
		default:
			return nil, nil, fmt.Errorf("invalid channel mode. expected 0 or 1, got %d", mode)
		}
	}
	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, nil
}

type PowerModeOpts struct {
	Write     bool  `json:"write"`
	Sleep     uint8 `json:"sleep"`
	Power     uint8 `json:"power"`
	LVDSClock uint8 `json:"lvds-clock"`
	MCLKDiv   uint8 `json:"mclk-div"`
}

func (adc *Adc7768) PowerMode(opts PowerModeOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= PowerMode

	if opts.Sleep == 1 {
		l |= 0x80
	}

	switch opts.Power {
	case 0:
		l |= 0x00
	case 2:
		l |= 0x20
	case 3:
		l |= 0x30
	default:
		return nil, nil, fmt.Errorf("invalid value for power. got %d, expected 0, 2 or 3", opts.Power)
	}

	switch opts.LVDSClock {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("invalid value for LVDS Clock. got %d, expected 0 or 1", opts.LVDSClock)
	}

	switch opts.MCLKDiv {
	case 0:
		l |= 0x00
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("invalid value for MCLK division. got %d, expected 0, 2 or 3", opts.MCLKDiv)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err

}

type GeneralConfOpts struct {
	Write        bool  `json:"write"`
	RETimeEnable uint8 `json:"retime-en"`
	VcmPd        uint8 `json:"vcm-pd"`
	VcmVSelect   uint8 `json:"vcm-vsel"`
}

func (adc *Adc7768) GeneralConf(opts GeneralConfOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= GeneralConfiguration

	switch opts.RETimeEnable {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for retime-en, got %d", opts.RETimeEnable)
	}

	switch opts.VcmPd {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for vcm-pd, got %d", opts.VcmPd)
	}

	// reserved bit(bit 3), should be 1
	l |= 0x04

	switch opts.VcmVSelect {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("expected 0..3 for vcm-vsel, got %d", opts.VcmVSelect)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err

}

type DataControlOpts struct {
	Write      bool  `json:"write"`
	SpiSync    uint8 `json:"spi-sync"`
	SingleShot uint8 `json:"single-shot"`
	SpiReset   uint8 `json:"spi-reset"`
}

func (adc *Adc7768) DataControl(opts DataControlOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if opts.Write {
		h = h | 0x80
	}

	h |= DataControl

	switch opts.SpiSync {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x80
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for spi-sync, got %d", opts.SpiSync)
	}

	switch opts.SingleShot {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for single-shot, got %d", opts.SingleShot)
	}

	switch opts.SpiReset {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for spi-sync, got %d", opts.SpiReset)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type InterfaceConfOpts struct {
	Write     bool  `json:"write"`
	CRCSelect uint8 `json:"crc-sel"`
	DclkDiv   uint8 `json:"dclk-div"`
}

func (adc *Adc7768) InterfaceConf(opts InterfaceConfOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= InterfaceConfiguration

	switch opts.CRCSelect {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x10
	case 2:
		l |= 0x20
	case 3:
		l |= 0x30
	default:
		return nil, nil, fmt.Errorf("expected 0..3 for crc-sel, got %d", opts.CRCSelect)
	}

	switch opts.DclkDiv {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	case 3:
		l |= 0x03
	default:
		return nil, nil, fmt.Errorf("expected 0..3 for dclk-div, got %d", opts.DclkDiv)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type BISTControlOpts struct {
	Write        bool  `json:"write"`
	RamBISTStart uint8 `json:"ram-bist-start"`
}

func (adc *Adc7768) BISTControl(opts BISTControlOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= BISTControl

	switch opts.RamBISTStart {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x01
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ram-bist-start, got %d", opts.RamBISTStart)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc7768) DeviceStatus(cs uint8) (tx []byte, rx []byte, err error) {
	var (
		h, l uint8
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	h |= 0x80
	h |= DeviceStatus

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc7768) RevisionID(cs uint8) (tx []byte, rx []byte, err error) {
	var (
		h, l uint8
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	h |= 0x80
	h |= RevisionID

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type GPIOControlOpts struct {
	Write       bool
	UGpioEnable uint8
	GPIO        [5]uint8
}

func (adc *Adc7768) GPIOControl(opts GPIOControlOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= GPIOControl

	switch opts.UGpioEnable {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x80
	default:
		return nil, nil, fmt.Errorf("expected 0 or 1 for ugpio-en, got %d", opts.UGpioEnable)
	}

	for i, u := range opts.GPIO {
		switch u {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01 << i
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for gpio%d, got %d", i, u)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type GPIOWriteDataOpts struct {
	Write bool
	GPIO  [5]uint8
}

func (adc *Adc7768) GPIOWriteData(opts GPIOWriteDataOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= GPIOWriteData

	for i, u := range opts.GPIO {
		switch u {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01 << i
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for gpio%d, got %d", i, u)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc7768) GPIOReadData(cs uint8) (tx []byte, rx []byte, err error) {
	var (
		h, l uint8
	)
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	h |= 0x80
	h |= GPIOReadData

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type PreChargeBufferOpts struct {
	Write            bool
	ChPositiveEnable [4]uint8
	ChNegativeEnable [4]uint8
}

func (adc *Adc7768) PrechargeBuffer1(opts PreChargeBufferOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= PrechargeBuffer1

	for i, u := range opts.ChPositiveEnable {
		switch u {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01 << (i * 2)
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d-pos, got %d", i, u)
		}
	}
	for i, u := range opts.ChNegativeEnable {
		switch u {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02 << (i * 2)
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d-neg, got %d", i, u)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc7768) PrechargeBuffer2(opts PreChargeBufferOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= PrechargeBuffer2

	for i, u := range opts.ChPositiveEnable {
		switch u {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01 << (i * 2)
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d-pos, got %d", i+4, u)
		}
	}
	for i, u := range opts.ChNegativeEnable {
		switch u {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02 << (i * 2)
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d-neg, got %d", i+4, u)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type ReferencePrechargeBufOpts struct {
	Write    bool
	Channels [8]uint8
}

func (adc *Adc7768) PositiveRefPrechargeBuf(opts ReferencePrechargeBufOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= PositiveReferencePrechargeBuffer

	for i, channel := range opts.Channels {
		switch channel {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01 << i
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d, got %d", i, channel)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc7768) NegativeRefPrechargeBuf(opts ReferencePrechargeBufOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= NegativeReferencePrechargeBuffer

	for i, channel := range opts.Channels {
		switch channel {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01 << i
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d, got %d", i, channel)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type ChannelOffsetOpts struct {
	Write   bool
	Channel uint8

	// MSB, Mid, LSB
	Offset [3]uint8
}

func (adc Adc7768) ChannelOffset(opts ChannelOffsetOpts, cs uint8, debug bool) (err error) {
	var h uint8
	tx := make([]byte, 2)
	rx := make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}
	if opts.Channel < 0 || opts.Channel > 7 {
		return fmt.Errorf("invalid channel: %s", err)
	}

	register := Ch0OffsetMSB + (opts.Channel * 3)
	for i := uint8(0); i < 3; i++ {
		tx = []byte{h | (register + i), opts.Offset[i]}

		err = adc.Write(tx, cs)
		if err != nil {
			return fmt.Errorf("write error: %s", err)
		}
		err = adc.Read(rx, cs)
		if err != nil {
			return fmt.Errorf("read error: %s", err)
		}
		if debug {
			log.Println(rx)
		}
	}
	return err
}

type ChannelGainOpts struct {
	Write   bool
	Channel uint8

	// MSB, Mid, LSB
	Offset [3]uint8
}

func (adc Adc7768) ChannelGain(opts ChannelGainOpts, cs uint8, debug bool) (rx []byte, err error) {
	var h uint8
	tx := make([]byte, 2)
	rx = make([]byte, 2*3)

	if !opts.Write {
		h = h | 0x80
	}
	if opts.Channel < 0 || opts.Channel > 7 {
		return nil, fmt.Errorf("invalid channel: %s", err)
	}

	register := Ch0GainMSB + (opts.Channel * 3)
	for i := uint8(0); i < 3; i++ {
		tx = []byte{h | (register + i), opts.Offset[i]}

		err = adc.Write(tx, cs)
		if err != nil {
			return nil, fmt.Errorf("write error: %s", err)
		}
		err = adc.Read(rx[i*2:i*2+2], cs)
		if err != nil {
			return nil, fmt.Errorf("read error: %s", err)
		}

		if debug {
			log.Println(rx)
		}
	}
	return rx, nil
}

type ChannelSyncOffsetOpts struct {
	Write   bool
	Channel uint8
	Offset  uint8
}

func (adc Adc7768) ChannelSyncOffset(opts ChannelSyncOffsetOpts, cs uint8) (err error) {
	var h uint8
	tx := make([]byte, 2)
	rx := make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}
	if opts.Channel < 0 || opts.Channel > 7 {
		return fmt.Errorf("invalid channel: %s", err)
	}

	h |= Ch0SyncOffset + opts.Channel
	tx = []byte{h, opts.Offset}

	err = adc.Write(tx, cs)
	if err != nil {
		return fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return fmt.Errorf("read error: %s", err)
	}
	return err
}

type DiagnosticRXOpts struct {
	Write    bool
	Channels [8]uint8
}

func (adc *Adc7768) DiagnosticRX(opts DiagnosticRXOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= DiagnosticRX

	for i, channel := range opts.Channels {
		switch channel {
		case 0:
			l |= 0x0
		case 1:
			l |= 0x01 << i
		default:
			return nil, nil, fmt.Errorf("expected 0 or 1 for ch%d, got %d", i, channel)
		}
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type DiagnosticMuxControlOpts struct {
	Write      bool
	GrpbSelect uint8
	GrpaSelect uint8
}

func (adc *Adc7768) DiagnosticMuxControl(opts DiagnosticMuxControlOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= DiagnosticMuxControl

	switch opts.GrpbSelect {
	case 0:
		l |= 0x00
	case 3:
		l |= 0x30
	case 4:
		l |= 0x40
	case 5:
		l |= 0x50
	default:
		return nil, nil, fmt.Errorf("expected 0, 3, 4 or 5 for GRPB-SEL, got %d", opts.GrpbSelect)
	}

	switch opts.GrpaSelect {
	case 0:
		l |= 0x00
	case 3:
		l |= 0x03
	case 4:
		l |= 0x04
	case 5:
		l |= 0x05
	default:
		return nil, nil, fmt.Errorf("expected 0, 3, 4 or 5 for GRPA-SEL, got %d", opts.GrpaSelect)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type ModulatorDelayControlOpts struct {
	Write    bool
	ModDelay uint8
}

func (adc *Adc7768) ModulatorDelayControl(opts ModulatorDelayControlOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= ModulatorDelayControl
	l = 0x2

	switch opts.ModDelay {
	case 0:
		l |= 0x00
	case 1:
		l |= 0x04
	case 2:
		l |= 0x08
	case 3:
		l |= 0x0c
	default:
		return nil, nil, fmt.Errorf("expected 0..3, for mod-delay, got %d", opts.ModDelay)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

type ChopControlOpts struct {
	Write    bool
	GrpaChop uint8
	GrpbChop uint8
}

func (adc *Adc7768) ChopControl(opts ChopControlOpts, cs uint8) (tx []byte, rx []byte, err error) {
	var h, l uint8
	tx = make([]byte, 2)
	rx = make([]byte, 2)

	if !opts.Write {
		h = h | 0x80
	}

	h |= ChopControl

	switch opts.GrpaChop {
	case 1:
		l |= 0x04
	case 2:
		l |= 0x08
	default:
		return nil, nil, fmt.Errorf("expected 1 or 2 for GRPA-CHOP, got %d", opts.GrpaChop)
	}

	switch opts.GrpbChop {
	case 1:
		l |= 0x01
	case 2:
		l |= 0x02
	default:
		return nil, nil, fmt.Errorf("expected 1 or 2 for GRPB-Chop, got %d", opts.GrpaChop)
	}

	tx = []byte{h, l}

	err = adc.Write(tx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("write error: %s", err)
	}
	err = adc.Read(rx, cs)
	if err != nil {
		return nil, nil, fmt.Errorf("read error: %s", err)
	}

	return tx, rx, err
}

func (adc *Adc7768) HardReset(_ *flag.FlagSet) (_ []byte, _ []byte, err error) {
	panic("this should be implemented")
	//r := raspi.NewAdaptor()
	//pin := gpio.NewDirectPinDriver(r, "22")
	//
	//_ = pin.DigitalWrite(0)
	//time.Sleep(3 * time.Second)
	//_ = pin.DigitalWrite(1)
	//if err := pin.Halt(); err != nil {
	//	log.Println(err)
	//	return nil, nil, err
	//}
	//return nil, nil, nil
}

func (adc *Adc7768) CilabrateChOffset(logicString string, debug bool) {
	const MSBMask uint32 = 0x00ff0000
	const MidMask uint32 = 0x0000ff00
	const LSBMask uint32 = 0x000000ff
	gainOpts := ChannelGainOpts{Write: true}
	for i := 0; i < 24; i++ {
		gainOpts.Channel = uint8(i) % 8
		val := uint32(1000 * 1000)
		gainOpts.Offset[0] = uint8((val & MSBMask) >> 16)
		gainOpts.Offset[1] = uint8((val & MidMask) >> 8)
		gainOpts.Offset[2] = uint8(val & LSBMask)
		adc.ChannelGain(gainOpts, uint8(i/8)+1, debug)
		time.Sleep(100 * time.Millisecond)
	}

	offsetOpts := ChannelOffsetOpts{Write: true}
	for i := 0; i < 24; i++ {
		offsetOpts.Channel = uint8(i) % 8
		adc.ChannelOffset(offsetOpts, uint8(i/8)+1, debug)
		time.Sleep(100 * time.Millisecond)
	}

	ChOpts := ChStandbyOpts{
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
			adc.ChStandby(ChOpts, uint8(i/8)+1)
			ChOpts.Channels = [8]bool{}
			time.Sleep(100 * time.Millisecond)
		}
	}

	SamplingStart(adc.Connection())

	homePath, _ := os.UserHomeDir()
	tempFilePath1 := path.Join(homePath, "quakeWorkingDir", "temp", "data1.raw")
	if err := execSigrokCLI(tempFilePath1, logicString, 1024); err != nil {
		log.Printf("failed to record data: %v", err)
		return
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024*24*4))
	file1, _ := os.Open(tempFilePath1)
	stat1, _ := file1.Stat()
	Convert(file1, buf, int(stat1.Size()), enabledCh)
	file1.Close()

	SamplingEnd(adc.Connection())

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

	offsetOpts = ChannelOffsetOpts{Write: true}
	for i := 0; i < 24; i++ {
		offsetOpts.Channel = uint8(i) % 8
		val := uint32(int32(float32(total[i]) * 4.2))
		offsetOpts.Offset[0] = uint8((val & MSBMask) >> 16)
		offsetOpts.Offset[1] = uint8((val & MidMask) >> 8)
		offsetOpts.Offset[2] = uint8(val & LSBMask)
		log.Println(offsetOpts.Offset)
		adc.ChannelOffset(offsetOpts, uint8(i/8)+1, debug)
		time.Sleep(100 * time.Millisecond)
	}
}

func SendSyncSignal() {
	bcm283x.GPIO7.FastOut(gpio.Low)
	bcm283x.GPIO7.FastOut(gpio.High)
}
