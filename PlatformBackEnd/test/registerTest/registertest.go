package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type NormalUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Path     string `json:"path"`
}

func RegisterHandler(c *gin.Context) {
	// 获取请求中的json数据
	var user NormalUser
	err_bind := c.ShouldBindJSON(&user)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err_bind.Error()})
		glog.Error("Failed to parse data form request, the error is %s", err_bind)
		return
	}

	// 检查账号是否已经被注册
	users, err_check := readUsers()
	if err_check != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read users file"})
		glog.Error("Failed to read users file!")
		return
	}
	for _, u := range users {
		if u.Username == user.Username {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
			glog.Error("Username already exists!")
			return
		}
		if u.Path == user.Path {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This path has already been used"})
			glog.Error("This path has already been used")
			return
		}
		if u.Role == "admin" && user.Role == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One cluster can have only one admin"})
			glog.Error("One cluster can have only one admin")
			return
		}
	}

	if user.Role != "admin" {
		if user.Path == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "path is nil"})
			glog.Error("User's path is nil!")
			return
		}
	}

	// 将新用户写入文件
	users = append(users, user)
	err_add := writeUsers(users)
	if err_add != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write users file"})
		glog.Error("Failed to write users file!")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "register success"})
}

func readUsers() ([]NormalUser, error) {
	// 从文件中读取用户信息
	data, err := ioutil.ReadFile("/Users/jiangyiming/Desktop/k8s_bishe/platform_back_end/test/registertest/users.json")
	if err != nil {
		return nil, err
	}

	// 解析JSON数据
	var users []NormalUser
	if len(data) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func writeUsers(users []NormalUser) error {
	// 将用户信息转为JSON格式
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	// 将JSON数据写入文件
	err = ioutil.WriteFile("/Users/jiangyiming/Desktop/k8s_bishe/platform_back_end/test/registertest/users.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
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
	// 启动glog
	flag.Parse()
	defer glog.Flush()
	router := gin.Default()
	router.Use(Core())
	router.POST("/register", RegisterHandler)

	router.Run(":8080")
}
