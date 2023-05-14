package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Create Image
func CreateImage(c *gin.Context) {
	var image_data data.ImageData
	// Parse data that from front-end
	err := c.ShouldBindJSON(&image_data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method CreateImage gets invalid request payload")
		return
	}

	// Create user's dockerfile
	dstFilepath := image_data.Dstpath

	err = tools.CopyFile(data.Srcfilepath, dstFilepath)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to create dockerfile, the error is %v", err)
		return
	}

	// Import OS used in user's pod
	osVersion := image_data.Osversion
	statement := "FROM " + osVersion + "\n"
	err = tools.WriteAtBeginning(dstFilepath, []byte(statement))
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": err.Error(),
		})
		glog.Errorf("Failed to write osVersion to dockerfile, the error is %v", err)
		return
	}

	// Import python used in user's pod
	pyVersion := image_data.Pythonversion
	err = tools.WriteAtTail(dstFilepath, pyVersion)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": err.Error(),
		})
		glog.Errorf("Failed to write PyVersion to dockerfile, the error is %v", err)
		return
	}

	// Import images used in user's pod
	// And write into dockerfile whoes path is user's working path
	imageArray := image_data.Imagearray
	for i := range imageArray {
		err = tools.WriteAtTail(dstFilepath, imageArray[i])
		if err != nil {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"code":    data.API_PARAMETER_ERROR,
				"message": err.Error(),
			})
			glog.Errorf("Failed to write image to dockerfile, the error is %v", err)
			return
		}
	}

	// Create dockerfile
	imageName := image_data.Imagename
	cmd := "docker"
	_, err = tools.ExecCommand(cmd, "build", "-t", imageName, "-f", dstFilepath, ".")
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    data.API_PARAMETER_ERROR,
			"message: ": err.Error(),
		})
		glog.Errorf("Failed to exec docker build, the error is %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    data.SUCCESS,
		"message: ": fmt.Sprintf("Succeed to build image: %v", imageName),
	})
}
