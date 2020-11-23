package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/MShoaei/quakeADC/driver/xmega"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-cmd/cmd"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Println(r.Header["Origin"][0])
		return true
	},
}

var sigrokRunning = false
var dataFile afero.File

type usbDevice struct {
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	MountPoint string      `json:"mountpoint"`
	Size       string      `json:"size"`
	Children   []usbDevice `json:"children"`
}

var connectedUSB = map[int]usbDevice{}

// NewAPI creates a new API which receives commands and executes them
// the server should be started with:
// $ rpiCMD server
// the server will listen on port "9090" and accepts POST request to "/command".
// POST request should be a JSON object. e.g.:
//{
//	"Command": "chStandby",
//	"Flags": "--write --ch3 t --ch2 t --ch1 t --ch0 f"
//}
func NewAPI() *gin.Engine {
	api := gin.Default()
	if debug {
		api.Any("/api/:path", func(c *gin.Context) {
			r := c.Request
			r.URL.Path = strings.Replace(r.URL.Path, "/api", "", 1)
			//c.Application().ServeHTTPC(c)
		})
	}
	api.OPTIONS("/login", loginOptionsHandler)

	api.Use(cors.Default())

	api.GET("/status", samplingStatusHandler)

	api.GET("/tree/*dir", treeHandler)

	api.GET("/plot", readDataHandler)
	api.POST("/plot", readDataPostHandler)

	api.GET("/plot/:file", plotHandler)

	api.POST("/setup", setupHandler)
	api.POST("/command/:cmd/:adc", commandHandler)
	api.GET("/getfile", getFileHandler)

	api.PATCH("/update", updateStack)

	api.GET("/usb/all", getAllUSB)
	api.POST("/rpi/shutdown", shutdownSequenceHandler)
	api.POST("/rpi/restart", restartSequenceHandler)
	api.GET("/channels", getChannelsHandler)
	api.POST("/channels", setChannelsHandler)
	api.GET("/gains", getGainsHandler)
	api.POST("/gains", setGainsHandler)
	api.GET("/info", boardInfoHandler)

	return api
}

var wsOpen bool

func boardInfoHandler(c *gin.Context) {
	if wsOpen {
		c.JSON(http.StatusConflict, gin.H{
			"err": "another connection exists",
		})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket creation error: ", err)
		return
	}
	type message struct {
		Voltage []int16 `json:"voltage"`
		Current []int16 `json:"current"`
	}
	wsOpen = true
	conn.SetCloseHandler(func(code int, text string) error {
		log.Println(code, text)
		wsOpen = false
		return nil
	})
	for wsOpen {
		voltage := xmega.GetVoltage(adcConnection.Connection())
		current := xmega.GetCurrent(adcConnection.Connection())
		m := message{
			Voltage: voltage,
			Current: current,
		}
		conn.WriteJSON(&m)
		time.Sleep(2 * time.Second)
	}
}

func setGainsHandler(c *gin.Context) {
	const MSBMask uint32 = 0x00ff0000
	const MidMask uint32 = 0x0000ff00
	const LSBMask uint32 = 0x000000ff
	gains := [24]uint32{}
	if err := c.BindJSON(&gains); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}
	opts := driver.ChannelGainOpts{Write: true}
	for i := 0; i < len(gains); i++ {
		opts.Channel = uint8(i) - (uint8(i/8) * 8)
		opts.Offset[0] = uint8(gains[i] & MSBMask)
		opts.Offset[1] = uint8(gains[i] & MidMask)
		opts.Offset[2] = uint8(gains[i] & LSBMask)
		if err := adcConnection.ChannelGain(opts, uint8(i/8)+1); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, nil)
}

func getGainsHandler(c *gin.Context) {

}

func restartSequenceHandler(c *gin.Context) {
	xmega.Reset()
	cmd.NewCmd("/usr/bin/sudo", "/sbin/shutdown", "-r", "now").Start()
}

func setChannelsHandler(c *gin.Context) {
	ch := [24]bool{}
	if err := c.BindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}

func getChannelsHandler(c *gin.Context) {
	ch := [24]bool{}
	ch[0] = true
	opt := driver.ChStandbyOpts{
		Write:    false,
		Channels: [8]bool{},
	}
	_, rx, err := adcConnection.ChStandby(opt, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"rx": rx,
	})
}

func shutdownSequenceHandler(c *gin.Context) {
	xmega.Shutdown(adcConnection.Connection())
}

func plotHandler(c *gin.Context) {
	dir, err := afero.IsDir(dataFS, "/"+c.Param("file"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	if dir {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid path. path is a directory",
		})
		return
	}
	f, _ := dataFS.Open("/" + c.Param("file"))
	c.File(f.Name() + ".bin")
}

func treeHandler(c *gin.Context) {
	list, err := afero.ReadDir(dataFS, "/"+c.Param("dir"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid path parameter in url",
		})
		return
	}
	type item struct {
		Name string `json:"name"`
		Dir  bool   `json:"dir"`
	}
	fd := make([]item, 0)
	for _, info := range list {
		fd = append(fd, item{Name: info.Name(), Dir: info.IsDir()})
	}
	c.JSON(http.StatusOK, gin.H{
		"directory": path.Clean("/" + c.Param("dir")),
		"items":     fd,
	})
}

func configSamplingTime(st float32) {
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
		chOpt.DecRate = 1256
		powerOpt.Power = 0
		powerOpt.MCLKDiv = 0
		interfaceOpt.DclkDiv = 0
	case 500:
		chOpt.DecRate = 128
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
		adcConnection.ChModeA(chOpt, i)
		adcConnection.ChModeB(chOpt, i)
		adcConnection.PowerMode(powerOpt, i)
		adcConnection.InterfaceConf(interfaceOpt, i)
	}
}

func setupHandler(c *gin.Context) {
	if sigrokRunning {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "sampling is already running",
		})
		return
	}
	setupData := struct {
		StartMode    string  `json:"startMode"`
		RecordTime   int     `json:"recordTime"`
		SamplingTime float32 `json:"samplingTime"`
		FileName     string  `json:"fileName"`
		ProjectName  string  `json:"projectName"`
	}{}
	if err := c.BindJSON(&setupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if setupData.FileName == "" || setupData.ProjectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file name or project name",
		})
		return
	}
	if err := dataFS.MkdirAll(setupData.ProjectName, os.ModeDir|0755); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file project path",
		})
		return
	}

	if exists, _ := afero.Exists(dataFS, filepath.Join(setupData.ProjectName, setupData.FileName)); exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file already exists",
		})
		return
	}

	dataFile, _ = dataFS.Create(filepath.Join(setupData.ProjectName, setupData.FileName))

	configSamplingTime(setupData.SamplingTime)
	SendSyncSignal()
	xmega.SamplingStart(adcConnection.Connection())
	defer xmega.SamplingEnd(adcConnection.Connection())
	if err := execSigrokCLI(setupData.RecordTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func updateStack(_ *gin.Context) {
	//TODO: How to self update?
}

func commandHandler(c *gin.Context) {
	adc, err := strconv.ParseUint(c.Param("adc"), 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}
	switch c.Param("cmd") {
	case "ChStandby":
		opts := driver.ChStandbyOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChStandby(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.ChStandby(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeA":
		opts := driver.ChModeOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChModeA(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.ChModeA(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeB":
		opts := driver.ChModeOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChModeB(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.ChModeB(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeSel":
		opts := driver.ChModeSelectOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChModeSel(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.ChModeSel(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PowerMode":
		opts := driver.PowerModeOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PowerMode(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.PowerMode(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GeneralConf":
		opts := driver.GeneralConfOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GeneralConf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.GeneralConf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DataControl":
		opts := driver.DataControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DataControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.DataControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "InterfaceConf":
		opts := driver.InterfaceConfOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.InterfaceConf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.InterfaceConf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "BISTControl":
		opts := driver.BISTControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.BISTControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.BISTControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DeviceStatus":
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DeviceStatus(uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.DeviceStatus(i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "RevisionID":
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.RevisionID(uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.RevisionID(i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOControl":
		opts := driver.GPIOControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GPIOControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.GPIOControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOWriteData":
		opts := driver.GPIOWriteDataOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GPIOWriteData(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.GPIOWriteData(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOReadData":
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GPIOReadData(uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.GPIOReadData(i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PrechargeBuffer1":
		opts := driver.PreChargeBufferOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PrechargeBuffer1(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.PrechargeBuffer1(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PrechargeBuffer2":
		opts := driver.PreChargeBufferOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PrechargeBuffer2(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.PrechargeBuffer2(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PositiveRefPrechargeBuf":
		opts := driver.ReferencePrechargeBufOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PositiveRefPrechargeBuf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.PositiveRefPrechargeBuf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "NegativeRefPrechargeBuf":
		opts := driver.ReferencePrechargeBufOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.NegativeRefPrechargeBuf(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.NegativeRefPrechargeBuf(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChannelOffset":
		opts := driver.ChannelOffsetOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			err := adcConnection.ChannelOffset(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, nil)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := adcConnection.ChannelOffset(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusOK, nil)
		return

	case "ChannelGain":
		opts := driver.ChannelGainOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			err := adcConnection.ChannelGain(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, nil)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := adcConnection.ChannelGain(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
	case "ChannelSyncOffset":
		opts := driver.ChannelSyncOffsetOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			err := adcConnection.ChannelSyncOffset(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			c.JSON(http.StatusOK, nil)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := adcConnection.ChannelSyncOffset(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
	case "DiagnosticRX":
		opts := driver.DiagnosticRXOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DiagnosticRX(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.DiagnosticRX(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DiagnosticMuxControl":
		opts := driver.DiagnosticMuxControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DiagnosticMuxControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.DiagnosticMuxControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ModulatorDelayControl":
		opts := driver.ModulatorDelayControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ModulatorDelayControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.ModulatorDelayControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChopControl":
		opts := driver.ChopControlOpts{}
		if err := c.BindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChopControl(opts, uint8(adc))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"tx": tx,
				"rx": rx,
			})
			return
		}
		txResp := make([][]byte, 0, 9)
		rxResp := make([][]byte, 0, 9)
		for i := uint8(1); i < 10; i++ {
			tx, rx, err := adcConnection.ChopControl(opts, i)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		c.JSON(http.StatusOK, gin.H{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "HardReset":
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "command not found",
		})
		return
	}
}

func readDataHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket creation error: ", err)
		return
	}
	var file string
	if file = c.Query("file"); file == "" {
		_ = conn.WriteJSON(gin.H{
			"err": "invalid query parameter value",
		})
		_ = conn.Close()
		return
	}
	log.Println(file)

	f, err := dataFS.Open(file)
	if err != nil {
		_ = conn.WriteJSON(gin.H{
			"err": fmt.Errorf("failed to open file: %v", err),
		})
		_ = conn.Close()
		return
	}
	b, _ := ioutil.ReadAll(f)

	for i := 0; i < len(b); i += 80 {
		_ = conn.WriteMessage(websocket.BinaryMessage, b[i:i+80])
	}
	_ = conn.Close()
}

func readDataPostHandler(c *gin.Context) {
	form := struct {
		File string `json:"file"`
	}{}

	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed with error",
		})
		return
	}

	info, err := dataFS.Stat(form.File)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": fmt.Errorf("failed to open file: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"size": info.Size() / 80,
	})
}

func getFileHandler(c *gin.Context) {
	c.FileAttachment(path.Base(dataFile.Name()), dataFile.Name())
}

func samplingStatusHandler(c *gin.Context) {
	if sigrokRunning {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "running",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "not running",
	})
}

func loginOptionsHandler(c *gin.Context) {
	//c.Header("Allow", "OPTIONS, POST")
}

func getAllUSB(c *gin.Context) {
	var allDevices []usbDevice
	status := <-cmd.NewCmd("lsblk", "-o", "NAME,LABEL,SIZE,MOUNTPOINT", "-J").Start()
	str := strings.Builder{}
	for _, s := range status.Stdout {
		str.WriteString(s)
	}

	v := viper.New()

	v.SetConfigType("json")
	_ = v.ReadConfig(strings.NewReader(str.String()))
	if err := v.UnmarshalKey("blockdevices", &allDevices); err != nil {
		log.Printf("unmarshal error: %v", err)
	}

	//for i, dev := range allDevices {
	//	for _, child := range dev.Children {
	//		if match, _ := regexp.MatchString(`^/media.*`, child.MountPoint); match {
	//			connectedUSB[i] = child
	//		}
	//	}
	//}

	c.JSON(http.StatusOK, gin.H{
		"devices": connectedUSB,
	})
}
