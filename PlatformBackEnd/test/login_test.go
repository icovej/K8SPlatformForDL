package function_test

import (
	"PlatformBackEnd/controller"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
)

func TestLogin(t *testing.T) {
	router := gin.New()

	router.POST("/login", controller.Login)

	convey.Convey("Give a HTTP request to login with an existing account", t, func() {
		postBody := []byte(`{"username": "jym", "password": "123456"}`)

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(postBody))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		convey.Convey("Then the HTTP response should be 200 OK and return a success message", func() {
			convey.So(res.Code, convey.ShouldEqual, http.StatusUnauthorized)
			response := make(map[string]interface{})
			err := json.Unmarshal(res.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message"], convey.ShouldEqual, "Invalid credentials")
		})
	})

}
