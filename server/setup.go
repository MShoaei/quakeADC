package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
)

func (s *Server) SetGainsHandler(c *gin.Context) {
	const MSBMask uint32 = 0x00ff0000
	const MidMask uint32 = 0x0000ff00
	const LSBMask uint32 = 0x000000ff
	gains := [24]uint32{}
	if err := c.BindJSON(&gains); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// for i := 0; i < len(gains); i++ {
	// 	gains[i] *= gainMultiply
	// }
	opts := driver.ChannelGainOpts{Write: true}
	for i := 0; i < len(gains); i++ {
		opts.Channel = uint8(i) % 8
		val := gains[i] * s.GainMultiply
		opts.Offset[0] = uint8((val & MSBMask) >> 16)
		opts.Offset[1] = uint8((val & MidMask) >> 8)
		opts.Offset[2] = uint8(val & LSBMask)
		log.Println(opts.Offset)
		if _, err := s.adc.ChannelGain(opts, uint8(i/8)+1, s.Debug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	s.hd.Gains = gains
	c.JSON(http.StatusOK, nil)
}

func (s *Server) GetGainsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"gains": s.hd.Gains,
	})
}

func (s *Server) SetChannelsHandler(c *gin.Context) {
	ch := [24]bool{}
	if err := c.BindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	force := false
	if !(ch[0] || ch[1] || ch[2] || ch[3]) {
		ch[0] = true
		force = true
	}
	s.hd.EnabledChannels = [24]bool{}
	opts := driver.ChStandbyOpts{
		Write:    true,
		Channels: [8]bool{},
	}
	for i, enable := range ch {
		s.hd.EnabledChannels[i] = enable
		opts.Channels[i%8] = !enable
		if i%8 == 7 {
			if i < 8 {
				opts.Channels[0] = false
			}
			log.Println(opts, uint8(i/8)+1)
			s.adc.ChStandby(opts, uint8(i/8)+1)
			opts.Channels = [8]bool{}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if force {
		s.hd.EnabledChannels[0] = false
	}
	c.Status(http.StatusOK)
}

func (s *Server) GetChannelsHandler(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"channels": s.hd.EnabledChannels,
	})
}

func (s *Server) SetupHandler(c *gin.Context) {
	if s.sigrokRunning {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "sampling is already running",
		})
		return
	}
	setupData := struct {
		StartMode        string  `json:"startMode"`
		TriggerThreshold int     `json:"threshold"`
		TriggerChannel   int     `json:"triggerChannel"`
		RecordTime       int     `json:"recordTime"`
		SamplingTime     float32 `json:"samplingTime"`
		Window           int     `json:"window"`
		FileName         string  `json:"fileName"`
	}{}
	if err := c.BindJSON(&setupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Println(setupData)
	if setupData.Window < 1 || setupData.Window > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid window size. Should be between 1 and 100",
		})
		return
	}
	s.hd.Window = setupData.Window

	if setupData.FileName == "" || s.activePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file name or project name",
		})
		return
	}
	if err := s.dataFS.MkdirAll(s.activePath, os.ModeDir|0755); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file project path",
		})
		return
	}

	if exists, _ := afero.Exists(s.dataFS, filepath.Join(s.activePath, setupData.FileName)); exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file already exists",
		})
		return
	}

	s.dataFile, _ = s.dataFS.Create(filepath.Join(s.activePath, setupData.FileName))
	defer s.dataFile.Close()

	switch strings.ToLower(setupData.StartMode) {
	case "asap":
		configureSamplingTime(s.adc, setupData.SamplingTime)
		driver.SendSyncSignal()
		driver.SamplingStart(s.adc.Connection())
		defer driver.SamplingEnd(s.adc.Connection())
		f, size, err := driver.ExecSigrokCLI(s.logics[0], setupData.RecordTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer f.Close()

		if err := json.NewEncoder(s.dataFile).Encode(s.hd); err != nil {
			// this should never happen!
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Errorf("error while encoding enabled channels: %v", err).Error(),
			})
			return
		}
		driver.Convert(f, s.dataFile, int(size), s.hd.EnabledChannels)
		c.JSON(http.StatusOK, nil)
		return
	case "hammer":
		configureSamplingTime(s.adc, setupData.SamplingTime)
		driver.SendSyncSignal()
		driver.SamplingStart(s.adc.Connection())
		defer driver.SamplingEnd(s.adc.Connection())

		rawData := driver.ReadWithThreshold(setupData.TriggerThreshold, setupData.RecordTime, setupData.TriggerChannel)
		if rawData == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "did not reach threshold",
			})
			return
		}
		if err := json.NewEncoder(s.dataFile).Encode(s.hd); err != nil {
			// this should never happen!
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Errorf("error while encoding enabled channels: %v", err).Error(),
			})
			return
		}
		driver.Convert(bytes.NewReader(rawData), s.dataFile, len(rawData), s.hd.EnabledChannels)
		c.JSON(http.StatusOK, nil)
		return
	case "trigger":
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "not implemented",
		})
		return
	}
}

func (s *Server) ReadDataHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket creation error: ", err)
		return
	}
	var file string
	if file = c.Query("file"); file == "" {
		_ = conn.WriteJSON(gin.H{
			"error": "invalid query parameter value",
		})
		_ = conn.Close()
		return
	}

	f, err := s.dataFS.Open(file)
	if err != nil {
		_ = conn.WriteJSON(gin.H{
			"error": fmt.Errorf("failed to open file: %v", err),
		})
		_ = conn.Close()
		return
	}
	defer f.Close()

	var header HeaderData
	b, _ := ioutil.ReadAll(f)
	infoBytes, _ := bufio.NewReader(bytes.NewReader(b)).ReadBytes('\n')
	if err := json.Unmarshal(infoBytes, &header); err != nil {
		_ = conn.WriteJSON(gin.H{
			"error": err.Error(),
		})
		_ = conn.Close()
		return
	}

	count := 0
	for _, enabled := range header.EnabledChannels {
		if enabled {
			count++
		}
	}

	b = b[len(infoBytes):]

	dataLength := count * 4
	for i := 0; i < len(b); i += dataLength {
		_ = conn.WriteMessage(websocket.BinaryMessage, b[i:i+dataLength])
	}
	_ = conn.Close()
}

func (s *Server) ReadDataPostHandler(c *gin.Context) {
	form := struct {
		File string `json:"file"`
	}{}

	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed with error",
		})
		return
	}

	info, err := s.dataFS.Stat(form.File)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("failed to open file: %v", err),
		})
		return
	}

	f, _ := s.dataFS.Open(form.File)
	b, _ := bufio.NewReader(f).ReadBytes('\n')
	var header HeaderData
	if err := json.Unmarshal(b, &header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer f.Close()

	count := int64(0)
	for _, enabled := range header.EnabledChannels {
		if enabled {
			count++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": header.EnabledChannels,
		"window":   header.Window,
		"size":     (info.Size() - int64(len(b))) / (count * 4),
	})
}

func configureSamplingTime(adc *driver.Adc7768, st float32) {
	chOpt := driver.ChModeOpts{Write: true, FType: 1}
	powerOpt := driver.PowerModeOpts{Write: true}
	interfaceOpt := driver.InterfaceConfOpts{Write: true, CRCSelect: 0}
	switch st {
	case 16:
		return
	case 31.25:
		chOpt.DecRate = 128
		powerOpt.Power = 2
		powerOpt.MCLKDiv = 2
		interfaceOpt.DclkDiv = 1
	case 62.5:
		chOpt.DecRate = 256
		powerOpt.Power = 2
		powerOpt.MCLKDiv = 2
		interfaceOpt.DclkDiv = 1
	case 125:
		chOpt.DecRate = 512
		powerOpt.Power = 2
		powerOpt.MCLKDiv = 2
		interfaceOpt.DclkDiv = 1
	case 250:
		chOpt.DecRate = 256
		powerOpt.Power = 0
		powerOpt.MCLKDiv = 0
		interfaceOpt.DclkDiv = 0
	case 500:
		chOpt.DecRate = 512
		powerOpt.Power = 0
		powerOpt.MCLKDiv = 0
		interfaceOpt.DclkDiv = 0
	case 1000:
		chOpt.DecRate = 1024
		powerOpt.Power = 0
		powerOpt.MCLKDiv = 0
		interfaceOpt.DclkDiv = 0
	case 2000:
		return
	}

	for i := uint8(1); i < 10; i++ {
		adc.ChModeA(chOpt, i)
		adc.ChModeB(chOpt, i)
		adc.PowerMode(powerOpt, i)
		adc.InterfaceConf(interfaceOpt, i)
	}
}
