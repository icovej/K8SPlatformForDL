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
	fmt.Printf("token = %v", tokenString)
	if tokenString == "" {
		// 没有传递token
		fmt.Println("1111")
		return
	}
	token, err := j.Parse_Token(tokenString)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// claims, _ := token.Claims.(*data.CustomClaims)

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
