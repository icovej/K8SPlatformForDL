package tools

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/mem"
)

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
func CopyFile(filepath string, newFilepath string) {
	// 打开原始文件
	src, err_src := os.Open(filepath)
	if err_src != nil {
		glog.Error("Failed to open original dockerfile, the error is %s", err_src)
		return
	}
	defer src.Close()

	// 创建目标文件
	dst, err_dst := os.Create(newFilepath)
	if err_dst != nil {
		glog.Error("Failed to create target dockerfile, the error is %s", err_dst)
		return
	}
	defer dst.Close()

	// 复制文件内容
	_, err_copy := io.Copy(dst, src)
	if err_copy != nil {
		glog.Error("Failed to copy file from src to target, the error is %s", err_copy)
		return
	}
}

// 从文件头追加写入数据
func WriteAtHead(filepath string, content string) {

	// 读取文件内容
	data, err_read := ioutil.ReadFile(filepath)
	if err_read != nil {
		glog.Error("Failed to read dockerfile: %s", filepath)
		return
	}

	// 打开文件并准备写入数据
	file, err_open := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err_open != nil {
		glog.Error("Failed to open dockerfile: %s", filepath)
		return
	}
	defer file.Close()

	// 将新数据追加到原始数据之前
	_, err_write := file.WriteAt([]byte(content), 0)
	if err_write != nil {
		glog.Error("Failed to write content: %s", content)
		return
	}
	_, err := file.WriteAt(data, int64(len(filepath)))
	if err != nil {
		glog.Error("Failed to write data: %s", data)
		return
	}

}

// 从文件尾追加写入数据
func WriteAtTail(filepath string, imageArray []string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		glog.Error("Failed to open original dockerfile, the error is %s", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for i := range imageArray {
		s := "RUN pip install " + imageArray[i] + "\n"
		writer.WriteString(s)
	}
}

// 执行系统命令
func ExecCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		glog.Error("Failed to build new images, the error is %s", err)
		return err
	}
	return nil
}
