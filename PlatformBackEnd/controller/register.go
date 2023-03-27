package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func RegisterHandler(c *gin.Context) {
	var user data.User
	err_bind := c.ShouldBindJSON(&user)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method RegisterHandler gets invalid request payload")
		return
	}

	// Check if the account has existed
	users, err_check := tools.CheckUsers()
	if err_check != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_check.Error(),
		})
		glog.Error("Failed to check user info, the error is %v", err_check)
		return
	}
	for _, u := range users {
		if u.Username == user.Username {
			c.JSON(http.StatusForbidden, gin.H{
				"code: ":    http.StatusForbidden,
				"message: ": "Username already exists",
			})
			glog.Error("Username already exists!")
			return
		}
		if u.Path == user.Path {
			c.JSON(http.StatusForbidden, gin.H{
				"code: ":    http.StatusForbidden,
				"message: ": "This path has already been used",
			})
			glog.Error("This path has already been used")
			return
		}
		if u.Role == "admin" && user.Role == "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code: ":    http.StatusForbidden,
				"message: ": "One cluster can have only one admin",
			})
			glog.Error("One cluster can have only one admin")
			return
		}
	}

	if user.Role != "admin" {
		if user.Path == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"code: ":    http.StatusForbidden,
				"message: ": "User's path is nil!",
			})
			glog.Error("User's path is nil!")
			return
		}
	}

	// add new user info to file
	users = append(users, user)
	err_add := tools.WriteUsers(users)
	if err_add != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": "Failed to write users file!",
		})
		glog.Error("Failed to write users file!")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code: ":    http.StatusOK,
		"message: ": "Succeed to registe",
	})
}
