package cmd

import (
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-cmd/cmd"
	"github.com/gorilla/websocket"
	"github.com/iris-contrib/middleware/cors"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
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

var readParams struct {
	File     string `json:"file"`
	Skip     int    `json:"skip"`
	Duration int    `json:"duration"`
}

var sigrokRunning = false
var dataFile *os.File

// CommandsList is the list of all available commands
var CommandsList map[string]func(*flag.FlagSet) ([]byte, []byte, error)

var flagsList = map[string]*flag.FlagSet{}

//secret is the default secret for signing tokens
const secret = "g3%k2Qi2j8B*XZVLfhi#7UMjJQ$#anVH"

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
	api.Options("/login", loginOptionsHandler)
	api.Post("/login", loginPostHandler)

	authMiddleware := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(_ *jwt.Token) (interface{}, error) {
			if os.Getenv("SECRET") != "" {
				return []byte(os.Getenv("SECRET")), nil
			}
			return []byte(secret), nil
		},

		SigningMethod: jwt.SigningMethodHS512,
	})

	api.Use(cors.AllowAll())

	api.Get("/", homeHandler)

	api.Get("/readlive", readLiveHandler)
	api.Post("/readlive", readLivePostHandler)
	// api.Any("/readlive", readLiveHandler)

	api.Post("/setup", setupHandler)
	api.Options("/command", homeHandler)
	api.Post("/command", commandHandler)
	api.Get("/getfile", getFileHandler)

	api.Patch("/update", updateStack)

	api.Get("/usb/all", getAllUSB)

	api.Use(authMiddleware.Serve)

	return api
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
	//TODO: this should be changed to allow sub-projects.
	if strings.Contains(setupData.ProjectName, "/") || strings.Contains(setupData.FileName, "/") {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": "invalid file name or project name",
		})
		return
	}

	wd, _ := os.Getwd()
	if err := os.MkdirAll(filepath.Join(wd, setupData.ProjectName), os.ModeDir|0755); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": err,
		})
		return
	}

	_, err := os.Stat(filepath.Join(wd, setupData.ProjectName, setupData.FileName+".bin"))
	if !os.IsNotExist(err) {
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{
			"error": "file already exists",
		})
		return
	}
	dataFile, _ = os.Create(filepath.Join(wd, setupData.ProjectName, setupData.FileName+".bin"))
	execSigrokCLI(setupData.RecordTime)

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
			"message": "Unknown command",
			//"error":   err.Error(),
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
	log.Printf("tx: %v, rx: %v", tx, rx)
	_, _ = ctx.JSON(iris.Map{
		"message": "success",
		"tx":      fmt.Sprintf("%v", tx),
		"rx":      fmt.Sprintf("%v", rx),
	})
}

func readLiveHandler(ctx iris.Context) {
	log.Println("readLiveHandler called")
	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Println("WebSocket creation error: ", err)
		return
	}
	f, _ := os.Open("direct.bin")
	b, _ := ioutil.ReadAll(f)

	// w, _ := conn.NextWriter(websocket.BinaryMessage)
	now := time.Now()
	for i := 0; i < len(b); i += 80 {
		_ = conn.WriteMessage(websocket.BinaryMessage, b[i:i+80])
		//time.Sleep(100*time.Millisecond)
	}
	time.Sleep(1 * time.Second)
	_ = conn.Close()
	fmt.Println(time.Since(now), len(b)/80)
	//dataFile, err = os.OpenFile(path.Join("/", "tmp", readParams.File+".txt"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	//if err != nil {
	//	log.Println("failed to open file: ", err)
	//	return
	//}
	//rcvToSend := make(chan string)
	//go read(readOptions{
	//	file:     dataFile,
	//	skip:     readParams.Skip,
	//	duration: readParams.Duration,
	//	ch:       rcvToSend,
	//})
	//for {
	//	data, ok := <-rcvToSend
	//	if !ok {
	//		conn.WriteMessage(websocket.TextMessage, []byte("Send finished"))
	//		conn.Close()
	//		return
	//	}
	//	conn.WriteMessage(websocket.TextMessage, []byte(data))
	//}
}

func readLivePostHandler(ctx iris.Context) {
	info, _ := os.Stat("direct.bin")
	log.Println("readLivePostHandler called")
	_ = ctx.ReadJSON(&readParams)
	log.Println(readParams, info.Size()/80)
	_, _ = ctx.JSON(iris.Map{
		"code": 200,
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

func loginPostHandler(ctx iris.Context) {
	s := sha512.New()
	s.Write([]byte(ctx.FormValue("username")))

	var t *jwt.Token

	/*pass: @12348765@ */
	if ctx.FormValue("username") == "admin" && string(s.Sum(nil)) == "2871d000b43b5c6220e2a0e210966f5f8ce7ebbbc198eb0da7069aea08f4659160316a7e98b1d8bc86b949c693d1b561eecc05d4e67499bf490c3e30bd207588" {
		t = jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
			Id:        "1",
			ExpiresAt: time.Now().Add(8 * time.Hour).Unix(),
		})

		/*pass: randomuserpass*/
	} else if ctx.FormValue("username") == "user" && string(s.Sum(nil)) == "9412a466356c0fb54f742f0e39ac07677c41d6fb814344baed544db4f98ab1e00b74110e6f91a5f88ded89a9c0d0b5a5e382d0af708c0f8cbcd7ab62cbc13156" {
		t = jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
			Id:        "2",
			ExpiresAt: time.Now().Add(8 * time.Hour).Unix(),
		})
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		_, _ = ctx.JSON(iris.Map{"success": false, "error": "incorrect username and/or password"})
		return
	}

	var token string
	var err error
	if os.Getenv("SECRET") != "" {
		token, err = t.SignedString([]byte(os.Getenv("SECRET")))
	} else {
		token, err = t.SignedString([]byte(secret))
	}

	if err != nil {
		log.Printf("fialed to sign token: %v", err)
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.JSON(iris.Map{"success": false})
		return
	}
	ctx.StatusCode(iris.StatusOK)
	_, _ = ctx.JSON(iris.Map{"success": true, "token": token})
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
