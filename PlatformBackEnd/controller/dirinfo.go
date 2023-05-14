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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Invalid request payload, err is %v", err.Error()),
		})
		glog.Error("Method GetDirInfo gets invalid request payload")
		return
	}

	dir := Dir.Dir
	depth := Dir.Depth
	output, err := tools.ExecCommand("du -h --max-depth=", depth, dir)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Failed to get %v info, the error is %v", dir, err)
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
		"message": "Succeed to get fir info",
		"data":    result,
	})
}
