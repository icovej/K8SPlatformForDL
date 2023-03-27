package tools

import (
	"PlatformBackEnd/data"
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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/mem"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"k8s.io/cri-api/pkg/errors"
)

// Init docker client
func initDocker() (*client.Client, error) {
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
func initK8S() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// var p string = "/home/gpu-server/.kube/config"
	// kubeconfig = &p

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

func CreatePod(poddata data.PodData, pod *corev1.Pod) (*v1.Pod, error) {
	clientset, err := initK8S()
	if err != nil {
		glog.Error("Failed to start k8s, the error is %v", err.Error())
		return nil, err
	}
	pod_container, err := clientset.CoreV1().Pods(poddata.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		glog.Error("Failed to create pod, the error is %v", err.Error())
		return nil, err
	}
	return pod_container, nil
}

func GetAllNamespace() ([]string, error) {
	clientset, err := initK8S()
	if err != nil {
		glog.Error("Failed to start k8s, the error is %v", err.Error())
		return nil, err
	}

	var nameSpaces []string
	namespace, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		glog.Error("Failed to list ns, the error is %v", err.Error())
		return nil, err
	}
	for _, ns := range namespace.Items {
		nameSpaces = append(nameSpaces, ns.Name)
	}
	return nameSpaces, nil
}

func GetAllPod(namespace string) ([]map[string]interface{}, error) {
	clientset, err := initK8S()
	if err != nil {
		glog.Error("Failed to start k8s, the error is %v", err.Error())
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		glog.Error("Failed to list ns, the error is %v", err.Error())
	}

	podList := make([]map[string]interface{}, 0, len(pods.Items))
	for _, pod := range pods.Items {
		createdTime := pod.GetCreationTimestamp().Time
		ageInDays := int(time.Since(createdTime).Hours() / 24)
		podInfo := map[string]interface{}{
			"name":      pod.ObjectMeta.Name,
			"ageInDays": ageInDays,
		}
		podList = append(podList, podInfo)
	}
	return podList, nil
}

func ClearExpiredPod(namespace string) error {
	clientset, err := initK8S()
	if err != nil {
		glog.Error("Failed to start k8s, the error is %v", err.Error())
		return err
	}

	go func() {
		for {
			err := wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
				// 获取所有 Pod
				pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
				if err != nil {
					return false, err
				}

				// 遍历 Pod 并删除超过十分钟的
				for _, pod := range pods.Items {
					if pod.ObjectMeta.CreationTimestamp.Add(1 * time.Minute).Before(time.Now()) {
						err := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).Delete(context.Background(), pod.ObjectMeta.Name, metav1.DeleteOptions{})
						if err != nil {
							if errors.IsNotFound(err) {
								continue
							}
							return false, err
						}
					}
				}

				return true, nil
			})
			if err != nil {
				fmt.Println("", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	return nil
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
