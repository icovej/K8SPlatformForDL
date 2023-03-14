package tools

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	// 加载kubeconfog
	kubeConfig, err_config := rest.InClusterConfig()
	if err_config != nil {
		glog.Error("Failed to get kubeconfig, the error is %s", err_config)
		return nil, err_config
	}

	// 创建k8s客户端
	kubeClient, err_client := kubernetes.NewForConfig(kubeConfig)
	if err_client != nil {
		glog.Error("Failed to init kubeClient, the error is %s", err_client)
		return nil, err_client
	}

	glog.Info("Succeed to init kubeClient")
	return kubeClient, nil
}

// 获取内存和GPU
func GetAvailableMemoryAndGPU() (uint64, int, map[int]uint64, error) {
	// 获取系统可用内存
	memInfo, _ := mem.VirtualMemory()
	memAva := memInfo.Available / 1024 / 1024 / 1024

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
		usedMem := deviceStatus.Utilization.Memory
		avaMem := *device.Memory - uint64(*usedMem)
		m[0] = avaMem

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
func WriteAtHead(filepath string, content string) error {
	// 读取文件内容
	data, err_read := ioutil.ReadFile(filepath)
	if err_read != nil {
		glog.Error("Failed to read dockerfile: %s", filepath)
		return err_read
	}

	// 打开文件并准备写入数据
	file, err_open := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err_open != nil {
		glog.Error("Failed to open dockerfile: %s", filepath)
		return err_open
	}
	defer file.Close()

	// 将新数据追加到原始数据之前
	_, err_write := file.WriteAt([]byte(content), 0)
	if err_write != nil {
		glog.Error("Failed to write content: %s", content)
		return err_write
	}
	_, err := file.WriteAt(data, int64(len(filepath)))
	if err != nil {
		glog.Error("Failed to write data: %s", data)
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
		glog.Error("Failed to build new images, the error is %s", err)
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
