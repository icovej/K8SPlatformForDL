package function_test

import (
	"PlatformBackEnd/controller"
	"PlatformBackEnd/data"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
)

type User data.User

func TestRegister(t *testing.T) {
	var filepath = "/home/gpu-server/set_k8s/biyesheji/PlatformBackEnd/test/users.json"
	// Clear user info file
	// err := ioutil.WriteFile(filepath, []byte{}, 0644)
	// convey.So(err, convey.ShouldBeNil)

	router := gin.New()

	router.POST("/register", controller.RegisterUser)

	// The first test example to registe a new account with Role=admin
	// Test "Succeed to registe"
	convey.Convey("Create a new account with Role=admin", t, func() {
		newUser := data.User{
			Username: "Mary",
			Password: "123456",
			Role:     "admin",
			Path:     "/home/Mary/xx",
		}
		newUserData, err := json.Marshal(newUser)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(newUserData))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		convey.Convey("Then the HTTP response should be 200 OK and return a success message", func() {
			convey.So(res.Code, convey.ShouldEqual, http.StatusOK)

			response := make(map[string]interface{})
			err := json.Unmarshal(res.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message: "], convey.ShouldEqual, "Succeed to registe")
		})

	})

	// The second test example to registe a new account with Role=admin
	// Test "One cluster can have only one admin"
	convey.Convey("Create a new account with Role=admin", t, func() {
		newUser := data.User{
			Username: "David",
			Password: "123456",
			Role:     "admin",
			Path:     "/home/David/xx",
		}
		newUserData, err := json.Marshal(newUser)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(newUserData))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		convey.Convey("Then the HTTP response should be 403 StatusForbidden and return a message about admin", func() {
			convey.So(res.Code, convey.ShouldEqual, http.StatusForbidden)

			response := make(map[string]interface{})
			err := json.Unmarshal(res.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message: "], convey.ShouldEqual, "One cluster can have only one admin")
		})
	})

	// The third test example to registe a new account with Role=user
	// Test "Succeed to registe" again
	convey.Convey("Create a new account with Role=user", t, func() {
		newUser := data.User{
			Username: "David",
			Password: "123456",
			Role:     "user",
			Path:     "/home/David/xx",
		}
		newUserData, err := json.Marshal(newUser)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(newUserData))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		convey.Convey("Then the HTTP response should be 200 OK and return a success message", func() {
			convey.So(res.Code, convey.ShouldEqual, http.StatusOK)

			response := make(map[string]interface{})
			err := json.Unmarshal(res.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message: "], convey.ShouldEqual, "Succeed to registe")
		})
	})

	// The forth test example to registe a new account with Path=the third
	// Test "This path has already been used"
	convey.Convey("Create a new account with Path=the third one", t, func() {
		newUser := data.User{
			Username: "Cathy",
			Password: "123456",
			Role:     "user",
			Path:     "/home/David/xx",
		}
		newUserData, err := json.Marshal(newUser)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(newUserData))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		convey.Convey("Then the HTTP response should be 403 StatusForbidden and return a message about path", func() {
			convey.So(res.Code, convey.ShouldEqual, http.StatusForbidden)

			response := make(map[string]interface{})
			err := json.Unmarshal(res.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message: "], convey.ShouldEqual, "This path has already been used")
		})
	})

	// The fifth test example to registe a new account with Role=user
	// Test "Succeed to registe" again
	convey.Convey("Create a new account with Role=admin", t, func() {
		newUser := data.User{
			Username: "Tom",
			Password: "123456",
			Role:     "user",
			Path:     "/home/Tom/xx",
		}
		newUserData, err := json.Marshal(newUser)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(newUserData))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		convey.Convey("Then the HTTP response should be 200 OK and return a success message", func() {
			convey.So(res.Code, convey.ShouldEqual, http.StatusOK)

			response := make(map[string]interface{})
			err := json.Unmarshal(res.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)

			convey.So(response["message: "], convey.ShouldEqual, "Succeed to registe")
		})
	})

	// Test the function of write user info
	convey.Convey("Read json to get user number", t, func() {
		data, err := ioutil.ReadFile(filepath)
		convey.So(err, convey.ShouldBeNil)

		var user []User
		err = json.Unmarshal(data, &user)
		convey.So(err, convey.ShouldBeNil)

		count := len(user)
		convey.So(count, convey.ShouldEqual, 3)
		convey.So(user[0].Username, convey.ShouldEqual, "Mary")
		convey.So(user[1].Username, convey.ShouldEqual, "David")
		convey.So(user[2].Username, convey.ShouldEqual, "Tom")
	})
}
