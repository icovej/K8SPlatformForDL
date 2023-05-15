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
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method RegisterHandler gets invalid request payload")
		return
	}

	users, err := tools.LoadUsers(data.UserFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Error("Failed to load saved users info")
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
			"message": err.Error(),
		})
		glog.Error("Failed to convert to user info")
		return
	}

	err = os.WriteFile(data.UserFile, update, 0644)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Error("Failed to re-write userfile after delete")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": "Succeed to delete user",
	})
}
