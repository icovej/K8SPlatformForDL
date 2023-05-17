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
			"message": fmt.Sprintf("Method Login gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method Login gets invalid request payload")
		return
	}
	glog.Infof("Succeed to get request to login, user name is %v", user.Username)

	users, err := tools.LoadUsers(data.UserFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to load saved users info, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to load saved users info, the error is %v", err.Error())
		return
	}
	glog.Info("Succeed to read user file")

	var flag = 1

	for i := range users {
		if user.Username == users[i].Username && user.Password == users[i].Password {
			glog.Info("Succed to match login's info with local user info, start to gernerate token")
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
	glog.Info("Failed to login, the error is Invalid credentials")
}

func GetUserInfo_NoToken(c *gin.Context) {
	glog.Info("Succeed to get request to get user info with no token")
	j := tools.NewJWT()
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.IDENTITY_FAILURE,
			"message": "Failed to get token, the error is Token is empty!",
		})
		glog.Error("Failed to get token, the error is Token is empty!")
		return
	}
	token, err := j.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.IDENTITY_FAILURE,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	user := data.User{
		Username: token.Username,
		Path:     token.Path,
		Role:     token.Role,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"data": user,
	})
	glog.Info("Succeed to get users' info without token")
}
