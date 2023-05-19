package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func GetGPUShareData(c *gin.Context) {
	var pod data.PodData
	err := c.ShouldBindJSON(&pod)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method GetGPUShareData gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method GetGPUShareData gets invalid request payload")
		return
	}
	glog.Info("Succeed to get request to get gpu_share data")

	podlist, err := tools.GetGPUData(pod)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to get gpu pod list, theerror is %v", err.Error()),
		})
		glog.Error("Failed to get gpu pod list, theerror is %v", err.Error())
		return
	}

	d, _ := os.ReadFile(data.PodFile)
	var pusers []data.PodUser
	_ = json.Unmarshal(d, &pusers)

	for i := 0; i < len(podlist); i++ {
		for j := 0; j < len(pusers); j++ {
			if pusers[j].PodName == podlist[i].Name {
				podlist[i].Username = pusers[j].UserName
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": "Succeed to get gpu pod lis",
		"data":    podlist,
	})
	glog.Info("Succeed to get gpu pod lis")

}
