package controller

import (
	"net/http"
	"platform_back_end/tools"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	users, err := tools.LoadUsers("")
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":  http.StatusMethodNotAllowed,
			"error: ": err.Error(),
		})
		glog.Error("Failed to load saved users info")
		return
	}

	for i := range users {
		if username == users[i].Username && password == users[i].Password {
			tools.GenerateToken(c, users[i])
			break
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code: ":  http.StatusUnauthorized,
				"message": "Invalid credentials",
			})
		}
	}
}
