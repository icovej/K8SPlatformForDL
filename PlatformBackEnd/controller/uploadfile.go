package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Method UploadFile gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method UploadFile gets invalid request payload, the error is %v", err.Error())
		return
	}

	path := c.PostForm("path")

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

	father_path := token.Path + "/" + path
	_, err = os.Stat(father_path)
	if err != nil {
		err = tools.CreatePath(father_path, 0777)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    data.SUCCESS,
				"message": fmt.Sprintf("Failed to create path %v, the error is %v", father_path, err.Error()),
			})
			glog.Errorf("Failed to create path %v, the error is %v", father_path, err.Error())
			return
		}
	}

	uploadpath := father_path + "/" + filepath.Base(file.Filename)

	err = c.SaveUploadedFile(file, uploadpath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to upload file, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to get workpath from JWT, Please check it. The error is %v", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to upload file, file name is %v", file.Filename),
	})
}
