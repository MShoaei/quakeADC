package server

import (
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
)

func (s *server) TreeHandler(c *gin.Context) {
	list, err := afero.ReadDir(s.dataFS, "/"+c.Param("dir"))
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

func (s *server) TreeDeleteHandler(c *gin.Context) {
	if err := s.dataFS.RemoveAll(c.Param("path")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (s *server) TreePatchHandler(c *gin.Context) {
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
	if err := s.dataFS.Rename(p, path.Join(path.Dir(p), patchReq.NewName)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (s *server) GetFileHandler(c *gin.Context) {
	c.FileAttachment(path.Base(s.dataFile.Name()), s.dataFile.Name())
}

func (s *server) CreateNewProject(c *gin.Context) {
	data := struct {
		Name string `json:"name" bind:"required"`
	}{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := s.dataFS.Mkdir(data.Name, os.ModeDir|0755)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
