package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"platform_back_end/data"
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

// Init docker client
func InitDocker() (*client.Client, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		glog.Error("Failed to init docker client, the error is %v", err)
		return nil, err
	}

	_, err = dockerClient.Ping(context.Background())
	if err != nil {
		glog.Error("Failed to connect to docker client, the error is %v", err)
		return nil, err
	}

	glog.Info("Succeed to init docker client.")
	return dockerClient, nil
}

// init Kubernetes client
func InitK8S() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Use kubeconfig context to load config file
	config, err_config := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err_config != nil {
		glog.Error("Failed to use load kubeconfig, the error is %v", err_config)
		return nil, err_config
	}

	// build clientset
	clientset, err_client := kubernetes.NewForConfig(config)
	if err_client != nil {
		glog.Error("Failed to create clientset, the error is %v", err_client)
		return nil, err_client
	}

	return clientset, nil
}

// Get mem and GPU
func GetAvailableMemoryAndGPU() (uint64, int, map[int]uint64, error) {
	// Get avaliable mem of host machine
	memInfo, _ := mem.VirtualMemory()
	// the unit is bytes
	memAva := memInfo.Available

	// Get CPU cores
	cpuCore := runtime.NumCPU()

	// Get GPU data
	err_init := nvml.Init()
	if err_init != nil {
		glog.Error("Failed to init nvml to get futher info, the error is %v", err_init)
		return 0, 0, nil, err_init
	}
	defer nvml.Shutdown()

	m := make(map[int]uint64)
	// Get the number of graphics card and their data
	deviceCount, err_gpu := nvml.GetDeviceCount()
	if err_gpu != nil {
		glog.Error("Failed to get all GPU num, the error is %v", err_gpu)
		return 0, 0, nil, err_gpu
	}
	for i := uint(0); i < deviceCount; i++ {
		device, err_device := nvml.NewDeviceLite(uint(i))
		if err_device != nil {
			glog.Error("Failed to get GPU device, the error is %v", err_device)
			return 0, 0, nil, err_device
		}

		deviceStatus, _ := device.Status()
		// Get free num, the unit is bytes
		avaMem := *deviceStatus.Memory.Global.Free
		m[int(i)] = avaMem

		glog.Info("GPU %v, the avaMem is %v", i, avaMem)
	}

	return memAva, cpuCore, m, nil
}

// Copy original dockerfile to dstpath
func CopyFile(filepath string, newFilepath string) error {
	src, err_src := os.Open(filepath)
	if err_src != nil {
		glog.Error("Failed to open original dockerfile, the error is %v", err_src)
		return err_src
	}
	defer src.Close()

	dst, err_dst := os.Create(newFilepath)
	if err_dst != nil {
		glog.Error("Failed to create target dockerfile, the error is %v", err_dst)
		return err_dst
	}
	defer dst.Close()

	_, err_copy := io.Copy(dst, src)
	if err_copy != nil {
		glog.Error("Failed to copy file from src to target, the error is %v", err_copy)
		return err_copy
	}

	return nil
}

// Write new words at the head of file
func WriteAtBeginning(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		glog.Error("Failed to open file, the error is %v", err)
		return err
	}
	defer file.Close()

	oldData, err := ioutil.ReadAll(file)
	if err != nil {
		glog.Error("Failed to read file, the error is %v", err)
		return err
	}

	newData := append(data, oldData...)
	err = ioutil.WriteFile(filename, newData, 0644)
	if err != nil {
		glog.Error("Failed to open write file, the error is %v", err)
		return err
	}

	return nil
}

// Write new words at the tail of file
func WriteAtTail(filepath string, image string) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Error("Failed to open original dockerfile, the error is %v", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	s := "RUN pip install " + image + "\n"
	_, err = writer.WriteString(s)
	if err != nil {
		glog.Error("Failed to open write file, the error is %v", err)
		return err
	}

	return nil
}

func ExecCommand(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	output, err := cmd.Output()
	if err != nil {
		glog.Error("Failed to exec, the error is ", err.Error())
		return nil, err
	}
	return output, nil
}

// Create work path
func CreatePath(dirpath string, perm os.FileMode) error {
	_, err_stat := os.Stat(dirpath)
	if err_stat == nil {
		glog.Error("Stat dirpath successfully, please change!")
		return err_stat
	}

	err_mk := os.MkdirAll(dirpath, perm)
	if err_mk != nil {
		glog.Error("Failed to create user path %v, the error is %v", dirpath, err_mk)
		return err_mk
	}
	return nil
}

// float32 to string
func FloatToString(input_num float32) string {
	// to convert a float number to a string
	return strconv.FormatFloat(float64(input_num), 'f', 6, 64)
}

// extract number from string
func extractNumber(s string) int {
	parts := strings.Split(s, "_")
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		glog.Error("Failed to extract number, the error is %v", err.Error())
		return 0
	}
	return n
}

// read data and cacculate their average
func CalculateAvg(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		glog.Error("Failed to open file, the error is %v", err.Error())
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

	outputFile, err := os.Create(filepath)
	if err != nil {
		glog.Error("Failed to open output file, the error is %v", err.Error())
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

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}

func LoadUsers(filename string) ([]data.User, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var users []data.User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func VerifyChecksum(d []byte, crcMasked uint32) bool {
	rot := crcMasked - data.MaskDelta
	unmaskedCrc := ((rot >> 17) | (rot << 15))

	crc := crc32.Checksum(d, data.Crc32c)

	return crc == unmaskedCrc
}

func CheckUsers() ([]data.User, error) {
	datas, err := ioutil.ReadFile("")
	if err != nil {
		glog.Error("Failed to read file, the error is %v", err.Error())
		return nil, err
	}

	var users []data.User
	if len(datas) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(datas, &users)
	if err != nil {
		glog.Error("Failed to unmarshal user data, the error is %v", err.Error())
		return nil, err
	}

	return users, nil
}

func WriteUsers(users []data.User) error {
	data, err := json.Marshal(users)
	if err != nil {
		glog.Error("Failed to marshal user data, the error is %v", err.Error())
		return err
	}

	err = ioutil.WriteFile("", data, 0644)
	if err != nil {
		glog.Error("Failed to write file, the error is %v", err.Error())
		return err
	}

	return nil
}

func GetLastTwoChars(str string) (string, string) {
	length := len(str)
	if length < 2 {
		return "", ""
	}
	lastTwo := str[length-2:]
	others := str[:length-2]
	return lastTwo, others
}

func GiBToBytes(gib float64) int64 {
	return int64(gib * math.Pow(1024, 3))
}

func MiBToBytes(mib float64) int64 {
	return int64(mib * math.Pow(1024, 2))
}
