package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string
var adcConnection *driver.Adc7768

var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "rpiCMD",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var (
			err             error
			speed           int64
			bus, chip, mode int
		)

		if skip, _ := cmd.Flags().GetBool("skip"); skip {
			log.Println("Skipping")
			return nil
		}

		if adcConnection != nil {
			return nil
		}
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

		if adcConnection != nil {
			return nil
		}
		adcConnection, err = driver.GetSpiConnection(bus, chip, mode, 8, speed)
		if err != nil {
			return err
		}

		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if adcConnection != nil {
		_ = adcConnection.Close()
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var f *flag.FlagSet
	f = rootCmd.PersistentFlags()
	f.Int("bus", 0, "spi bus number and is usually 0")
	f.Int("chip", 0, "spi chipSelect number")
	f.Int("mode", 0, "spi mode number [0..3]")
	f.Int64("speed", 50000, "spi connection speed in Hz")
	f.BoolVarP(&debug, "debug", "V", false, "Debug Mode. Print Sent and received values.")
	f.BoolP("skip", "S", false, "Skip initializing spi connection. ONLY FOR TEST")
	c := f.Lookup("debug")
	c.NoOptDefVal = "true"
	c = f.Lookup("chip")
	c.Hidden = true
	f.SortFlags = false

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/rpiGo/.rpiCMD.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".rpiCMD" (without extension).
		viper.AddConfigPath(path.Join(home, "rpiGo"))
		viper.SetConfigName(".rpiCMD")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
