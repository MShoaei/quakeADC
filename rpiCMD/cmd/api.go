package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/MShoaei/quakeADC/driver/xmega"
	"github.com/MShoaei/quakeADC/seg2"
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

type headerData struct {
	EnabledChannels [24]bool   `json:"EnabledChannels"`
	Gains           [24]uint32 `json:"Gains"`
	Window          int        `json:"Window"`
}

var hd headerData = headerData{
	Gains: [24]uint32{
		1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1,
	},
}

type usbDevice struct {
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	MountPoint string      `json:"mountpoint"`
	Size       string      `json:"size"`
	Children   []usbDevice `json:"children"`
}

type RXResponse []byte

func (r RXResponse) MarshalJSON() ([]byte, error) {
	var result string
	if r == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", r)), ",")
	}
	return []byte(result), nil
}

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
	api.DELETE("/tree/*path", treeDeleteHandler)
	api.PATCH("/tree/*path", treePatchHandler)

	api.GET("/plot", readDataHandler)
	api.POST("/plot", readDataPostHandler)

	api.GET("/dl/*path", func(c *gin.Context) {
		c.Header("cache-control", "no-store, max-age=0")
	}, downloadSampleHandler)

	api.POST("/setup", setupHandler)
	api.POST("/command/:cmd/:adc", commandHandler)
	api.GET("/getfile", getFileHandler)

	api.PATCH("/update", updateStack)

	api.GET("/usb", getAllUSBHandler)

	api.POST("/rpi/shutdown", shutdownSequenceHandler)
	api.POST("/rpi/restart", restartSequenceHandler)
	api.GET("/channels", getChannelsHandler)
	api.POST("/channels", setChannelsHandler)
	api.GET("/gains", getGainsHandler)
	api.POST("/gains", setGainsHandler)
	api.GET("/info", boardInfoHandler)
	api.POST("/calibrate", func(c *gin.Context) {
		cilabrateChOffset()
		for i := 0; i < len(hd.EnabledChannels); i++ {
			hd.EnabledChannels[i] = true
			hd.Gains[i] = 1000
		}
		c.Status(http.StatusOK)
	})

	api.POST("/save/project", saveProjectFolder)
	api.POST("/save/sample", saveSampleFile)

	api.PATCH("/multiplier", func(c *gin.Context) {
		val, err := strconv.Atoi(c.Query("val"))
		if err != nil {
			c.String(http.StatusBadRequest, "%v", err)
		}
		gainMultiply = uint32(val)
		c.String(http.StatusOK, "%d", gainMultiply)
	})
	return api
}

func downloadSampleHandler(c *gin.Context) {
	fileType, exists := c.GetQuery("type")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "requested file type not specified",
		})
		return
	}
	fileType = strings.ToLower(fileType)
	switch fileType {
	case "seg2", "raw":
		break
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file type",
		})
		return
	}

	dir, err := afero.IsDir(dataFS, "/"+c.Param("path"))
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

	requestedFile, err := dataFS.Open("/" + c.Param("path"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer requestedFile.Close()

	switch fileType {
	case "seg2":
		byteRes, err := extractData(requestedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		traces := seg2.NewTraceDescriptor(make([]string, len(byteRes)), byteRes, seg2.Fixed32)
		w := seg2.NewWriter(time.Now(), int16(len(traces)), "")
		f, err := memFS.Create(requestedFile.Name() + ".DAT")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer f.Close()

		err = w.Write(f, traces)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		var fs http.FileSystem = afero.NewHttpFs(memFS)
		c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(f.Name())))
		c.FileFromFS(f.Name(), fs)
		memFS.Remove(f.Name())
		return
	case "raw":
		var fs http.FileSystem = afero.NewHttpFs(dataFS)
		c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(requestedFile.Name())+".RAW"))
		c.FileFromFS(requestedFile.Name(), fs)
		return
	}
}

func getAllUSBHandler(c *gin.Context) {
	devices, err := getAllUSB()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

func extractData(src io.Reader) ([][]byte, error) {
	var header headerData
	b, _ := ioutil.ReadAll(src)
	infoBytes, _ := bufio.NewReader(bytes.NewReader(b)).ReadBytes('\n')
	if err := json.Unmarshal(infoBytes, &header); err != nil {
		return nil, err
	}
	b = b[len(infoBytes):]
	count := 0
	for i := 0; i < len(header.EnabledChannels); i++ {
		if !header.EnabledChannels[i] {
			continue
		}
		count++
	}

	res := make([][]byte, count)
	for i := 0; i < len(res); i++ {
		res[i] = make([]byte, 0, len(b)/4)
	}

	for i := 0; i < count; i++ {
		for j := i * 4; j < len(b); j += count * 4 {
			res[i] = append(res[i], b[j], b[j+1], b[j+2], b[j+3])
		}
	}
	return res, nil
}

func saveSampleFile(c *gin.Context) {
	const pathPrefix = "HITECH"
	data := struct {
		File string `json:"file"`
	}{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	dir, err := afero.IsDir(dataFS, data.File)
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

	fileType, exists := c.GetQuery("type")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "requested file type not specified",
		})
		return
	}
	var fileExtension string
	fileType = strings.ToLower(fileType)
	switch fileType {
	case "seg2":
		fileExtension = ".DAT"
	case "raw":
		fileExtension = ".RAW"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file type",
		})
		return
	}

	connectedUSB, err := getAllUSB()
	if connectedUSB.MountPoint == "" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No USB connected",
		})
		return
	}

	requestedFile, err := dataFS.Open(data.File)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer requestedFile.Close()

	usbFS := afero.NewBasePathFs(afero.NewOsFs(), connectedUSB.MountPoint)
	_ = usbFS.Mkdir(pathPrefix, os.ModeDir|0755)
	if exists, _ := afero.DirExists(usbFS, pathPrefix); !exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "a file with name 'HITECH' exists on USB device.",
		})
		return
	}

	force := strings.ToLower(c.Query("force")) != "" && strings.ToLower(c.Query("force")) == "true"
	usbFS = afero.NewBasePathFs(usbFS, pathPrefix)
	if fileExists, _ := afero.Exists(usbFS, data.File+fileExtension); fileExists && !force {
		c.JSON(http.StatusConflict, gin.H{
			"error": "file exists and not forced",
		})
		return
	}
	usbFS.MkdirAll(path.Dir(data.File), os.ModeDir|0755)
	dst, err := usbFS.Create(data.File + fileExtension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer dst.Close()

	switch fileType {
	case "seg2":
		byteRes, err := extractData(requestedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		traces := seg2.NewTraceDescriptor(make([]string, len(byteRes)), byteRes, seg2.Fixed32)
		w := seg2.NewWriter(time.Now(), int16(len(traces)), "")
		_ = w.Write(dst, traces)
		c.JSON(http.StatusOK, gin.H{
			"files": []string{data.File},
		})
	case "raw":
		io.Copy(dst, requestedFile)
		c.JSON(http.StatusOK, nil)
	}
}

func saveProjectFolder(c *gin.Context) {
	const pathPrefix = "HITECH"
	data := struct {
		Project string `json:"project"`
	}{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileType, exists := c.GetQuery("type")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "requested file type not specified",
		})
		return
	}
	var fileExtension string
	fileType = strings.ToLower(fileType)
	switch fileType {
	case "seg2":
		fileExtension = ".DAT"
	case "raw":
		fileExtension = ".RAW"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file type",
		})
		return
	}

	connectedUSB, err := getAllUSB()
	if connectedUSB.MountPoint == "" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No USB connected",
		})
		return
	}

	if exists, _ := afero.DirExists(dataFS, data.Project); !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "requested folder does not exist",
		})
		return
	}

	usbFS := afero.NewBasePathFs(afero.NewOsFs(), connectedUSB.MountPoint)
	_ = usbFS.Mkdir(pathPrefix, os.ModeDir|0755)
	if exists, _ := afero.DirExists(usbFS, pathPrefix); !exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "a file with name 'HITECH' exists on USB device.",
		})
		return
	}

	force := strings.ToLower(c.Query("force")) != "" && strings.ToLower(c.Query("force")) == "true"
	usbFS = afero.NewBasePathFs(usbFS, pathPrefix)
	usbFS.MkdirAll(path.Dir(data.Project+"/"), os.ModeDir|0755)
	if empty, _ := afero.IsEmpty(usbFS, data.Project); !empty && !force {
		c.JSON(http.StatusConflict, gin.H{
			"error": "project path already exists and not forced",
		})
		return
	}

	copiedList := make([]string, 0)
	err = afero.Walk(dataFS, path.Dir(data.Project+"/"), func(srcPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch mode := f.Mode(); {
		case mode.IsRegular():
			src, _ := dataFS.Open(srcPath)
			dst, _ := usbFS.Create(srcPath + fileExtension)
			defer src.Close()
			defer dst.Close()

			switch fileType {
			case "seg2":
				byteRes, err := extractData(src)
				if err != nil {
					return err
				}
				traces := seg2.NewTraceDescriptor(make([]string, len(byteRes)), byteRes, seg2.Fixed32)
				w := seg2.NewWriter(time.Now(), int16(len(traces)), "")
				_ = w.Write(dst, traces)
			case "raw":
				_, err := io.Copy(dst, src)
				if err != nil {
					return err
				}
			}
			copiedList = append(copiedList, srcPath)

		case mode.IsDir():
			if err := usbFS.Mkdir(srcPath, f.Mode()); err != nil && !os.IsExist(err) {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"files": copiedList,
	})
}

func boardInfoHandler(c *gin.Context) {
	type message struct {
		Voltage []int16 `json:"voltage"`
		Current []int16 `json:"current"`
	}
	voltage := xmega.GetVoltage(adcConnection.Connection())
	current := xmega.GetCurrent(adcConnection.Connection())
	m := message{
		Voltage: voltage,
		Current: current,
	}
	c.JSON(http.StatusOK, &m)
}

var gainMultiply uint32 = 1000

func setGainsHandler(c *gin.Context) {
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
		val := gains[i] * gainMultiply
		opts.Offset[0] = uint8((val & MSBMask) >> 16)
		opts.Offset[1] = uint8((val & MidMask) >> 8)
		opts.Offset[2] = uint8(val & LSBMask)
		log.Println(opts.Offset)
		if _, err := adcConnection.ChannelGain(opts, uint8(i/8)+1, debug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	hd.Gains = gains
	c.JSON(http.StatusOK, nil)
}

func getGainsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"gains": hd.Gains,
	})
}

func restartSequenceHandler(c *gin.Context) {
	xmega.Reset()
	cmd.NewCmd("/usr/bin/sudo", "/sbin/shutdown", "-r", "now").Start()
}

func setChannelsHandler(c *gin.Context) {
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
	hd.EnabledChannels = [24]bool{}
	opts := driver.ChStandbyOpts{
		Write:    true,
		Channels: [8]bool{},
	}
	for i, enable := range ch {
		hd.EnabledChannels[i] = enable
		opts.Channels[i%8] = !enable
		if i%8 == 7 {
			if i < 8 {
				opts.Channels[0] = false
			}
			log.Println(opts, uint8(i/8)+1)
			adcConnection.ChStandby(opts, uint8(i/8)+1)
			opts.Channels = [8]bool{}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if force {
		hd.EnabledChannels[0] = false
	}
	c.Status(http.StatusOK)
}

func getChannelsHandler(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"channels": hd.EnabledChannels,
	})
}

func shutdownSequenceHandler(c *gin.Context) {
	xmega.Shutdown(adcConnection.Connection())
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

func treeDeleteHandler(c *gin.Context) {
	if err := dataFS.RemoveAll(c.Param("path")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func treePatchHandler(c *gin.Context) {
	patchReq := struct {
		NewName string `json:"newName"`
	}{}
	if err := c.BindJSON(&patchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	p := c.Param("path")
	if err := dataFS.Rename(p, path.Join(path.Dir(p), patchReq.NewName)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
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
		StartMode        string  `json:"startMode"`
		TriggerThreshold int     `json:"threshold"`
		TriggerChannel   int     `json:"triggerChannel"`
		RecordTime       int     `json:"recordTime"`
		SamplingTime     float32 `json:"samplingTime"`
		Window           int     `json:"window"`
		FileName         string  `json:"fileName"`
		ProjectName      string  `json:"projectName"`
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
	hd.Window = setupData.Window

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
	defer dataFile.Close()

	switch strings.ToLower(setupData.StartMode) {
	case "asap":
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
		return
	case "hammer":
		configSamplingTime(setupData.SamplingTime)
		SendSyncSignal()
		xmega.SamplingStart(adcConnection.Connection())
		defer xmega.SamplingEnd(adcConnection.Connection())
		rawData := readWithThreshold(setupData.TriggerThreshold, setupData.RecordTime, setupData.TriggerChannel)
		if rawData == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "did not reach threshold",
			})
			return
		}
		if err := json.NewEncoder(dataFile).Encode(hd); err != nil {
			// this should never happen!
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Errorf("error while encoding enabled channels: %v", err).Error(),
			})
			return
		}
		convert(bytes.NewReader(rawData), nil, nil, dataFile, len(rawData), hd.EnabledChannels)
		c.JSON(http.StatusOK, nil)
		return
	case "trigger":
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "not implemented",
		})
		return
	}
}

func updateStack(_ *gin.Context) {
	//TODO: How to self update?
}

func commandHandler(c *gin.Context) {
	adc, err := strconv.ParseUint(c.Param("adc"), 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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
			err := adcConnection.ChannelOffset(opts, uint8(adc), debug)
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
			err := adcConnection.ChannelOffset(opts, i, debug)
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
			_, err := adcConnection.ChannelGain(opts, uint8(adc), debug)
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
			_, err := adcConnection.ChannelGain(opts, i, debug)
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
			"error": "invalid query parameter value",
		})
		_ = conn.Close()
		return
	}

	f, err := dataFS.Open(file)
	if err != nil {
		_ = conn.WriteJSON(gin.H{
			"error": fmt.Errorf("failed to open file: %v", err),
		})
		_ = conn.Close()
		return
	}
	defer f.Close()

	var header headerData
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
			"error": fmt.Errorf("failed to open file: %v", err),
		})
		return
	}

	f, _ := dataFS.Open(form.File)
	b, _ := bufio.NewReader(f).ReadBytes('\n')
	var header headerData
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

func getAllUSB() (usbDevice, error) {
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

	var connectedUSB usbDevice
	for _, dev := range allDevices {
		for _, child := range dev.Children {
			if match, _ := regexp.MatchString(`^/(media|mnt/USB).*`, child.MountPoint); match {
				connectedUSB = child
				return connectedUSB, nil
			}
		}
	}

	if connectedUSB.MountPoint == "" {
		for _, dev := range allDevices {
			if strings.Contains(dev.Name, "mmcblk0") {
				continue
			}
			for _, child := range dev.Children {
				if child.Children != nil {
					continue
				}
				status = <-cmd.NewCmd("/usr/bin/sudo", "mount", "-o", "uid=pi,gid=pi", path.Join("/", "dev", child.Name), "/mnt/USB").Start()
				if status.Exit == 0 {
					child.MountPoint = "/mnt"
					connectedUSB = child
					return connectedUSB, nil
				}
			}
		}
	}
	if connectedUSB.MountPoint == "" {
		return connectedUSB, fmt.Errorf("USB not found")
	}
	return connectedUSB, status.Error
}

func init() {
	<-cmd.NewCmd("/usr/bin/sudo", "umount", "/mnt/USB").Start()
	_, _ = getAllUSB()
}
