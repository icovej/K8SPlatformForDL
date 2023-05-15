package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method UploadFile gets invalid request payload")
		return
	}

	// cookie, err := c.Request.Cookie("token")
	// if err != nil {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code":    data.SUCCESS,
	// 		"message": "Failed to get workpath from cookie, Please check it",
	// 	})
	// 	glog.Error("Failed to get workpath from cookie, Please check it")
	// 	return
	// }

	// uploadpath := cookie.Value + "/" + filepath.Base(file.Filename)

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
	token, err := j.Parse_Token(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	uploadpath := token.Path + "/" + filepath.Base(file.Filename)

	err = c.SaveUploadedFile(file, uploadpath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to upload file, the error is %v", err.Error()),
		})
		glog.Error("Failed to get workpath from JWT, Please check it")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": "Succeed to upload file",
	})
}
