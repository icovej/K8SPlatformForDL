package controller

import (
	"fmt"
	"net/http"
	"platform_back_end/data"
	"platform_back_end/tools"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Create Image
func CreateImage(c *gin.Context) {
	var image_data data.ImageData
	// Parse data that from front-end
	err_bind := c.ShouldBindJSON(&image_data)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method CreateImage gets invalid request payload")
		return
	}

	// Create user's dockerfile
	dstFilepath := image_data.Dstpath

	err_create := tools.CopyFile(data.Srcfilepath, dstFilepath)
	if err_create != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_create.Error(),
		})
		glog.Error("Failed to create dockerfile, the error is %v", err_create)
		return
	}

	// Import OS used in user's pod
	osVersion := image_data.Osversion
	statement := "FROM " + osVersion + "\n"
	err_version := tools.WriteAtBeginning(dstFilepath, []byte(statement))
	if err_version != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_version.Error(),
		})
		glog.Error("Failed to write osVersion to dockerfile, the error is %v", err_version)
		return
	}

	// Import python used in user's pod
	pyVersion := image_data.Pythonversion
	err_py := tools.WriteAtTail(dstFilepath, pyVersion)
	if err_py != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_py.Error(),
		})
		glog.Error("Failed to write PyVersion to dockerfile, the error is %v", err_py)
		return
	}

	// Import images used in user's pod
	// And write into dockerfile whoes path is user's working path
	imageArray := image_data.Imagearray
	for i := range imageArray {
		err_image := tools.WriteAtTail(dstFilepath, imageArray[i])
		if err_image != nil {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"code: ":    http.StatusMethodNotAllowed,
				"message: ": err_image.Error(),
			})
			glog.Error("Failed to write image to dockerfile, the error is %v", err_image)
			return
		}
	}

	// Create dockerfile
	imageName := image_data.Imagename
	cmd := "docker"
	_, err_exec := tools.ExecCommand(cmd, "build", "-t", imageName, "-f", dstFilepath, ".")
	if err_exec != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_exec.Error(),
		})
		glog.Error("Failed to exec docker build, the error is %v", err_exec)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    http.StatusOK,
		"message: ": fmt.Sprintf("Succeed to build image: %v", imageName),
	})
	return
}
