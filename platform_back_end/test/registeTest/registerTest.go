package main

import (
	"flag"

	"github.com/IcoveJ/biyesheji/the_second_go/controller"
	"github.com/IcoveJ/biyesheji/the_second_go/tools"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush()

	router := gin.Default()
	router.Use(tools.Core())

	router.POST("/register", controller.RegisterHandler)
}
