package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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

const (
	TOKEN_MAX_EXPIRE_HOUR      = 1 * 24 * 7 // token最长有效期
	TOKEN_MAX_REMAINING_MINUTE = 15         // token还有多久过期就返回新token
)

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	users, err := loadUsers("/Users/jiangyiming/Desktop/k8s_bishe/platform_back_end/test/users.json")
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
	j := &JWT{
		SigningKey: []byte("newtoken"),
	}

	claims := CustomClaims{
		Username: user.Username,
		Role:     user.Role,
		Path:     user.Path,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(TOKEN_MAX_EXPIRE_HOUR * time.Hour)},
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
	claims := c.MustGet("claims").(*CustomClaims)
	if claims != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": 0,
			"msg":    "token有效",
			"data":   claims,
		})
	}
}

// JWT中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("token")
		if tokenString == "" {
			c.JSON(http.StatusOK, gin.H{
				"status": -1,
				"msg":    "请求未携带tken, 无权限访问",
			})
			c.Abort()
			return
		}

		glog.Info("get token: ", tokenString)

		j := NewJWT()
		// parseToken 解析token包含的信息
		token, err := j.ParseToken(tokenString)
		if err != nil && err == TokenExpired {
			c.JSON(http.StatusOK, gin.H{
				"status": -1,
				"msg":    "token has been expired",
			})
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			if !claims.VerifyExpiresAt(time.Now(), false) {
				c.JSON(http.StatusForbidden, gin.H{
					"status": -1,
					"msg":    "access token expired",
				})
			}
			if t := claims.ExpiresAt.Time.Add(-time.Minute * TOKEN_MAX_REMAINING_MINUTE); t.Before(time.Now()) {
				claims := CustomClaims{
					Username: claims.Username,
					Role:     claims.Role,
					Path:     claims.Path,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(TOKEN_MAX_EXPIRE_HOUR * time.Hour)},
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(j.SigningKey)
				c.Header("new-token", tokenString)
			}
			c.Set("claims", claims)
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": -1,
				"msg":  fmt.Sprintf("Claims parse error: %v", err),
			})
			return
		}
		c.Next()
	}
}

type JWT struct {
	SigningKey []byte
}

var (
	TokenExpired     error  = errors.New("Token is expired")
	TokenNotValidYet error  = errors.New("Token not active yet")
	TokenMalformed   error  = errors.New("That's not even a token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = "newtrekWang"
)

// 载荷，可以加一些自己需要的信息
type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Path     string `json:"path"`
	jwt.RegisteredClaims
}

// 新建一个jwt实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// 获取signKey
func GetSignKey() string {
	return SignKey
}

// 这是SignKey
func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

// CreateToken 生成一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析Token
func (j *JWT) ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	return token, nil
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

	router.POST("/login", Login)

	taR := router.Group("/data")
	taR.Use(JWTAuth())

	{
		taR.GET("/dataByTime", GetDataByTime)
	}

	router.Run(":8080")

}
