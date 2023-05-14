package controller

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method UploadFile gets invalid request payload")
		return
	}

	cookie, err := c.Request.Cookie("token")
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": "Failed to get workpath from JWT, Please check it",
		})
		glog.Error("Failed to get workpath from JWT, Please check it")
		return
	}

	uploadpath := cookie.Value + "/" + filepath.Base(file.Filename)

	err = c.SaveUploadedFile(file, uploadpath)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": fmt.Sprintf("Failed to upload file, the error is %v", err.Error()),
		})
		glog.Error("Failed to get workpath from JWT, Please check it")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Succeed to upload",
	})
}
