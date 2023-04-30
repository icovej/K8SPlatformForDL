package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func MonitorPod(c *gin.Context) {
	var mpod data.Monitor
	err := c.ShouldBindJSON(&mpod)
	if err != nil {
		c.JSON(data.API_PARAMETER_ERROR, gin.H{
			"code: ":    data.API_PARAMETER_ERROR,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

	podList, err := tools.GetAllPod(mpod.Namespace)
	if err != nil {
		c.JSON(data.API_PARAMETER_ERROR, gin.H{
			"code: ":    data.API_PARAMETER_ERROR,
			"message: ": err.Error(),
		})
		glog.Error("Failed to get pod info, the error is %v", err)
		return
	}

	nsList, err := tools.GetAllNamespace()
	if err != nil {
		c.JSON(data.API_PARAMETER_ERROR, gin.H{
			"code: ":    data.API_PARAMETER_ERROR,
			"message: ": err.Error(),
		})
		glog.Error("Failed to get all namespace, the error is %v", err)
		return
	}

	c.JSON(data.SUCCESS, gin.H{
		"code: ":       data.SUCCESS,
		"message:":     "Succeed to get pod info",
		"pods: ":       podList,
		"namespaces: ": nsList,
	})
}
