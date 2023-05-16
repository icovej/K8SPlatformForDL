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
	path, _ := os.Getwd()
	path = path + "/" + data.UserFile
	file, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to read user file, err is %v", err.Error()),
		})
		glog.Errorf("Failed to read user file, the error is %v", err.Error())
		return
	}

	var users [](data.User)
	err = json.Unmarshal(file, &users)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to unmarshal user file, err is %v", err.Error()),
		})
		glog.Errorf("Failed to unmarshal user file, the error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"data": users,
	})

}
