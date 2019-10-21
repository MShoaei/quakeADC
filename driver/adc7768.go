package driver

import (
	"gobot.io/x/gobot/drivers/spi"
)

// ADC7768-4 Register Addresses
const (
	ChannelStandby uint8 = 0x00 + iota
	ChannelModeA
	ChannelModeB
	ChannelModeSelect
	PowerMode
	GeneralConfiguration
	DataControl
	InterfaceConfiguration
	BISTControl
	DeviceStatus
	RevisionID
)

// ADC7768-4 Register Addresses
const (
	GPIOControl uint8 = 0x0E + iota
	GPIOWriteData
	GPIOReadData
	PrechargeBuffer1
	PrechargeBuffer2
	PositiveRefrencePrechargeBuffer1
	NegativeRefrencePrechargeBuffer1
)

// ADC7768-4 Register Addresses
const (
	CH0OffsetMSB uint8 = 0x1E + iota
	CH0OffsetMid
	CH0OffsetLSB
	CH1OffsetMSB
	CH1OffsetMid
	CH1OffsetLSB
)

// ADC7768-4 Register Addresses
const (
	CH2OffsetMSB uint8 = 0x2A + iota
	CH2OffsetMid
	CH2OffsetLSB
	CH3OffsetMSB
	CH3OffsetMid
	CH3OffsetLSB
)

// ADC7768-4 Register Addresses
const (
	CH0GainMSB uint8 = 0x36 + iota
	CH0GainMid
	CH0GainLSB
	CH1GainMSB
	CH1GainMid
	CH1GainLSB
)

// ADC7768-4 Register Addresses
const (
	CH2GainMSB uint8 = 0x42 + iota
	CH2GainMid
	CH2GainLSB
	CH3GainMSB
	CH3GainMid
	CH3GainLSB
)

// ADC7768-4 Register Addresses
const (
	CH0SyncOffset uint8 = 0x4E + iota
	CH1SyncOffset
	_
	_
	CH2SyncOffset
	CH3SyncOffset
	_
	_
	DiagnosticRX
	DiagnosticMuxControl
	ModulatorDelayControl
	ChopControl
)

type Adc77684 struct {
	connection *spi.SpiConnection
}

func GetSpiConnection(busNum, chipNum, mode, bits int, maxSpeed int64) (*Adc77684, error) {
	c, err := spi.GetSpiConnection(busNum, chipNum, mode, bits, maxSpeed)
	if err != nil {
		return nil, err
	}

	return &Adc77684{connection: c.(*spi.SpiConnection)}, nil
}

func (adc *Adc77684) Transmit(tx, rx []byte) error {
	return adc.connection.Tx(tx, rx)
}

func (adc *Adc77684) Write(tx []byte) error {
	return adc.connection.Tx(tx, nil)
}

func (adc *Adc77684) Read(rx []byte) error {
	return adc.connection.Tx(nil, rx)
}

func (adc *Adc77684) Close() error {
	return adc.connection.Close()
}
