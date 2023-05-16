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

func DeleteUser(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Method DeleteUser gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method DeleteUser gets invalid request payload")
		return
	}

	users, err := tools.LoadUsers(data.UserFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to load saved users info, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to load saved users info, the error is %v", err.Error())
		return
	}

	for i, u := range users {
		if u.Username == user.Username {
			users = append(users[:i], users[i+1:]...)
			break
		}
	}

	update, err := json.Marshal(users)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to convert to user info, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to convert to user info, the error is %v", err.Error())
		return
	}

	err = os.WriteFile(data.UserFile, update, 0644)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to re-write userfile after delete, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to re-write userfile after delete, the error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to delete user, user.name = %v", user.Username),
	})
}
