package controller

import (
	"PlatformBackEnd/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Operation data.Operation

// 操作接口信息
func OperationInfo(c *gin.Context) {
	list := []Operation{
		{
			"/regoster",
			"username, password, permission, workpath",
			"注册账号",
		},
		{
			"/login",
			"username, password",
			"登录账号",
		},
		{
			"/search_dir",
			"dir_name, depth",
			"查询目录存储",
		},
		{
			"/image",
			"dstpath, osversion, pyversion, imagearray, imagename",
			"创建镜像",
		},
		{
			"/pod",
			"podname, container_name, memeory, cpu, etc",
			"创建容器",
		},
		{
			"/data",
			"logdir",
			"模型训练的损失值和正确率",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"list": list,
	})
}
