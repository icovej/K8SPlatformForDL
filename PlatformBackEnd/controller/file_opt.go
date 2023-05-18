package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func GetAllFiles(c *gin.Context) {
	var Path data.DirData
	err := c.ShouldBindJSON(&Path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method GetAllFiles gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method GetAllFiles gets invalid request payload")
		return
	}
	glog.Infof("Succeed to get request to get path %v all files", Path.Dir)

	path := Path.Dir

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
	token, err := j.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	path = token.Path + "/" + path
	glog.Infof("path is %v", path)

	var file_result []string
	var dir_result []string

	err = filepath.Walk(path, func(path string, file os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if file.IsDir() {
			dir_result = append(dir_result, file.Name())
		} else {
			file_result = append(file_result, file.Name())
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to read path %v, the error is %v", path, err.Error()),
		})
		glog.Errorf("Failed to read path %v, the error is %v", path, err.Error())
		return
	}

	result := data.FileData{
		Dir:  dir_result,
		File: file_result,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"data": result,
	})
	glog.Info("Succeed to get all files")
}

func DeleteFile(c *gin.Context) {
	var Path data.DirData
	err := c.ShouldBindJSON(&Path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method DeleteFile gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method DeleteFile gets invalid request payload")
		return
	}
	glog.Info("Succeed to get request to delete file %v", Path.Dir)

	j := tools.NewJWT()
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.IDENTITY_FAILURE,
			"message": "Failed to get token, because the token is empty!",
		})
		glog.Error("Failed to get token, because the token is empty!")
		return
	}
	token, err := j.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	path := token.Path + "/" + Path.Dir
	glog.Infof("path is %v", path)

	err = tools.DeleteFile_Dir(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to delete target %v, the error is %v", path, err.Error()),
		})
		glog.Errorf("Failed to delete target %v, the error is %v", path, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to delete file %v", path),
	})
	glog.Info("Succeed to delete file %v", Path.Dir)
}

func CreateDir(c *gin.Context) {
	var Dir data.DirData
	err := c.ShouldBindJSON(&Dir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method CreateDir gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method CreateDir gets invalid request payload")
		return
	}
	glog.Info("Succeed to get request to create dir %v", Dir.Dir)

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
	token, err := j.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	err = tools.CreatePath(token.Path+"/"+Dir.Dir, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to create path %v, the error is %v", Dir.Dir, err.Error()),
		})
		glog.Errorf("Failed to create path %v, the error is %v", Dir.Dir, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to create dir %v", Dir.Dir),
	})
	glog.Info("Succeed to create dir %v", Dir.Dir)
}
