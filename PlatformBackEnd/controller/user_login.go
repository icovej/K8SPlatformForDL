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
			"message": fmt.Sprintf("Failed to load saved users info, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to load saved users info, the error is %v", err.Error())
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

func GetUserInfo_WithoutToken(c *gin.Context) {

	j := tools.NewJWT()
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": "Failed to get token, because the token is empty!",
		})
		glog.Error("Failed to get token, because the token is empty!")
		return
	}
	token, err := j.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
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
}
