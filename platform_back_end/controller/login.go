package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"platform_back_end/tools"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	users, err := loadUsers("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		glog.Error("Failed to load saved users info")
		return
	}

	for i := range users {
		if username == users[i].Username && password == users[i].Password {
			tools.GenerateToken(c, users[i])
			break
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
	}
}

func loadUsers(filename string) ([]tools.User, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var users []tools.User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
