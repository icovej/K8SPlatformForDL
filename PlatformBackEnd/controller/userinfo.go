package controller

import (
	"PlatformBackEnd/data"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func GetAllUsers(c *gin.Context) {
	path := "/home/gpu-server/all_test/biyesheji/PlatformBackEnd/User.json"
	file, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to read user file, err is %v", err.Error()),
		})
		glog.Error("Failed to read user file")
		return
	}

	var users [](data.User)
	err = json.Unmarshal(file, &users)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to unmarshal user file, err is %v", err.Error()),
		})
		glog.Error("Failed to unmarshal user file")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"data": users,
	})

}
