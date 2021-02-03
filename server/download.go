package server

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
	"regexp"
	"strings"
	"time"

	"github.com/MShoaei/quakeADC/seg2"
	"github.com/gin-gonic/gin"
	"github.com/go-cmd/cmd"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func (s *server) DownloadSampleHandler(c *gin.Context) {
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

	dir, err := afero.IsDir(s.dataFS, "/"+c.Param("path"))
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

	requestedFile, err := s.dataFS.Open("/" + c.Param("path"))
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
		f, err := s.memFS.Create(requestedFile.Name() + ".DAT")
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
		var fs http.FileSystem = afero.NewHttpFs(s.memFS)
		c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(f.Name())))
		c.FileFromFS(f.Name(), fs)
		s.memFS.Remove(f.Name())
		return
	case "raw":
		var fs http.FileSystem = afero.NewHttpFs(s.dataFS)
		c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(requestedFile.Name())+".RAW"))
		c.FileFromFS(requestedFile.Name(), fs)
		return
	}
}

func (s *server) GetAllUSBHandler(c *gin.Context) {
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
	var header HeaderData
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

func (s *server) SaveSampleFile(c *gin.Context) {
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

	dir, err := afero.IsDir(s.dataFS, data.File)
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

	requestedFile, err := s.dataFS.Open(data.File)
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

func (s *server) SaveProjectFolder(c *gin.Context) {
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

	if exists, _ := afero.DirExists(s.dataFS, data.Project); !exists {
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
	err = afero.Walk(s.dataFS, path.Dir(data.Project+"/"), func(srcPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch mode := f.Mode(); {
		case mode.IsRegular():
			src, _ := s.dataFS.Open(srcPath)
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
