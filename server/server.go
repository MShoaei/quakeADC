package server

import (
	"fmt"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
)

type HeaderData struct {
	EnabledChannels [24]bool   `json:"EnabledChannels"`
	Gains           [24]uint32 `json:"Gains"`
	Window          int        `json:"Window"`
}

type server struct {
	api           *gin.Engine
	adc           *driver.Adc7768
	hd            HeaderData
	logics        []string
	sigrokRunning bool

	dataFS   afero.Fs
	memFS    afero.Fs
	dataFile afero.File

	Debug        bool
	GainMultiply uint32
}

func NewServer(dataFS, memFS afero.Fs, adcConnection *driver.Adc7768, debug bool) *server {
	s := &server{
		adc: adcConnection,
		hd: HeaderData{
			Gains: [24]uint32{
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1,
			},
		},
		dataFS:       dataFS,
		memFS:        memFS,
		GainMultiply: 1000,
	}
	s.api = s.NewAPI()
	return s
}

func (s *server) Run(addr ...string) error {
	return s.api.Run(addr...)
}

func (s *server) HardwareInitSeq() error {
	driver.ReadID(s.adc.Connection())
	time.Sleep(100 * time.Millisecond)

	if err := driver.TurnOnAllADC(s.adc.Connection()); err != nil {
		return fmt.Errorf("failed to reset ADCs: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := driver.ResetAllADC(s.adc.Connection()); err != nil {
		return fmt.Errorf("failed to reset ADCs: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	list, err := driver.DetectLogicConnString(s.adc.Connection())
	if err != nil {
		return fmt.Errorf("failed to detect logic analyzers conn string: %v", err)
	}
	s.logics = list
	time.Sleep(100 * time.Millisecond)

	if err := driver.EnableMCLK(s.adc.Connection()); err != nil {
		return fmt.Errorf("failed to enable MCLK: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := driver.StatusLED(s.adc.Connection(), driver.On); err != nil {
		return fmt.Errorf("failed to turn on LED: %v", err)
	}
	time.Sleep(5000 * time.Millisecond)

	driver.SendSyncSignal()
	s.adc.CilabrateChOffset(s.logics[0], s.Debug)

	return nil
}
