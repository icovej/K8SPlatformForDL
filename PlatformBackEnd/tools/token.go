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

// Parse tokne
// TODO: some signs haven't been used
func (j *JWT) ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &data.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, data.TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, data.TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, data.TokenNotValidYet
			} else {
				return nil, data.TokenInvalid
			}
		}
	}
	return token, nil
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

		glog.Info("Get token: ", tokenString)

		j := NewJWT()
		token, err := j.ParseToken(tokenString)
		if err != nil && err == data.TokenExpired {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": "Token has been expired",
			})
			glog.Error("Token has been expired")
			return
		}
		if claims, ok := token.Claims.(*data.CustomClaims); ok && token.Valid {
			if !claims.VerifyExpiresAt(time.Now(), false) {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    data.OPERATION_FAILURE,
					"message": "Access token expired",
				})
				glog.Error("Access token expired")
			}
			if t := claims.ExpiresAt.Time.Add(-time.Minute * data.TOKEN_MAX_REMAINING_MINUTE); t.Before(time.Now()) {
				claims := data.CustomClaims{
					Username: claims.Username,
					Role:     claims.Role,
					Path:     claims.Path,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(data.TOKEN_MAX_EXPIRE_HOUR * time.Hour)},
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(j.SigningKey)
				c.Header("new-token", tokenString)
			}
			c.Set("claims", claims)
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    data.OPERATION_FAILURE,
				"message": fmt.Sprintf("Failed to parse claims, the error is %v", err.Error()),
			})
			glog.Errorf("Failed to parse claims, the error is %v", err.Error())
			return
		}

		c.Next()
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
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(data.TOKEN_MAX_EXPIRE_HOUR * time.Hour)},
		},
	}

	token, err := j.CreateToken(claims)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": err.Error(),
		})
		glog.Error("Failed to create token, the error is %v", err.Error())
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

func (j *JWT) Parse_Token(tokenString string) (*data.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &data.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, data.TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
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
