package server

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
)

func (s *Server) TreeHandler(c *gin.Context) {
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

func (s *Server) TreeDeleteHandler(c *gin.Context) {
	if err := s.dataFS.RemoveAll(c.Param("path")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// TreePatchHandler renames a file or directory with the NewName
func (s *Server) TreePatchHandler(c *gin.Context) {
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

func (s *Server) GetFileHandler(c *gin.Context) {
	c.FileAttachment(path.Base(s.dataFile.Name()), s.dataFile.Name())
}

func (s *Server) CreateNewProject(c *gin.Context) {
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

func (s *Server) GetActiveProjectPath(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"path": s.activePath,
	})
}

func (s *Server) SetActiveProjectPath(c *gin.Context) {
	data := struct {
		Path string `json:"path"`
	}{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	p, err := s.dataFS.(*afero.BasePathFs).RealPath(data.Path)
	if err != nil {
		s.l.Debugf("invalid path result: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	s.activePath = filepath.Clean(data.Path)
	s.activeFS = afero.NewBasePathFs(afero.NewOsFs(), p)

	c.JSON(http.StatusOK, gin.H{})
}
