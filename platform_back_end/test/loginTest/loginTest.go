package main

import (
	"flag"
	"platform_back_end/tools"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush()

	router := gin.Default()
	router.Use(tools.Core())

}
