package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/MShoaei/rpiADC/driver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

// adcCmd represents the adc command
var adcCmd = &cobra.Command{
	Use:   "adc",
	Short: "command to control the ADC over SPI",
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

		if adcConnection == nil {
			adcConnection, err = driver.GetSpiConnection(bus, chip, mode, 8, speed)
		}

		if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
			log.Println(speed)
		}
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
			rx    = make([]byte, 2)
			h, l  uint8
			c     bool
			write bool
			flags = cmd.Flags()
		)

		write, _ = flags.GetBool("write")
		if !write {
			h = h | 0x80
		}

		h |= driver.ChannelStandby

		c, err = flags.GetBool("ch3")
		if err != nil {
			return err
		}
		if c {
			l |= 0x08
		}

		c, err = flags.GetBool("ch2")
		if err != nil {
			return err
		}
		if c {
			l |= 0x04
		}

		c, err = flags.GetBool("ch1")
		if err != nil {
			return err
		}
		if c {
			l |= 0x02
		}

		c, err = flags.GetBool("ch0")
		if err != nil {
			return err
		}
		if c {
			l |= 0x01
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			c     = make([]uint, 4)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

// TODO: Reset needs 2 successive commands which should be implemented.
var adcDataControl = &cobra.Command{
	Use:   "DataControl",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    = make([]byte, 2)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.BISTControl

		s, err = flags.GetUint8("ram-bist-start")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for ram-bist-start, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			err  error
			rx   = make([]byte, 2)
			h, l uint8
		)

		h |= 0x80
		h |= driver.DeviceStatus

		err = adcConnection.Transmit([]byte{h, l}, rx)

		log.Println([]byte{h, l}, rx)

		return err

	},
}

var adcRevisionID = &cobra.Command{
	Use:   "RevisionID",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err  error
			rx   = make([]byte, 2)
			h, l uint8
		)

		h |= 0x80
		h |= driver.RevisionID

		err = adcConnection.Transmit([]byte{h, l}, rx)

		log.Println([]byte{h, l}, rx)

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
			rx    = make([]byte, 2)
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

		h |= driver.GPIOControl

		s, err = flags.GetUint8("ugpio-en")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x80
		default:
			return fmt.Errorf("expected 0 or 1 for ugpio-en, got %d", s)
		}

		s, err = flags.GetUint8("gpio4")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x10
		default:
			return fmt.Errorf("expected 0 or 1 for gpio4, got %d", s)
		}

		s, err = flags.GetUint8("gpio3")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x08
		default:
			return fmt.Errorf("expected 0 or 1 for gpio3, got %d", s)
		}

		s, err = flags.GetUint8("gpio2")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x04
		default:
			return fmt.Errorf("expected 0 or 1 for gpio2, got %d", s)
		}

		s, err = flags.GetUint8("gpio1")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02
		default:
			return fmt.Errorf("expected 0 or 1 for gpio1, got %d", s)
		}

		s, err = flags.GetUint8("gpio0")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for gpio0, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

// TODO: Need better explanation for flags
var adcGPIOWriteData = &cobra.Command{
	Use:   "GPIOWriteData",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    = make([]byte, 2)
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

		h |= driver.GPIOWriteData

		s, err = flags.GetUint8("gpio4")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x10
		default:
			return fmt.Errorf("expected 0 or 1 for gpio4, got %d", s)
		}

		s, err = flags.GetUint8("gpio3")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x08
		default:
			return fmt.Errorf("expected 0 or 1 for gpio3, got %d", s)
		}

		s, err = flags.GetUint8("gpio2")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x04
		default:
			return fmt.Errorf("expected 0 or 1 for gpio2, got %d", s)
		}

		s, err = flags.GetUint8("gpio1")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02
		default:
			return fmt.Errorf("expected 0 or 1 for gpio1, got %d", s)
		}

		s, err = flags.GetUint8("gpio0")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for gpio0, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			err  error
			rx   = make([]byte, 2)
			h, l uint8
		)

		h |= 0x80
		h |= driver.GPIOReadData

		err = adcConnection.Transmit([]byte{h, l}, rx)

		log.Println([]byte{h, l}, rx)

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
			rx    = make([]byte, 2)
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

		h |= driver.PrechargeBuffer1

		s, err = flags.GetUint8("ch1-neg")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x08
		default:
			return fmt.Errorf("expected 0 or 1 for ch1-neg, got %d", s)
		}

		s, err = flags.GetUint8("ch1-pos")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x04
		default:
			return fmt.Errorf("expected 0 or 1 for ch1-pos, got %d", s)
		}

		s, err = flags.GetUint8("ch0-neg")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02
		default:
			return fmt.Errorf("expected 0 or 1 for ch0-neg, got %d", s)
		}

		s, err = flags.GetUint8("ch0-pos")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for ch0-pos, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.PrechargeBuffer2

		s, err = flags.GetUint8("ch3-neg")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x08
		default:
			return fmt.Errorf("expected 0 or 1 for ch3-neg, got %d", s)
		}

		s, err = flags.GetUint8("ch3-pos")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x04
		default:
			return fmt.Errorf("expected 0 or 1 for ch3-pos, got %d", s)
		}

		s, err = flags.GetUint8("ch2-neg")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02
		default:
			return fmt.Errorf("expected 0 or 1 for ch2-neg, got %d", s)
		}

		s, err = flags.GetUint8("ch2-pos")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for ch2-pos, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.PositiveReferencePrechargeBuffer

		s, err = flags.GetUint8("ch3")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x20
		default:
			return fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
		}

		s, err = flags.GetUint8("ch2")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x10
		default:
			return fmt.Errorf("expected 0 or 1 for ch2, got %d", s)
		}

		s, err = flags.GetUint8("ch1")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02
		default:
			return fmt.Errorf("expected 0 or 1 for ch1, got %d", s)
		}

		s, err = flags.GetUint8("ch0")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for ch0, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.NegativeReferencePrechargeBuffer

		s, err = flags.GetUint8("ch3")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x20
		default:
			return fmt.Errorf("expected 0 or 1 for ch3, got %d", s)
		}

		s, err = flags.GetUint8("ch2")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x10
		default:
			return fmt.Errorf("expected 0 or 1 for ch2, got %d", s)
		}

		s, err = flags.GetUint8("ch1")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x02
		default:
			return fmt.Errorf("expected 0 or 1 for ch1, got %d", s)
		}

		s, err = flags.GetUint8("ch0")
		if err != nil {
			return err
		}
		switch s {
		case 0:
			l |= 0x00
		case 1:
			l |= 0x01
		default:
			return fmt.Errorf("expected 0 or 1 for ch0, got %d", s)
		}

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH0OffsetMSB

		s, err = flags.GetUint8("MSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH0OffsetMid = &cobra.Command{
	Use:   "CH0OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    = make([]byte, 2)
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

		h |= driver.CH0OffsetMid

		s, err = flags.GetUint8("Mid")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH0OffsetLSB

		s, err = flags.GetUint8("LSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH1OffsetMSB

		s, err = flags.GetUint8("MSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH1OffsetMid = &cobra.Command{
	Use:   "CH1OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    = make([]byte, 2)
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

		h |= driver.CH1OffsetMid

		s, err = flags.GetUint8("Mid")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH1OffsetLSB

		s, err = flags.GetUint8("LSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH2OffsetMSB

		s, err = flags.GetUint8("MSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH2OffsetMid = &cobra.Command{
	Use:   "CH2OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    = make([]byte, 2)
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

		h |= driver.CH2OffsetMid

		s, err = flags.GetUint8("Mid")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH2OffsetLSB

		s, err = flags.GetUint8("LSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH3OffsetMSB

		s, err = flags.GetUint8("MSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcCH3OffsetMid = &cobra.Command{
	Use:   "CH3OffsetMid",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			err   error
			rx    = make([]byte, 2)
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

		h |= driver.CH3OffsetMid

		s, err = flags.GetUint8("Mid")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
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

		h |= driver.CH3OffsetLSB

		s, err = flags.GetUint8("LSB")
		if err != nil {
			return err
		}
		l |= s

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
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
			rx    = make([]byte, 2)
			h, l  uint8
			s     uint8
			write bool
			flags = cmd.Flags()
		)
		log.Println(s, write, flags)

		err = adcConnection.Transmit([]byte{h, l}, rx)

		if debug, _ := adcCmd.PersistentFlags().GetBool("debug"); debug {
			log.Println([]byte{h, l}, rx)
		}
		return err
	},
}

var adcResetSequence = &cobra.Command{
	Use:   "reset",
	Short: "Perform hard reset",
	Long:  `Hard reset is recommended before using the adc`,
	Run: func(cmd *cobra.Command, args []string) {
		r := raspi.NewAdaptor()
		pin := gpio.NewDirectPinDriver(r, "22")

		_ = pin.DigitalWrite(0)
		time.Sleep(5 * time.Second)
		_ = pin.DigitalWrite(1)
		if err := pin.Halt(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	var f *pflag.FlagSet
	rootCmd.AddCommand(adcCmd)
	adcCmd.AddCommand(adcChStandby, adcChModeA, adcChModeB, adcChModeSelect,
		adcPowerMode, adcGeneralConfiguration, adcDataControl, adcInterfaceConfiguration, adcBISTControl, adcDeviceStatus,
		adcRevisionID, adcGPIOControl, adcGPIOWriteData, adcGPIOReadData, adcPrechargeBuffer1, adcPrechargeBuffer2,
		adcPositiveRefPrechargeBuf, adcNegativeRefPrechargeBuf,
		adcCH0OffsetMSB, adcCH0OffsetMid, adcCH0OffsetLSB, adcCH1OffsetMSB, adcCH1OffsetMid, adcCH1OffsetLSB,
		adcCH2OffsetMSB, adcCH2OffsetMid, adcCH2OffsetLSB, adcCH3OffsetMSB, adcCH3OffsetMid, adcCH3OffsetLSB,
		adcCH0GainMSB, adcCH0GainMID, adcCH0GainLSB, adcCH1GainMSB, adcCH1GainMID, adcCH1GainLSB,
		adcCH2GainMSB, adcCH2GainMID, adcCH2GainLSB, adcCH3GainMSB, adcCH3GainMID, adcCH3GainLSB,
		adcCH0SyncOffset, adcCH1SyncOffset, adcCH2SyncOffset, adcCH3SyncOffset,
		adcDiagnosticRX, adcDiagnosticMuxControl, adcDiagnosticDelayControl, adcChopControl,
		adcResetSequence)

	f = adcCmd.PersistentFlags()
	f.Int("bus", 0, "spi bus number and is usually 0")
	f.Int("chip", 0, "spi chipSelect number")
	f.Int("mode", 0, "spi mode number [0..3]")
	f.Int64("speed", 50000, "spi connection speed in Hz")
	f.BoolP("debug", "V", false, "Debug Mode. Print Sent and received values.")
	c := f.Lookup("debug")
	c.NoOptDefVal = "true"
	f.SortFlags = false

	// ------------------------

	f = adcChStandby.Flags()
	f.Bool("write", false, "set the write bit")
	// true==standby
	// false==enabled
	f.Bool("ch3", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.Bool("ch2", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.Bool("ch1", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.Bool("ch0", true, "channels 0 standby mode. true/t: Standby, false/f: Enabled")
	f.SortFlags = false

	// ------------------------

	f = adcChModeA.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	f.Uint16("dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")
	f.SortFlags = false

	f = adcChModeB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("f-type", 1, "Filter Type Selection 0: Wideband, 1: Sinc5")
	f.Uint16("dec-rate", 1024, "Decimation Rate Selection accepted values: 32, 64, 128, 256, 512, 1024")
	f.SortFlags = false

	// ------------------------

	f = adcChModeSelect.Flags()
	f.Bool("write", false, "set the write bit")
	f.UintSlice("ch", []uint{0, 0, 0, 0}, "set channel mode for channels 0..3 0:Mode A, 1:Mode B")
	f.SortFlags = false

	// ------------------------

	f = adcPowerMode.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("sleep", 0, "0: Normal operation, 1: Sleep mode")
	f.Uint8("power", 0, "0: Low power, 2: Median, 3: Fast")
	f.Uint8("lvds-clk", 0, "0: disable LVDS clock, 1: enable LVDS clock")
	f.Uint8("mclk-div", 0, "0: set to MCLK/32 for low power mode, 2: set to MCLK/8 for median mode, 3: set to MCLK/4 for fast mode")
	f.SortFlags = false

	// ------------------------

	f = adcGeneralConfiguration.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("retime-en", 0, "SYNC_OUT signal retime enable bit. 0: disabled, 1: enabled")
	f.Uint8("vcm-pd", 0, "VCM buffer power-down. 0: enabled, 1: disabled")
	f.Uint8("vcm-vsel", 0, "VCM voltage select bits. 0: (AVDD1 − AVSS)/2 V, 1: 1.65 V, 2: 2.5 V, 3: 2.14 V")
	f.SortFlags = false

	// ------------------------

	f = adcDataControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("spi-sync", 1, "Software synchronization of the AD7768-4. This command has the same effect as sending a signal pulse to the START pin. 0: Change to SPI_SYNC low, 1: Change to SPI_SYNC high")
	f.Uint8("single-shot", 0, "One-shot mode. Enables one-shot mode. In one-shot mode, the AD7768-4 output a conversion result in response to a SYNC_IN rising edge. 0: Disabled, 1: Enabled")
	f.Uint8("spi-reset", 0, "Soft reset. Two successive commands must be received in the correct order to generate a reset. 0: No effect, 1: No effect, 2: Second reset command, 3: First reset command")
	f.SortFlags = false

	// ------------------------

	f = adcInterfaceConfiguration.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("crc-sel", 0, "CRC select. These bits allow the user to implement a CRC on the data interface. 0: No CRC. Status bits with every conversion, 1: Replace the header with CRC message every 4 samples, 2: Replace the header with CRC message every 16 samples, 3: Replace the header with CRC message every 16 samples")
	f.Uint8("dclk-div", 0, "DCLK divider. These bits control division of the DCLK clock used to clock out conversion data on the DOUTx pins. 0: Divide by 8, 1: Divide by 4, 2: Divide by 2, 3: No division")
	f.SortFlags = false

	// ------------------------

	f = adcBISTControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ram-bist-start", 0, "RAM BIST. 0: Off, 1: Begin RAM BIST")
	f.SortFlags = false

	// ------------------------

	// adcDeviceStatus

	// ------------------------

	// adcRevisionID

	// ------------------------

	f = adcGPIOControl.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ugpio-en", 0, "User GPIO enable. 0: GPIO Disabled, 1: GPIO Enabled")
	f.Uint8("gpio4", 0, "GPIO4 Direction. 0: Input, 1: Output")
	f.Uint8("gpio3", 0, "GPIO3 Direction. 0: Input, 1: Output")
	f.Uint8("gpio2", 0, "GPIO2 Direction. 0: Input, 1: Output")
	f.Uint8("gpio1", 0, "GPIO1 Direction. 0: Input, 1: Output")
	f.Uint8("gpio0", 0, "GPIO0 Direction. 0: Input, 1: Output")
	f.SortFlags = false

	// ------------------------

	f = adcGPIOWriteData.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("gpio4", 0, "GPIO4/FILTER")
	f.Uint8("gpio3", 0, "GPIO3/MODE3")
	f.Uint8("gpio2", 0, "GPIO2/MODE2")
	f.Uint8("gpio1", 0, "GPIO1/MODE1")
	f.Uint8("gpio0", 0, "GPIO0/MODE0")
	f.SortFlags = false

	// ------------------------

	// adcGPIOReadData

	// ------------------------

	f = adcPrechargeBuffer1.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch1-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch1-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch0-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch0-pos", 1, "0: Off, 1: On (default: 1)")
	f.SortFlags = false

	// ------------------------

	f = adcPrechargeBuffer2.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch3-pos", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch2-neg", 1, "0: Off, 1: On (default: 1)")
	f.Uint8("ch2-pos", 1, "0: Off, 1: On (default: 1)")
	f.SortFlags = false

	// ------------------------

	f = adcPositiveRefPrechargeBuf.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch2", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch1", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch0", 1, "0: Off, 1: On (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcNegativeRefPrechargeBuf.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("ch3", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch2", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch1", 1, "0: Off, 1: On (default: 0)")
	f.Uint8("ch0", 1, "0: Off, 1: On (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH0OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH0OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH0OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH1OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH1OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH1OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	// ------------------------

	f = adcCH2OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH2OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH2OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false
	// ------------------------

	f = adcCH3OffsetMSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("MSB", 0, "Channel 0 gain MSB signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH3OffsetMid.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("Mid", 0, "Channel 0 gain Mid signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

	f = adcCH3OffsetLSB.Flags()
	f.Bool("write", false, "set the write bit")
	f.Uint8("LSB", 0, "Channel 0 gain LSB signed 8 bit integer (default: 0)")
	f.SortFlags = false

	// ------------------------

}
