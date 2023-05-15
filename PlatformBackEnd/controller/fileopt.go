package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func GetAllFiles(c *gin.Context) {
	path := c.PostForm("path")

	j := tools.NewJWT()
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": "Failed to get token, because the token is empty!",
		})
		glog.Error("Failed to get token, because the token is empty!")
		return
	}
	token, err := j.Parse_Token(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	path = token.Path + "/" + path

	files, err := os.ReadDir(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to read path %v, the error is %v", path, err.Error()),
		})
		glog.Errorf("Failed to read path %v, the error is %v", path, err.Error())
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

	j := tools.NewJWT()
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": "Failed to get token, because the token is empty!",
		})
		glog.Error("Failed to get token, because the token is empty!")
		return
	}
	token, err := j.Parse_Token(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	path = token.Path + "/" + path

	fi, err := os.Stat(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to stat file/dir %v, the error is %v", path, err.Error()),
		})
		glog.Errorf("Failed to stat file/dir %v, the error is %v", path, err.Error())
		return
	}
	if fi.IsDir() {
		err = os.RemoveAll(path)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": fmt.Sprintf("Failed to remove dir %v, the error is %v", path, err.Error()),
			})
			glog.Errorf("Failed to remove dir %v, the error is %v", path, err.Error())
			return
		}
	} else {
		err = os.Remove(path)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": fmt.Sprintf("Failed to remove file %v, the error is %v", path, err.Error()),
			})
			glog.Errorf("Failed to remove file %v, the error is %v", path, err.Error())
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to delete file %v", path),
	})
}
