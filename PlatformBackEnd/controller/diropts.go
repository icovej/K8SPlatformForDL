package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Get all dir infor which user request
func GetDirInfo(c *gin.Context) {
	var Dir data.DirData
	err := c.ShouldBindJSON(&Dir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method GetDirInfo gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

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

	path := token.Path + "/" + Dir.Dir

	output, err := tools.ExecCommand("du", "-h", path, "--max-depth", Dir.Depth)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to get %v info, the error is %v", path, err),
		})
		glog.Errorf("Failed to get %v info, the error is %v", path, err)
		return
	}

	lines := strings.Split(string(output), "\n")
	result := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		result[fields[1]] = fields[0]
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to get dir %v info", path),
		"data":    result,
	})
}

func CreateDir(c *gin.Context) {
	var Dir data.DirData
	err := c.ShouldBindJSON(&Dir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method CreateDir gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method CreateDir gets invalid request payload")
		return
	}

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

	err = tools.CreatePath(token.Path+"/"+Dir.Dir, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to create path %v, the error is %v", Dir.Dir, err.Error()),
		})
		glog.Errorf("Failed to create path %v, the error is %v", Dir.Dir, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to create dir %v", Dir.Dir),
	})
}

func DeleteDir(c *gin.Context) {
	var Dir data.DirData
	err := c.ShouldBindJSON(&Dir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method DeleteDir gets invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method DeleteDir gets invalid request payload")
		return
	}

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

	path := token.Path + "/" + Dir.Dir

	_, err = tools.ExecCommand("rm", "-rf", path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to delete dir %v, the error is %v", Dir.Dir, err),
		})
		glog.Errorf("Failed to delete dir %v, the error is %v", Dir.Dir, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": fmt.Sprintf("Succeed to delete dir %v", Dir.Dir),
	})
}
