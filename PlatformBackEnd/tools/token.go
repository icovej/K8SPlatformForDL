package tools

import (
	"PlatformBackEnd/data"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/glog"
)

type JWT data.JWT

// Create a jwt instance
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// Get signKey
func GetSignKey() string {
	return data.SignKey
}

func SetSignKey(key string) string {
	data.SignKey = key
	return data.SignKey
}

// CreateToken create a token
func (j *JWT) CreateToken(claims data.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// JWT middleware
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("token")
		if tokenString == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "The request is nil with no token, unauthorized access",
			})
			glog.Error("The request is nil with no token, unauthorized access")
			return
		}

		j := NewJWT()
		claims, err := j.ParseToken(tokenString)
		if err != nil && err == data.TokenExpired {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "Token has been expired",
			})
			glog.Error("Token has been expired")
			return
		}
		c.Set("claims", claims)
	}
}

func GenerateToken(c *gin.Context, user data.User) {
	j := &JWT{
		SigningKey: []byte("newtoken"),
	}

	claims := data.CustomClaims{
		Username: user.Username,
		Role:     user.Role,
		Path:     user.Path,
		// RegisteredClaims: jwt.RegisteredClaims{
		// 	ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(data.TOKEN_MAX_EXPIRE_HOUR * time.Hour)},
		// },
		StandardClaims: jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000), // teh time sign to work
			ExpiresAt: int64(time.Now().Unix() + 3600), // Expired time, an hour过期时间 一小时
			Issuer:    data.SignKey,                    // Signer
		},
	}

	token, err := j.CreateToken(claims)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to create token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to create token, the error is %v", err.Error())
		return
	}

	token_data := data.LoginResult{
		User:  user,
		Token: token,
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    user.Path,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{
		"code":    data.SUCCESS,
		"message": "Succeed to login",
		"data":    token_data,
	})
}

func GetDataByTime(c *gin.Context) {
	claims := c.MustGet("claims").(*data.CustomClaims)
	if claims != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": "Token works",
			"data":    claims,
		})
	}
}

func (j *JWT) ParseToken(tokenString string) (*data.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &data.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, data.TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, data.TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, data.TokenNotValidYet
			} else {
				return nil, data.TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*data.CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, data.TokenInvalid
}

func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &data.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*data.CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", data.TokenInvalid
}
