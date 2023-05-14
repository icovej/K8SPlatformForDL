package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func K8SBuild(c *gin.Context) {
	var IPs data.IP
	master_ip := IPs.Master
	node_ip := IPs.Node
	nodeIPsStr := strings.Join(node_ip, ",")
	cmdArgs := []string{"-master", master_ip, "-nodes", nodeIPsStr}
	_, err := tools.ExecCommand("../shell/k8s/install_k8s.sh", cmdArgs...)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Errorf("Command execution failed: %v", err)
		return
	}
}
