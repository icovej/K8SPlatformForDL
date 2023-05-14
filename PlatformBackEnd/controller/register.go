package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func RegisterHandler(c *gin.Context) {
	var user data.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method RegisterHandler gets invalid request payload")
		return
	}

	// Check if the account has existed
	users, err := tools.CheckUsers()
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to check user info, the error is %v", err)
		return
	}
	for _, u := range users {
		if u.Username == user.Username {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Username already exists",
			})
			glog.Error("Username already exists!")
			return
		}
		if u.Path == user.Path {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "This path has already been used",
			})
			glog.Error("This path has already been used")
			return
		}
		if u.Role == "admin" && user.Role == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "One cluster can have only one admin",
			})
			glog.Error("One cluster can have only one admin")
			return
		}
	}

	if user.Role != "admin" {
		if user.Path == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "User's path is nil!",
			})
			glog.Error("User's path is nil!")
			return
		}
	}

	// add new user info to file
	users = append(users, user)
	err = tools.WriteUsers(users)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": "Failed to write users file!",
		})
		glog.Error("Failed to write users file!")
		return
	}

	// create user's path
	err = os.MkdirAll(user.Path, 0777)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": "Failed to create user's path!",
		})
		glog.Errorf("Failed to create user's path!, the error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": "Succeed to registe",
	})
}
