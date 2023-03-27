package server

import (
	"context"
	"crypto/md5"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resourceName              string = "nvida.com/device_name"
	defaultDeviceNameLocation string = "/etc/device_name"
	dnSocket                  string = "dn.sock"
	KubeletSocket             string = "kubelet.sock"
	DevicePluginPath          string = "/var/lib/kubelet/device-plugins/"
	device_name_prefix        string = "nvidia"
)

type DnServer struct {
	srv         *grpc.Server
	devices     map[string]*pluginapi.Device
	notify      chan bool
	ctx         context.Context
	cancel      context.CancelFunc
	restartFlag bool // 本次是否是重启
}

// NewDnServer 实例化 dnServer
func NewDNServer() *DnServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &DnServer{
		devices:     make(map[string]*pluginapi.Device),
		srv:         grpc.NewServer(grpc.EmptyServerOption{}),
		notify:      make(chan bool),
		ctx:         ctx,
		cancel:      cancel,
		restartFlag: false,
	}
}

func (dn *DnServer) Run() error {
	// 发现本地设备
	err := dn.listDevice()
	if err != nil {
		glog.Fatal("Failed to list device, the error is %s", err)
		return nil
	}

	go func() {
		err := dn.watchDevice()
		if err != nil {
			glog.Error("Failed to watch devices, the error is %s", err)
		}
	}()

	pluginapi.RegisterDevicePluginServer(dn.srv, dn)
	err = syscall.Unlink(DevicePluginPath + dnSocket)
	if err != nil && !os.IsNotExist(err) {
		glog.Error("Failed to unlink dnsocket, the error is %s", err)
		return err
	}

	l, err := net.Listen("unix", DevicePluginPath+dnSocket)
	if err != nil {
		glog.Error("Failed to listen dnsocket, the error is %s", err)
		return err
	}

	go func() {
		lastCrashTime := time.Now()
		restartCount := 0
		for {
			glog.Info("start GPPC server for %s", resourceName)
			err = dn.srv.Serve(l)
			if err == nil {
				break
			}

			glog.Error("GRPC server for %s crashed with error: %s", resourceName, err)

			if restartCount > 5 {
				glog.Fatal("GRPC server for '%s' has repeatedly crashed recently. Quitting!", resourceName)
			}
			timeSinceLastCrash := time.Since(lastCrashTime).Seconds()
			lastCrashTime = time.Now()
			if timeSinceLastCrash > 3600 {
				restartCount = 1
			} else {
				restartCount++
			}
		}
	}()

	// Wait for server to start by lauching a blocking connection
	conn, err := dn.dial(dnSocket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

func (dn *DnServer) RegisterToKubelet() error {
	socketFile := filepath.Join(DevicePluginPath + KubeletSocket)

	conn, err := dn.dial(socketFile, 5*time.Second)
	if err != nil {
		glog.Error("Failed to connect to kubelet socket, the error is %s", err)
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(DevicePluginPath + dnSocket),
		ResourceName: resourceName,
	}
	glog.Info("Register to kubelet with endpoint %s", req.Endpoint)
	_, err = client.Register(context.Background(), req)
	if err != nil {
		glog.Error("Failed to registe kubelet with endpoint, the error is %s", err)
		return err
	}

	return nil
}

func (dn *DnServer) GetPreferredAllocation(ctx context.Context, reqs *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	glog.Info("GetPreferredAllocation called")
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (dn *DnServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	glog.Info("GetDevicePluginOptions called")
	return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil
}

func (dn *DnServer) ListAndWatch(e *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {
	glog.Info("ListAndWatch called")
	devs := make([]*pluginapi.Device, len(dn.devices))

	i := 0
	for _, dev := range dn.devices {
		devs[i] = dev
		i++
	}

	err := srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		glog.Error("ListAndWatch send device error: %v", err)
		return err
	}

	// 更新 device list
	for {
		glog.Info("Waiting for device change")
		select {
		case <-dn.notify:
			glog.Info("Start to update device list, device number is :", len(dn.devices))
			devs := make([]*pluginapi.Device, len(dn.devices))

			i := 0
			for _, dev := range dn.devices {
				devs[i] = dev
				i++
			}

			srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

		case <-dn.ctx.Done():
			glog.Info("ListAndWatch exit")
			return nil
		}
	}
}

func (dn *DnServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	glog.Info("Allocate called")
	resps := &pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		glog.Info("received request: %v", strings.Join(req.DevicesIDs, ","))
		resp := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"COLA_DEVICES": strings.Join(req.DevicesIDs, ","),
			},
		}
		resps.ContainerResponses = append(resps.ContainerResponses, &resp)
	}
	return resps, nil
}

func (dn *DnServer) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	glog.Info("PreStartContainer called")
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (dn *DnServer) listDevice() error {
	glog.Info("List devices")
	dir, err := ioutil.ReadDir(defaultDeviceNameLocation)
	if err != nil {
		glog.Error("Failed to get device name, the error is %s", err)
		return err
	}
	for _, d := range dir {
		if d.Mode().Type()&os.ModeDevice != 0 && strings.HasPrefix(d.Name(), "nvidia") {
			sum := md5.Sum([]byte(d.Name()))
			dn.devices[d.Name()] = &pluginapi.Device{
				ID:     string(sum[:]),
				Health: pluginapi.Healthy,
			}
			glog.Info("Succeed to find device %s", d.Name())
		}
	}

	return err
}

func (dn *DnServer) watchDevice() error {
	glog.Info("Watching devices")
	w, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Error("NewWatcher error : %s", err)
		return nil
	}
	defer w.Close()

	done := make(chan bool)
	go func() {
		defer func() {
			done <- true
			glog.Info("Watch device exit")
		}()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				glog.Info("device even : %s", event.Op.String())

				if event.Op&fsnotify.Create == fsnotify.Create {
					// 创建文件，增加 device
					sum := md5.Sum([]byte(event.Name))
					dn.devices[event.Name] = &pluginapi.Device{
						ID:     string(sum[:]),
						Health: pluginapi.Healthy,
					}
					dn.notify <- true
					glog.Info("Find new device : %s", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					// 删除文件，删除 device
					delete(dn.devices, event.Name)
					dn.notify <- true
					glog.Info("Deleted device : %s", event.Name)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				glog.Error("There is an error : %s", err)

			case <-dn.ctx.Done():
				break
			}
		}
	}()

	err = w.Add(defaultDeviceNameLocation)
	if err != nil {
		glog.Error("Failed to watch device, the error is : %s", err)
		return err
	}
	<-done

	return nil
}

func (dn *DnServer) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		glog.Error("Failed to grpc dial, the error is %s", err)
		return nil, err
	}

	return c, nil
}
