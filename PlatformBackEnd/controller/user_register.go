package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func RegisterUser(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Method RegisterUser gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method RegisterUser gets invalid request payload, the error is %v", err.Error())
		return
	}

	glog.Infof("Succeed to get request to regite user %v", user.Username)

	// Check if the account has existed
	users, err := tools.CheckUsers()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to check user info, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to check user info, the error is %v", err.Error())
		return
	}
	for _, u := range users {
		if u.Username == user.Username {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "Failed to registe account, the error is username has already existed",
			})
			glog.Error("Failed to registe account, the error is username has already existed")
			return
		}
		if u.Path == user.Path {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "Failed to registe account, the error is this path has already been used",
			})
			glog.Error("Failed to registe account, the error is this path has already been used")
			return
		}
		if u.Role == "admin" && user.Role == "admin" {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "Failed to registe account, the error is that one cluster can have only one admin",
			})
			glog.Error("Failed to registe account, the error is that one cluster can have only one admin")
			return
		}
	}

	if user.Role != "admin" {
		if user.Path == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "Failed to registe account, the error is user's path is nil!",
			})
			glog.Error("Failed to registe account, the error is user's path is nil!")
			return
		}
	}

	// create user's path
	err = tools.CreatePath(user.Path, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to create user's path!, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to create user's path!, the error is %v", err.Error())
		return
	}

	log_path := user.Path + "/log"
	err = tools.CreatePath(log_path, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to create log_path!, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to create log_path!, the error is %v", err.Error())
		return
	}

	data_path := user.Path + "/data"
	err = tools.CreatePath(data_path, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to create data_path!, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to create data_path!, the error is %v", err.Error())
		return
	}

	code_path := user.Path + "/code"
	err = tools.CreatePath(code_path, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to create code_path!, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to create code_path!, the error is %v", err.Error())
		return
	}

	glog.Infof("Succeed to create user's work path %v", user.Path)

	// add new user info to file
	users = append(users, user)
	err = tools.WriteUsers(users)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to write users file, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to write users file, the error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to registe user, user.name = %v", user.Username),
	})
	glog.Infof("Succeed to add new user info to local user file, user name is %v, path is %v, role is %v", user.Username, user.Path, user.Role)
}
