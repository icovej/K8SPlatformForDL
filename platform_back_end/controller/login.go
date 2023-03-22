package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"platform_back_end/tools"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Path     string `json:"path"`
}

type LoginResult struct {
	Token string `json:"token"`
	User
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	users, err := loadUsers("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		glog.Error("Failed to load saved users info")
		return
	}

	for i := range users {
		if username == users[i].Username && password == users[i].Password {
			// c.JSON(http.StatusOK, gin.H{
			// 	"Role": users[i].Role,
			// 	"Path": users[i].Path,
			// })
			generateToken(c, users[i])
			break
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
	}
}

func loadUsers(filename string) ([]User, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var users []User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func generateToken(c *gin.Context, user User) {
	j := &tools.JWT{
		SigningKey: []byte("newtoken"),
	}

	claims := tools.CustomClaims{
		Username: user.Username,
		Role:     user.Role,
		Path:     user.Path,
		StandardClaims: jwtgo.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000),      // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + 3600*24*7), // 过期时间 一小时
			Issuer:    "newtoken",                           //签名的发行者
		},
	}

	token, err := j.CreateToken(claims)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    err.Error(),
		})
		return
	}

	glog.Info(token)

	data := LoginResult{
		User:  user,
		Token: token,
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "登录成功！",
		"data":   data,
	})
}

// GetDataByTime 一个需要token认证的测试接口
func GetDataByTime(c *gin.Context) {
	claims := c.MustGet("claims").(*tools.CustomClaims)
	if claims != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": 0,
			"msg":    "token有效",
			"data":   claims,
		})
	}
}
