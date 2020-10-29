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

	"github.com/go-cmd/cmd"
	"github.com/gorilla/websocket"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
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

// CommandsList is the list of all available commands
var CommandsList map[string]func(*flag.FlagSet) ([]byte, []byte, error)

var flagsList = map[string]*flag.FlagSet{}

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

	api.Get("/tree", func(ctx iris.Context) {
		list, err := afero.ReadDir(dataFS, "/")
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
			"directory": "/",
			"items":     fd,
		})
	})
	api.Get("/tree/{dir:path}", treeHandler)

	api.Get("/plot", readDataHandler)
	api.Post("/plot", readDataPostHandler)

	api.Get("/plot/{file:path}", plotHandler)

	api.Post("/setup", setupHandler)
	api.Options("/command", homeHandler)
	api.Post("/command", commandHandler)
	api.Get("/getfile", getFileHandler)

	api.Patch("/update", updateStack)

	api.Get("/usb/all", getAllUSB)

	return api
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
	data := &struct {
		Command string
		Flags   []struct {
			Name  string
			Value string
		}
	}{}

	err := ctx.ReadJSON(data)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Println(err)
		return
	}

	command := CommandsList[data.Command]
	if command == nil {
		ctx.StatusCode(iris.StatusNotImplemented)
		_, _ = ctx.JSON(iris.Map{
			"error": "Unknown command",
		})
		log.Printf("unknown command: %s", data.Command)
		return
	}

	set := flagsList[data.Command]
	if set == nil { // should never happen!
		ctx.StatusCode(iris.StatusNotImplemented)
		_, _ = ctx.JSON(iris.Map{
			"message": "flag set not found",
			//"error":   err.Error(),
		})
		log.Printf("unknown command: %s", data.Command)
		return
	}

	flags := make([]string, 0, 7)
	for _, f := range data.Flags {
		flags = append(flags, f.Name+"="+f.Value)
	}
	err = set.Parse(flags)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Printf("flag parse failed with error: %s", err)
		return
	}

	tx, rx, err := command(set)
	if err != nil {
		ctx.StatusCode(iris.StatusNotAcceptable)
		_, _ = ctx.JSON(iris.Map{
			"message": "failed with error",
			//"error":   err.Error(),
		})
		log.Printf("command %s failed: %s", data.Command, err)
		return
	}

	ctx.StatusCode(iris.StatusOK)
	_, _ = ctx.JSON(iris.Map{
		"tx": fmt.Sprintf("%v", tx),
		"rx": fmt.Sprintf("%v", rx),
	})
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
