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
	var Mpod data.Monitor
	err := c.ShouldBindJSON(&Mpod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

	podList, err := tools.GetAllPod(Mpod.Namespace)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err.Error(),
		})
		glog.Error("Failed to get pod info, the error is %v", err)
		return
	}

	nsList, err := tools.GetAllNamespace()
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err.Error(),
		})
		glog.Error("Failed to get all namespace, the error is %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":       http.StatusOK,
		"pods: ":       podList,
		"namespaces: ": nsList,
	})
}
