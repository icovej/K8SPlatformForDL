package controller

import (
	"fmt"
	"net/http"
	"platform_back_end/data"
	"platform_back_end/tools"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Get all dir infor which user request
func GetDirInfo(c *gin.Context) {
	var Dir data.DirData
	err_bind := c.ShouldBindJSON(&Dir)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

	dir := Dir.Dir
	depth := Dir.Depth
	output, err_exec := tools.ExecCommand("du -h --max-depth=", depth, dir)
	if err_exec != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":    http.StatusMethodNotAllowed,
			"message: ": err_exec.Error(),
		})
		glog.Error("Failed to get %v info, the error is %v", dir, err_exec)
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
