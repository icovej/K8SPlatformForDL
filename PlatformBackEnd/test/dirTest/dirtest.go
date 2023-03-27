package main

import (
	"flag"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// 执行系统命令
func ExecCommand(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	output, err := cmd.Output()
	if err != nil {
		glog.Error("Failed to build new images, the error is ", err.Error())
		return nil, err
	}
	return output, nil
}

type DirData struct {
	Dir   string `json:"dir"`
	Depth string `json:"max-depth"`
}

func GetDirInfo(c *gin.Context) {
	var Dir DirData
	err_bind := c.ShouldBindJSON(&Dir)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		glog.Error("Invalid request payload")
		return
	}

	dir := Dir.Dir
	depth := Dir.Depth
	output, err_exec := ExecCommand("du -h --max-depth=", depth, dir)
	if err_exec != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err_exec.Error()})
		glog.Error("Failed to get %s info, the error is %s", dir, err_exec)
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
	c.JSON(http.StatusOK, result)
}

func Core() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token")
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "True")
		//放行索引options
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}

func main() {
	flag.Parse()
	defer glog.Flush()

	router := gin.Default()
	router.Use(Core())

	router.POST("/dir", GetDirInfo)

	router.Run(":8080")

}
