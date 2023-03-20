package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"new_device_plugin/server"

	"github.com/golang/glog"
	"gopkg.in/fsnotify.v1"
)

func main() {
	flag.Parse()
	defer glog.Flush()
	glog.Info("device plugin to get device name starting")
	dnSrv := server.NewDNServer()
	go dnSrv.Run()

	// 向 kubelet 注册
	if err := dnSrv.RegisterToKubelet(); err != nil {
		glog.Error("Failed to registe to kubelet, the error is : %s", err)
	} else {
		glog.Info("Succeed to registe to kubelet")
	}

	// 监听 kubelet.sock，一旦创建则重启
	devicePluginSocket := filepath.Join(server.DevicePluginPath, server.KubeletSocket)
	glog.Info("Get device plugin socket name : %s", devicePluginSocket)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Error("Failed to created FS watcher, the error is %s", err)
		os.Exit(1)
	}
	defer watcher.Close()
	err = watcher.Add(server.DevicePluginPath)
	if err != nil {
		glog.Error("Failed to watch kubelet, the error is %s", err)
		return
	}
	glog.Info("Succeed to watch kubelet.sock")
	for {
		select {
		case event := <-watcher.Events:
			glog.Infof("Watch kubelet events: %s, event name: %s, isCreate: %v", event.Op.String(), event.Name, event.Op&fsnotify.Create == fsnotify.Create)
			if event.Name == devicePluginSocket && event.Op&fsnotify.Create == fsnotify.Create {
				time.Sleep(time.Second)
				glog.Info("Created inotify: %s, restarting.", devicePluginSocket)
			}
		case err := <-watcher.Errors:
			glog.Info("Inotify: %s", err)
		}
	}
}
