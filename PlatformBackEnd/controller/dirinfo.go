package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Get all dir infor which user request
func GetDirInfo(c *gin.Context) {
	var Dir data.DirData
	err := c.ShouldBindJSON(&Dir)
	if err != nil {
		c.JSON(data.API_PARAMETER_ERROR, gin.H{
			"code: ":    data.API_PARAMETER_ERROR,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

	dir := Dir.Dir
	depth := Dir.Depth
	output, err := tools.ExecCommand("du -h --max-depth=", depth, dir)
	if err != nil {
		c.JSON(data.OPERATION_FAILURE, gin.H{
			"code: ":    data.OPERATION_FAILURE,
			"message: ": err.Error(),
		})
		glog.Error("Failed to get %v info, the error is %v", dir, err)
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
	c.JSON(data.SUCCESS, gin.H{
		"code:":    data.SUCCESS,
		"message:": "Succeed to get fir info",
		"data:":    result,
	})
}
