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

func GetAllUsers(c *gin.Context) {
	glog.Info("Succeed to get request to get all users' info")
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
	glog.Info("Succeed to read user file")

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
	glog.Info("Succeed to get all users' info")
}

func ModifyUser(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Method ModifyUser gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method ModifyUser gets invalid request payload, the error is %v", err.Error())
		return
	}
	glog.Infof("Succeed to get request to modify user %v info", user.Username)

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

	for i := range users {
		if user.Username == users[i].Username {
			glog.Info("Succed to match login's info with local user info, start to gernerate token")

			old_path := users[i].Path
			users[i].Path = user.Path

			err = tools.CreateUserPath(users[i].Path)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"code":    data.OPERATION_FAILURE,
					"message": fmt.Sprintf("Failed to create user's new path %v, the error is %v", users[i].Path, err.Error()),
				})
				glog.Errorf("Failed to create user's new path %v, the error is %v", users[i].Path, err.Error())
			}

			err = tools.DeleteFile_Dir(old_path)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"code":    data.OPERATION_FAILURE,
					"message": fmt.Sprintf("Failed to delete user's old path %v, the error is %v", users[i].Path, err.Error()),
				})
				glog.Errorf("Failed to delete user's old path %v, the error is %v", users[i].Path, err.Error())
			}

			break
		}
	}

	err = tools.WriteUsers(users)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to update users file, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to update users file, the error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to update user file, user.name = %v", user.Username),
	})
	glog.Infof("Succeed to update user file, user name is %v, path is %v", user.Username, user.Path)

}
