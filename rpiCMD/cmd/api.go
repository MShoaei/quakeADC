package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MShoaei/quakeADC/driver"
	"github.com/go-cmd/cmd"
	"github.com/gorilla/websocket"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
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
func NewAPI() *iris.Application {
	api := iris.Default()
	if debug {
		api.Any("/api/{path:path}", func(ctx iris.Context) {
			r := ctx.Request()
			r.URL.Path = strings.Replace(r.URL.Path, "/api", "", 1)
			ctx.Application().ServeHTTPC(ctx)
		})
	}
	api.Options("/login", loginOptionsHandler)

	api.Use(cors.AllowAll())

	api.Get("/", homeHandler)

	api.Get("/tree/{dir:path}", treeHandler)

	api.Get("/plot", readDataHandler)
	api.Post("/plot", readDataPostHandler)

	api.Get("/plot/{file:path}", plotHandler)

	api.Post("/setup", setupHandler)
	api.Options("/command", homeHandler)
	api.Post("/command/{cmd:string}/all", commandHandler)
	api.Post("/command/{cmd:string}/{adc:uint8}", commandHandler)
	api.Get("/getfile", getFileHandler)

	api.Patch("/update", updateStack)

	api.Get("/usb/all", getAllUSB)
	api.Post("/rpi/shutdown", shutdownSequenceHandler)

	return api
}

func shutdownSequenceHandler(ctx iris.Context) {

}

func plotHandler(ctx iris.Context) {
	dir, err := afero.IsDir(dataFS, "/"+ctx.Params().Get("file"))
	if err != nil {
		ctx.StatusCode(iris.StatusNotFound)
		_, _ = ctx.JSON(iris.Map{
			"error": err,
		})
		return
	}
	if dir {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": err,
		})
		return
	}
	f, _ := dataFS.Open("/" + ctx.Params().Get("file"))
	stat, _ := f.Stat()
	ctx.StatusCode(iris.StatusOK)
	ctx.ServeContent(f, f.Name()+".bin", stat.ModTime())
}

func treeHandler(ctx iris.Context) {
	list, err := afero.ReadDir(dataFS, "/"+ctx.Params().Get("dir"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
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
	ctx.StatusCode(iris.StatusOK)
	_, _ = ctx.JSON(iris.Map{
		"directory": "/" + ctx.Params().Get("dir"),
		"items":     fd,
	})
}

func setupHandler(ctx iris.Context) {
	if sigrokRunning {
		ctx.StatusCode(iris.StatusServiceUnavailable)
		_, _ = ctx.JSON(iris.Map{
			"error": "sampling is already running",
		})
		return
	}
	setupData := struct {
		Channels []string `json:"channels"`
		Gains    []struct {
			Ch    string `json:"ch"`
			Value int    `json:"value"`
		}
		StartMode    string `json:"startMode"`
		RecordTime   int    `json:"recordTime"`
		SamplingTime int    `json:"samplingTime"`
		FileName     string `json:"fileName"`
		ProjectName  string `json:"projectName"`
	}{}
	if err := ctx.ReadJSON(&setupData); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": err,
		})
		return
	}

	if setupData.FileName == "" || setupData.ProjectName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": "invalid file name or project name",
		})
		return
	}
	if err := dataFS.MkdirAll(setupData.ProjectName, os.ModeDir|0755); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": "invalid file project path",
		})
		return
	}

	if exists, _ := afero.Exists(dataFS, filepath.Join(setupData.ProjectName, setupData.FileName)); exists {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": "file already exists",
		})
		return
	}

	dataFile, _ = dataFS.Create(filepath.Join(setupData.ProjectName, setupData.FileName))
	SendSyncSignal()
	if err := execSigrokCLI(setupData.RecordTime); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.JSON(iris.Map{
			"error": err,
		})
		return
	}
	ctx.StatusCode(iris.StatusOK)
}

func updateStack(_ iris.Context) {
	//TODO: How to self update?
}

func commandHandler(ctx iris.Context) {
	switch ctx.Params().GetString("cmd") {
	case "ChStandby":
		opts := driver.ChStandbyOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChStandby(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeA":
		opts := driver.ChModeOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChModeA(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeB":
		opts := driver.ChModeOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChModeB(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChModeSel":
		opts := driver.ChModeSelectOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChModeSel(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PowerMode":
		opts := driver.PowerModeOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PowerMode(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GeneralConf":
		opts := driver.GeneralConfOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GeneralConf(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DataControl":
		opts := driver.DataControlOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DataControl(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "InterfaceConf":
		opts := driver.InterfaceConfOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.InterfaceConf(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "BISTControl":
		opts := driver.BISTControlOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.BISTControl(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DeviceStatus":
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DeviceStatus(adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "RevisionID":
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.RevisionID(adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOControl":
		opts := driver.GPIOControlOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GPIOControl(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOWriteData":
		opts := driver.GPIOWriteDataOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GPIOWriteData(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "GPIOReadData":
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.GPIOReadData(adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PrechargeBuffer1":
		opts := driver.PreChargeBufferOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PrechargeBuffer1(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PrechargeBuffer2":
		opts := driver.PreChargeBufferOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PrechargeBuffer2(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "PositiveRefPrechargeBuf":
		opts := driver.ReferencePrechargeBufOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.PositiveRefPrechargeBuf(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "NegativeRefPrechargeBuf":
		opts := driver.ReferencePrechargeBufOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.NegativeRefPrechargeBuf(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChannelOffset":
		opts := driver.ChannelOffsetOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			err := adcConnection.ChannelOffset(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := adcConnection.ChannelOffset(opts, i)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
		}
		ctx.StatusCode(iris.StatusOK)
		return
	case "ChannelGain":
		opts := driver.ChannelGainOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			err := adcConnection.ChannelGain(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := adcConnection.ChannelGain(opts, i)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
		}
		ctx.StatusCode(iris.StatusOK)
		return
	case "ChannelSyncOffset":
		opts := driver.ChannelSyncOffsetOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			err := adcConnection.ChannelSyncOffset(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			return
		}
		for i := uint8(1); i < 10; i++ {
			err := adcConnection.ChannelSyncOffset(opts, i)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
		}
		ctx.StatusCode(iris.StatusOK)
		return
	case "DiagnosticRX":
		opts := driver.DiagnosticRXOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DiagnosticRX(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "DiagnosticMuxControl":
		opts := driver.DiagnosticMuxControlOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.DiagnosticMuxControl(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ModulatorDelayControl":
		opts := driver.ModulatorDelayControlOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ModulatorDelayControl(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "ChopControl":
		opts := driver.ChopControlOpts{}
		if err := ctx.ReadJSON(&opts); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"error": err,
			})
			return
		}
		if adc := ctx.Params().GetUint8Default("adc", 0); adc != 0 && adc < 10 {
			tx, rx, err := adcConnection.ChopControl(opts, adc)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(iris.Map{
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
				ctx.StatusCode(iris.StatusInternalServerError)
				_, _ = ctx.JSON(iris.Map{
					"error": err,
				})
				return
			}
			txResp = append(txResp, tx)
			rxResp = append(rxResp, rx)
		}
		ctx.StatusCode(iris.StatusOK)
		_, _ = ctx.JSON(iris.Map{
			"tx": txResp,
			"rx": rxResp,
		})
		return
	case "HardReset":
	default:
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": "command not found",
		})
		return
	}
}

func readDataHandler(ctx iris.Context) {
	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Println("WebSocket creation error: ", err)
		return
	}

	form := struct {
		File string `json:"file"`
	}{}
	if err := ctx.ReadQuery(&form); err != nil {
		_ = conn.WriteJSON(iris.Map{
			"err": err,
		})
		_ = conn.Close()
		return
	}
	log.Println(form.File)

	f, err := dataFS.Open(form.File)
	if err != nil {
		_ = conn.WriteJSON(iris.Map{
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

func readDataPostHandler(ctx iris.Context) {
	form := struct {
		File string `json:"file"`
	}{}

	if err := ctx.ReadJSON(&form); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"message": "failed with error",
		})
		return
	}

	info, err := dataFS.Stat(form.File)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"err": fmt.Errorf("failed to open file: %v", err),
		})
		return
	}

	ctx.StatusCode(iris.StatusOK)
	_, _ = ctx.JSON(iris.Map{
		"size": info.Size() / 80,
	})
}

func getFileHandler(ctx iris.Context) {
	err := ctx.SendFile(dataFile.Name(), path.Base(dataFile.Name()))
	if err != nil {
		log.Println("sending file failed with error: ", err)
	}
	return
}

func homeHandler(ctx iris.Context) {
	ctx.StatusCode(iris.StatusOK)
	_, _ = ctx.JSON(iris.Map{
		"message": "Home api",
	})
}

func loginOptionsHandler(ctx iris.Context) {
	ctx.Header("Allow", "OPTIONS, POST")
}

func getAllUSB(ctx iris.Context) {
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

	_, _ = ctx.JSON(iris.Map{
		"devices": connectedUSB,
	})
}
