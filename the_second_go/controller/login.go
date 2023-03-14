package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var user User
	body, _ := ioutil.ReadAll(c.Request.Body)
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		glog.Error("Invalid request payload")
		return
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		glog.Error("Failed to unmaeshal request body info!")
		return
	}

	users, err := loadUsers("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		glog.Error("Failed to load saved users info")
		return
	}

	password := user.Password
	name := user.Username

	for i := range users {
		if name == users[i].Username && password == users[i].Password {
			c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
	}
}

func loadUsers(filename string) ([]User, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var users []User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
