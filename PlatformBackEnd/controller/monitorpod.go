package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func MonitorPod(c *gin.Context) {
	var mpod data.Monitor
	err := c.ShouldBindJSON(&mpod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method GetDirInfo gets invalid request payload")
		return
	}

	podList, err := tools.GetAllPod(mpod.Namespace)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": err.Error(),
		})
		glog.Errorf("Failed to get pod info, the error is %v", err)
		return
	}

	nsList, err := tools.GetAllNamespace()
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": err.Error(),
		})
		glog.Errorf("Failed to get all namespace, the error is %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       data.SUCCESS,
		"message":    "Succeed to get pod info",
		"pods":       podList,
		"namespaces": nsList,
	})
}
