package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func Login(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Method RegisterHandler gets invalid request payload, err is %v", err.Error()),
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

	var flag = 1

	for i := range users {
		if user.Username == users[i].Username && user.Password == users[i].Password {
			tools.GenerateToken(c, users[i])
			flag = 0
			break
		}
	}

	if flag == 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": "Invalid credentials",
		})
	}
}

func GetUserInfo_NoToken(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Method GetUserInfo_NoToken gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method GetUserInfo_NoToken gets invalid request payload")
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

	for i := range users {
		if user.Username == users[i].Username && user.Password == users[i].Password {
			c.JSON(http.StatusOK, gin.H{
				"code": data.SUCCESS,
				"data": users[i],
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.OPERATION_FAILURE,
		"message": "Failed to get user info!",
	})
	glog.Errorf("Failed to get user info without token!")
}
