package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

const srcfilepath = "/home/gpu-server/set_k8s/test_bishe/dockerfile"

type ImageData struct {
	Dstpath       string   `json:"dstpath"`
	Osversion     string   `json:"osversion"`
	Pythonversion string   `json:"pythonversion"`
	Imagearray    []string `json:"Imagearray"`
	Imagename     string   `json:"Imagename"`
}

// 镜像制作
func CreateImage(c *gin.Context) {
	var image_data ImageData
	// 解析数据
	err_bind := c.ShouldBindJSON(&image_data)
	if err_bind != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_bind.Error(),
		})
		glog.Error("Failed to parse data from request to struct, the error is %s", err_bind)
		return
	}

	// 创建用户的dockerfile文件
	dstFilepath := image_data.Dstpath

	err_create := CopyFile(srcfilepath, dstFilepath)
	if err_create != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_create.Error(),
		})
		glog.Error("Failed to create dockerfile, the error is %s", err_create)
		return
	}

	// 选择系统
	osVersion := image_data.Osversion
	cmd := "FROM " + osVersion + "\n"
	err_version := WriteAtBeginning(dstFilepath, []byte(cmd))
	if err_version != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_version.Error(),
		})
		glog.Error("Failed to write osVersion to dockerfile, the error is %s", err_version)
		return
	}

	// 选择python
	pyVersion := image_data.Pythonversion
	err_py := WriteAtTail(dstFilepath, pyVersion)
	if err_py != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_py.Error(),
		})
		glog.Error("Failed to write pyVersion to dockerfile, the error is %s", err_py)
		return
	}

	// 选择镜像，拉取后将镜像写入dockerfile
	imageArray := image_data.Imagearray
	// dockerClient, _ := InitDocker()
	// defer dockerClient.Close()
	for i := range imageArray {
		// reader, err := dockerClient.ImagePull(context.Background(), imageArray[i], types.ImagePullOptions{})
		// if err != nil {
		// 	c.JSON(http.StatusMovedPermanently, gin.H{
		// 		"code: ":    1,
		// 		"image: ":   imageArray[i],
		// 		"message: ": err.Error(),
		// 	})
		// 	glog.Error("Failed to pull docker image %s, the error is %s", imageArray[i], err)
		// 	return
		// }

		// 写入dockerfile
		err_image := WriteAtTail(dstFilepath, imageArray[i])
		fmt.Println("1")
		if err_image != nil {
			c.JSON(http.StatusMovedPermanently, gin.H{
				"code: ":    1,
				"message: ": err_image.Error(),
			})
			glog.Error("Failed to write image to dockerfile, the error is %s", err_image)
			return
		}
		// glog.Info("Succeed to pull docker image %s", imageArray[i])
		// defer reader.Close()
	}

	// 调用exec执行dockerfile，创建用户自定义镜像
	imageName := image_data.Imagename
	cmd = "docker"
	err_exec := ExecCommand(cmd, "build", "-t", imageName, "-f", dstFilepath, ".")
	if err_exec != nil {
		c.JSON(http.StatusMovedPermanently, gin.H{
			"code: ":    1,
			"message: ": err_exec.Error(),
		})
		glog.Error("Failed to exec docker build, the error is %s", err_exec)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Succeed to build image: %s", imageName),
	})
	return
}

// 初始化Docker客户端
func InitDocker() (*client.Client, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		glog.Error("Failed to init docker client, the error is %s", err)
		return nil, err
	}

	_, err = dockerClient.Ping(context.Background())
	if err != nil {
		glog.Error("Failed to connect to docker client, the error is %s", err)
		return nil, err
	}

	glog.Info("Succeed to init docker client.")
	return dockerClient, nil
}

// 复制dockerfile，不破坏基础dockerfile
func CopyFile(filepath string, newFilepath string) error {
	// 打开原始文件
	src, err_src := os.Open(filepath)
	if err_src != nil {
		glog.Error("Failed to open original dockerfile, the error is ", err_src)
		return err_src
	}
	defer src.Close()

	// 创建目标文件
	dst, err_dst := os.Create(newFilepath)
	if err_dst != nil {
		glog.Error("Failed to create target dockerfile, the error is ", err_dst)
		return err_dst
	}
	defer dst.Close()

	// 复制文件内容
	_, err_copy := io.Copy(dst, src)
	if err_copy != nil {
		glog.Error("Failed to copy file from src to target, the error is ", err_copy)
		return err_copy
	}

	return nil
}

// 从文件头追加写入数据
func WriteAtBeginning(filename string, data []byte) error {
	// 读取文件的原始数据
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	oldData, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// 将新数据插入到旧数据之前
	newData := append(data, oldData...)

	// 将新数据写入文件
	err = ioutil.WriteFile(filename, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// 从文件尾追加写入数据
func WriteAtTail(filepath string, image string) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Error("Failed to open original dockerfile, the error is %s", err)
		return err
	}
	defer file.Close()

	fmt.Println("RUN pip install " + image + "\n")
	s := "\nRUN pip install " + image
	_, err = file.WriteString(s)
	if err != nil {
		glog.Error("Failed to write image, the error is %s", err)
		return err
	}

	return nil
}

// 执行系统命令
func ExecCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	fmt.Println("Result: ", out.String())

	if err != nil {
		glog.Error("Failed to build new images, the error is ", err.Error()+" "+stderr.String())
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
	router.POST("/createimage", CreateImage)

	router.Run(":8080")
}
