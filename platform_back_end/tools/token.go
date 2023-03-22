package tools

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/glog"
)

const (
	TOKEN_MAX_EXPIRE_HOUR      = 1 * 24 * 7 // token最长有效期
	TOKEN_MAX_REMAINING_MINUTE = 15         // token还有多久过期就返回新token
)

type JWT struct {
	SigningKey []byte
}

var (
	TokenExpired     error  = errors.New("Token is expired")
	TokenExpiring    error  = errors.New("Token will be expired in one minute")
	TokenNotValidYet error  = errors.New("Token not active yet")
	TokenMalformed   error  = errors.New("That's not even a token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = "newkey"
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

// 解析Tokne
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
