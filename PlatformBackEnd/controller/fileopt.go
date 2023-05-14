package controller

import (
	"PlatformBackEnd/data"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func GetAllFiles(c *gin.Context) {
	path := c.PostForm("path")

	files, err := os.ReadDir(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  data.OPERATION_FAILURE,
			"error": err.Error(),
		})
		glog.Errorf("Failed to read path, the error is %v", path, err.Error())
		return
	}

	var result []map[string]interface{}

	for _, file := range files {
		info := make(map[string]interface{})
		info["name"] = file.Name()
		if file.IsDir() {
			info["type"] = "directory"
		} else {
			info["type"] = "file"
		}
		result = append(result, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"data": result,
	})
}

func DeleteFile(c *gin.Context) {
	path := c.PostForm("path")

	fi, err := os.Stat(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  data.OPERATION_FAILURE,
			"error": err.Error(),
		})
		glog.Error("Failed to stat file/dir")
		return
	}
	if fi.IsDir() {
		err = os.RemoveAll(path)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":  data.OPERATION_FAILURE,
				"error": err.Error(),
			})
			glog.Error("Failed to remove dir")
			return
		}
	} else {
		err = os.Remove(path)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":  data.OPERATION_FAILURE,
				"error": err.Error(),
			})
			glog.Error("Failed to remove file")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": "Succeed to delete target",
	})
}
