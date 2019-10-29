package cmd

import (
	"fmt"
	"log"

	"github.com/MShoaei/rpiGo/driver"

	"github.com/spf13/cobra"
)

// adcCmd represents the adc command
var adcCmd = &cobra.Command{
	Use:   "adc",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var (
			err             error
			speed           int64
			bus, chip, mode int
		)
		bus, err = cmd.Flags().GetInt("bus")
		if err != nil {
			return err
		}
		chip, err = cmd.Flags().GetInt("chip")
		if err != nil {
			return err
		}
		mode, err = cmd.Flags().GetInt("mode")
		if err != nil {
			return err
		}
		speed, err = cmd.Flags().GetInt64("speed")
		if err != nil {
			return err
		}
		if mode < 0 || mode > 3 {
			return fmt.Errorf("invalid mode! expected value [0..3], got %d", mode)
		}

		adcConnection, err = driver.GetSpiConnection(bus, chip, mode, 8, speed)
		return err
	},
}

var adcChStandby = &cobra.Command{
	Use:   "chStandby",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			c     []bool
			write bool
			flags = cmd.Flags()
		)

		write, _ = flags.GetBool("write")
		if !write {
			h = h | 0x80
		}

		h |= driver.ChannelStandby
		c, err = flags.GetBoolSlice("ch")
		if err != nil {
			return err
		}

		if c[3] {
			l |= 0x08
		}
		if c[2] {
			l |= 0x04
		}
		if c[1] {
			l |= 0x02
		}
		if c[0] {
			l |= 0x01
		}

		//TODO: rx == tx after 2 consecetive transmits!!! Why?
		// fmt.Println(rx)
		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err

	},
}

var adcChModeA = &cobra.Command{
	Use:   "chModeA",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			write bool
			flags = cmd.Flags()
		)

		write, _ = flags.GetBool("write")
		if !write {
			h |= 0x80
		}

		ft, err := flags.GetUint8("f-type")
		if err != nil {
			return err
		}
		if ft < 0 || ft > 1 {
			return fmt.Errorf("invalid filter type. expected 0 or 1, got %d", ft)
		}

		h |= driver.ChannelModeA

		dr, err := flags.GetUint16("dec-rate")
		if err != nil {
			return err
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
			return fmt.Errorf("invalid decimation rate. got %d", dr)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcChModeB = &cobra.Command{
	Use:   "chModeB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			write bool
			flags = cmd.Flags()
		)

		write, _ = flags.GetBool("write")
		if !write {
			h |= 0x80
		}

		ft, err := flags.GetUint8("f-type")
		if err != nil {
			return err
		}
		if ft < 0 || ft > 1 {
			return fmt.Errorf("invalid filter type. expected 0 or 1, got %d", ft)
		}

		h |= driver.ChannelModeB

		dr, err := flags.GetUint16("dec-rate")
		if err != nil {
			return err
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
			return fmt.Errorf("invalid decimation rate. got %d", dr)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcChModeSelect = &cobra.Command{
	Use:   "chModeSel",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			c     []uint = make([]uint, 4)
			write bool
			flags = cmd.Flags()
		)

		write, err = flags.GetBool("write")
		if err != nil {
			return err
		}
		if !write {
			h = h | 0x80
		}

		h |= driver.ChannelModeSelect

		c, err = flags.GetUintSlice("ch")
		if c[3] == 1 {
			l |= 0x20
		}
		if c[2] == 1 {
			l |= 0x10
		}
		if c[1] == 01 {
			l |= 0x02
		}
		if c[0] == 1 {
			l |= 0x01
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcPowerMode = &cobra.Command{
	Use:   "PowerMode",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		write, err = flags.GetBool("write")
		if err != nil {
			return err
		}
		if !write {
			h = h | 0x80
		}

		h |= driver.PowerMode

		s, err = flags.GetUint8("sleep")
		if err != nil {
			return err
		}
		if s == 1 {
			l |= 0x80
		}

		s, err = flags.GetUint8("power")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 2:
			l |= 0x20
		case 3:
			l |= 0x30
		default:
			return fmt.Errorf("invalid value for power. got %d, expected 0, 2 or 3", s)
		}

		s, err = flags.GetUint8("lvds-clk")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x08
		default:
			return fmt.Errorf("invalid value for LVDS Clock. got %d, expected 0 or 1", s)
		}

		s, err = flags.GetUint8("mclk-div")
		if err != nil {
			return err
		}

		switch s {
		case 0:
			l |= 0x0
		case 2:
			l |= 0x02
		case 3:
			l |= 0x03
		default:
			return fmt.Errorf("invalid value for MCLK division. got %d, expected 0, 2 or 3", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcGeneralConfiguration = &cobra.Command{
	Use:   "GeneralConf",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		write, err = flags.GetBool("write")
		if err != nil {
			return err
		}
		if !write {
			h = h | 0x80
		}

		h |= driver.GeneralConfiguration

		s, err = flags.GetUint8("retime-en")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x10
		default:
			return fmt.Errorf("expected 0 or 1 for retime-en, got %d", s)
		}

		s, err = flags.GetUint8("vcm-pd")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x08
		default:
			return fmt.Errorf("expected 0 or 1 for vcm-pd, got %d", s)
		}

		// reserved bit(bit 3), should be 1
		l |= 0x04

		s, err = flags.GetUint8("vcm-vsel")
		if err != nil {
			return err
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
			return fmt.Errorf("expected 0..3 for vcm-vsel, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

// TODO: Reset needs 2 succesive commands which should be implemented.
var adcDataControl = &cobra.Command{
	Use:   "DataControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		write, err = flags.GetBool("write")
		if err != nil {
			return err
		}
		if !write {
			h = h | 0x80
		}

		h |= driver.DataControl

		s, err = flags.GetUint8("spi-sync")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x80
		default:
			return fmt.Errorf("expected 0 or 1 for spi-sync, got %d", s)
		}

		s, err = flags.GetUint8("single-shot")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x10
		default:
			return fmt.Errorf("expected 0 or 1 for single-shot, got %d", s)
		}

		s, err = flags.GetUint8("spi-reset")
		if err != nil {
			return err
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
			return fmt.Errorf("expected 0 or 1 for spi-sync, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcInterfaceConfiguration = &cobra.Command{
	Use:   "InterfaceConf",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		write, err = flags.GetBool("write")
		if err != nil {
			return err
		}
		if !write {
			h = h | 0x80
		}

		h |= driver.InterfaceConfiguration

		s, err = flags.GetUint8("crc-sel")
		if err != nil {
			return err
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
			return fmt.Errorf("expected 0..3 for crc-sel, got %d", s)
		}

		s, err = flags.GetUint8("dclk-div")
		if err != nil {
			return err
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
			return fmt.Errorf("expected 0..3 for dclk-div, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcBISTControl = &cobra.Command{
	Use:   "BISTControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcDeviceStatus = &cobra.Command{
	Use:   "DeviceStatus",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcRevisionID = &cobra.Command{
	Use:   "RevisionID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcGPIOControl = &cobra.Command{
	Use:   "GPIOControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcGPIOWriteData = &cobra.Command{
	Use:   "GPIOWriteData",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcGPIOReadData = &cobra.Command{
	Use:   "GPIOReadData",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcPrechargeBuffer1 = &cobra.Command{
	Use:   "PrechargeBuffer1",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcPrechargeBuffer2 = &cobra.Command{
	Use:   "PrechargeBuffer2",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcPositiveRefPrechargeBuf = &cobra.Command{
	Use:     "PositiveRefPrechargeBuf",
	Aliases: []string{"prpb"},
	Short:   "",
	Long:    "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcNegativeRefPrechargeBuf = &cobra.Command{
	Use:     "NegativeRefPrechargeBuf",
	Aliases: []string{"nrpb"},
	Short:   "",
	Long:    "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0OffsetMSB = &cobra.Command{
	Use:   "CH0OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0OffsetMID = &cobra.Command{
	Use:   "CH0OffsetMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0OffsetLSB = &cobra.Command{
	Use:   "CH0OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1OffsetMSB = &cobra.Command{
	Use:   "CH1OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1OffsetMID = &cobra.Command{
	Use:   "CH1OffsetMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1OffsetLSB = &cobra.Command{
	Use:   "CH1OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2OffsetMSB = &cobra.Command{
	Use:   "CH2OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2OffsetMID = &cobra.Command{
	Use:   "CH2OffsetMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2OffsetLSB = &cobra.Command{
	Use:   "CH2OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3OffsetMSB = &cobra.Command{
	Use:   "CH3OffsetMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3OffsetMID = &cobra.Command{
	Use:   "CH3OffsetMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3OffsetLSB = &cobra.Command{
	Use:   "CH3OffsetLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0GainMSB = &cobra.Command{
	Use:   "CH0GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0GainMID = &cobra.Command{
	Use:   "CH0GainMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0GainLSB = &cobra.Command{
	Use:   "CH0GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1GainMSB = &cobra.Command{
	Use:   "CH1GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1GainMID = &cobra.Command{
	Use:   "CH1GainMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1GainLSB = &cobra.Command{
	Use:   "CH1GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2GainMSB = &cobra.Command{
	Use:   "CH2GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2GainMID = &cobra.Command{
	Use:   "CH2GainMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2GainLSB = &cobra.Command{
	Use:   "CH2GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3GainMSB = &cobra.Command{
	Use:   "CH3GainMSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3GainMID = &cobra.Command{
	Use:   "CH3GainMID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3GainLSB = &cobra.Command{
	Use:   "CH3GainLSB",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0SyncOffset = &cobra.Command{
	Use:   "CH0SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1SyncOffset = &cobra.Command{
	Use:   "CH1SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2SyncOffset = &cobra.Command{
	Use:   "CH2SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3SyncOffset = &cobra.Command{
	Use:   "CH3SyncOffset",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcDiagnosticRX = &cobra.Command{
	Use:   "DiagnosticRX",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcDiagnosticMuxControl = &cobra.Command{
	Use:   "DiagnosticMuxControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcDiagnosticDelayControl = &cobra.Command{
	Use:   "DiagnosticDelayControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcChopControl = &cobra.Command{
	Use:   "ChopControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    []byte = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(adcCmd)
	adcCmd.AddCommand(adcChStandby, adcChModeA, adcChModeB, adcChModeSelect, adcPowerMode)

	adcCmd.PersistentFlags().Int("bus", 0, "spi bus number and is usually 0")
	adcCmd.PersistentFlags().Int("chip", 0, "spi chipSelect number")
	adcCmd.PersistentFlags().Int("mode", 0, "spi mode number [0..3]")
	adcCmd.PersistentFlags().Int64("speed", 50000, "spi connection speed in Hz")
	adcCmd.PersistentFlags().BoolP("debug", "V", false, "Debug Mode. Print Sent and recived values.")
	c := adcCmd.PersistentFlags().Lookup("debug")
	c.NoOptDefVal = "true"
	c.Hidden = true

	// ------------------------

	adcChStandby.Flags().Bool("write", false, "set the write bit")
	// true==standby
	// false==enabled
	adcChStandby.Flags().BoolSlice("ch", []bool{true, true, true, true}, "channels 0..3 standy mode. true/t: Standby, false/f: Enabled")

	// -------------------------

	adcChModeA.Flags().Bool("write", false, "set the write bit")
	adcChModeA.Flags().Uint8("f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	adcChModeA.Flags().Uint16("dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")

	adcChModeB.Flags().Bool("write", false, "set the write bit")
	adcChModeB.Flags().Uint8("f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	adcChModeB.Flags().Uint16("dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")

	//------------------------

	adcChModeSelect.Flags().Bool("write", false, "set the write bit")
	adcChModeSelect.Flags().UintSlice("ch", []uint{0, 0, 0, 0}, "set channel mode for channels 0..3 0:Mode A, 1:Mode B")

	//------------------------

	adcPowerMode.Flags().Bool("write", false, "set the write bit")
	adcPowerMode.Flags().Uint8("sleep", 0, "0: Normal operation, 1: Sleep mode")
	adcPowerMode.Flags().Uint8("power", 0, "0: Low power, 2: Median, 3: Fast")
	adcPowerMode.Flags().Uint8("lvds-clk", 0, "0: disable LVDS clock, 1: enable LVDS clock")
	adcPowerMode.Flags().Uint8("mclk-div", 0, "0: set to MCLK/32 for low power mode, 2: set to MCLK/8 for median mode, 3: set to MCLK/4 for fast mode")

	//------------------------

	adcGeneralConfiguration.Flags().Bool("write", false, "set the write bit")
	adcGeneralConfiguration.Flags().Uint8("retime-en", 0, "SYNC_OUT signal retime enable bit. 0: disabled, 1: enabled")
	adcGeneralConfiguration.Flags().Uint8("vcm-pd", 0, "VCM buffer power-down. 0: enabled, 1: disabled")
	adcGeneralConfiguration.Flags().Uint8("vcm-vsel", 0, "VCM voltage select bits. 0: (AVDD1 − AVSS)/2 V, 1: 1.65 V, 2: 2.5 V, 3: 2.14 V")

	//------------------------

	adcDataControl.Flags().Bool("write", false, "set the write bit")
	adcDataControl.Flags().Uint8("spi-sync", 1, "Software synchronization of the AD7768-4. This command has the same effect as sending a signal pulse to the START pin. 0: Change to SPI_SYNC low, 1: Change to SPI_SYNC high")
	adcDataControl.Flags().Uint8("single-shot", 0, "One-shot mode. Enables one-shot mode. In one-shot mode, the AD7768-4 output a conversion result in response to a SYNC_IN rising edge. 0: Disabled, 1: Enabled")
	adcDataControl.Flags().Uint8("spi-reset", 0, "Soft reset. Two successive commands must be received in the correct order to generate a reset. 0: No effect, 1: No effect, 2: Second reset command, 3: First reset command")

	//------------------------

	adcInterfaceConfiguration.Flags().Bool("write", false, "set the write bit")
	adcInterfaceConfiguration.Flags().Uint8("crc-sel", 0, "CRC select. These bits allow the user to implement a CRC on the data interface. 0: No CRC. Status bits with every conversion, 1: Replace the header with CRC message every 4 samples, 2: Replace the header with CRC message every 16 samples, 3: Replace the header with CRC message every 16 samples")
	adcInterfaceConfiguration.Flags().Uint8("dclk-div", 0, "DCLK divider. These bits control division of the DCLK clock used to clock out conversion data on the DOUTx pins. 0: Divide by 8, 1: Divide by 4, 2: Divide by 2, 3: No division")

}
