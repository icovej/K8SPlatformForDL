package tools

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

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

// 初始化Kubernetes客户端
func InitK8S() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用kubeconfig中的当前上下文,加载配置文件
	config, err_config := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err_config != nil {
		glog.Error("Failed to use load kubeconfig, the error is %s ", err_config)
		return nil, err_config
	}

	// 创建clientset
	clientset, err_client := kubernetes.NewForConfig(config)
	if err_client != nil {
		glog.Error("Failed to create clientset, the error is %s ", err_client)
		return nil, err_client
	}

	return clientset, nil
}

// 获取内存和GPU
func GetAvailableMemoryAndGPU() (uint64, int, map[int]uint64, error) {
	// 获取系统可用内存
	memInfo, _ := mem.VirtualMemory()
	// MB
	memAva := memInfo.Available / 1024 / 1024

	// 获取cpu核数
	cpuCore := runtime.NumCPU()

	// 获取GPU显存
	err_init := nvml.Init()
	if err_init != nil {
		glog.Error("Failed to init nvml to get futher info, the error is %s", err_init)
		return 0, 0, nil, err_init
	}
	defer nvml.Shutdown()

	m := make(map[int]uint64)
	// 获取当前机器上显卡数量，并获取每张显卡的数据
	deviceCount, err_gpu := nvml.GetDeviceCount()
	if err_gpu != nil {
		glog.Error("Failed to get all GPU num, the error is %s", err_gpu)
		return 0, 0, nil, err_gpu
	}
	for i := uint(0); i < deviceCount; i++ {
		device, err_device := nvml.NewDeviceLite(uint(i))
		if err_device != nil {
			glog.Error("Failed to get GPU device, the error is %s", err_device)
			return 0, 0, nil, err_device
		}

		deviceStatus, _ := device.Status()
		// 显卡还剩多少G的内存可用
		avaMem := *deviceStatus.Memory.Global.Free / 1000
		m[int(i)] = avaMem

		glog.Info("GPU %s, the avaMem is %s", i, avaMem)
	}

	return memAva, cpuCore, m, nil
}

// 复制dockerfile，不破坏基础dockerfile
func CopyFile(filepath string, newFilepath string) error {
	// 打开原始文件
	src, err_src := os.Open(filepath)
	if err_src != nil {
		glog.Error("Failed to open original dockerfile, the error is %s", err_src)
		return err_src
	}
	defer src.Close()

	// 创建目标文件
	dst, err_dst := os.Create(newFilepath)
	if err_dst != nil {
		glog.Error("Failed to create target dockerfile, the error is %s", err_dst)
		return err_dst
	}
	defer dst.Close()

	// 复制文件内容
	_, err_copy := io.Copy(dst, src)
	if err_copy != nil {
		glog.Error("Failed to copy file from src to target, the error is %s", err_copy)
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

	writer := bufio.NewWriter(file)
	s := "RUN pip install " + image + "\n"
	writer.WriteString(s)

	return nil
}

// 执行系统命令
func ExecCommand(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	output, err := cmd.Output()
	if err != nil {
		glog.Error("Failed to exec, the error is ", err.Error())
		return nil, err
	}
	return output, nil
}

// 创建目录
func CreatePath(dirpath string, perm os.FileMode) error {
	_, err_stat := os.Stat(dirpath)
	if err_stat == nil {
		glog.Error("This path has exsited, please change!")
		return err_stat
	}

	err_mk := os.MkdirAll(dirpath, perm)
	if err_mk != nil {
		glog.Error("Failed to create user path %s, the error is %s", dirpath, err_mk)
		return err_mk
	}
	return nil
}

// float32 to string
func FloatToString(input_num float32) string {
	// to convert a float number to a string
	return strconv.FormatFloat(float64(input_num), 'f', 6, 64)
}

// 从字符串中提取最后的数字
func extractNumber(s string) int {
	parts := strings.Split(s, "_")
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0
	}
	return n
}

// read data and cacculate their average
func CalculateAvg(filepath string) error {
	// 打开输入文件
	f, err := os.Open(filepath)
	if err != nil {
		glog.Error("Failed to open file, the error is ", err.Error())
		return err
	}
	defer f.Close()

	numValue := make(map[string][]float64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		epoch := fields[0]
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			panic(err)
		}

		numValue[epoch] = append(numValue[epoch], value)
	}

	averages := make(map[string]float64)
	for e, v := range numValue {
		var total float64
		for _, price := range v {
			total += price
		}
		averages[e] = total / float64(len(v))
	}

	sortedItems := make([]string, 0, len(averages))
	for e := range averages {
		sortedItems = append(sortedItems, e)
	}
	sort.Slice(sortedItems, func(i, j int) bool {
		return extractNumber(sortedItems[i]) < extractNumber(sortedItems[j])
	})

	// 打开输出文件
	outputFile, err := os.Create(filepath)
	if err != nil {
		glog.Error("Failed to open output file, the error is ", err.Error())
		return err
	}
	defer outputFile.Close()

	for _, epoch := range sortedItems {
		average := averages[epoch]
		fmt.Fprintf(outputFile, "%s %.10f\n", epoch, average)
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
