package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Create Image
func CreateImage(c *gin.Context) {
	dockerfileContent, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": "Failed to get image request",
		})
		glog.Error("Failed to get image request")
		return
	}
	glog.Info("Succeed to get request to create image")

	var imageData data.ImageData
	err = json.Unmarshal(dockerfileContent, &imageData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to unmarshal image request, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to unmarshal image request, the error is %v", err.Error())
		return
	}

	dockerfile := imageData.Dockerfile
	dstFilepath := imageData.Dstpath
	imageName := imageData.Imagename

	filePath := filepath.Join(dstFilepath, "dockerfile")
	err = ioutil.WriteFile(filePath, []byte(dockerfile), 0644)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to write dockerfile, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to write dockerfile, the error is %v", err.Error())
		return
	}

	cmd := "docker"
	glog.Info(filePath)
	_, err = tools.ExecCommand(cmd, "build", "-t", imageName, "-f", filePath, ".")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Failed to exec docker build, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to exec docker build, the error is %v", err.Error())
		return
	}

	_, err = tools.ExecCommand(cmd, "push", imageName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Failed to psuh docker build, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to push docker build, the error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    data.SUCCESS,
		"message: ": fmt.Sprintf("Succeed to build image: %v", imageName),
	})
	glog.Infof("Succeed to build image: %v", imageName)
}
