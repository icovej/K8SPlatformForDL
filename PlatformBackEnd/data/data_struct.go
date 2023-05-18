package data

import (
	"bytes"
	"errors"
	"hash/crc32"
	"io"

	"github.com/golang-jwt/jwt/v4"
)

const (
	// user data file
	UserFile = "User.json"

	// the metric of nvidia/com
	// the value is set by Kubernetes
	GpuMetricName = "cnvrg.io/metagpu"

	// the mask of parsing event file
	MaskDelta = 0xa282ead8

	// header and tail padding of the event file data
	HeaderSize = 12
	FooterSize = 4

	// the prefix of the name of each epoch
	TestPrefix  = "test_loss"
	TrainPrefix = "train_loss"
	Accuracy    = "test_accuracy"

	// three data file to save data parsed from event file
	TestLossFile  = "test_loss_all.txt"
	TrainLossFile = "train_loss_all.txt"
	AccFile       = "acc.txt"

	// about token expire
	TOKEN_MAX_EXPIRE_HOUR      = 1 * 24 * 7
	TOKEN_MAX_REMAINING_MINUTE = 15
)

var (
	TokenExpired     error  = errors.New("Token is expired")
	TokenExpiring    error  = errors.New("Token will be expired in one minute")
	TokenNotValidYet error  = errors.New("Token not active yet")
	TokenMalformed   error  = errors.New("That's not even a token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = "newtoken"

	// when the crc of the heand or tail of event file is invalid, the return
	ErrInvalidChecksum = errors.New("invalid crc")
	Crc32c             = crc32.MakeTable(crc32.Castagnoli)
	// the basic path of dockerfile
	Srcfilepath = "/home/gpu-server/all_test/biyesheji/PlatformBackEnd/dockerfile"
)

// Api information structure
type Operation struct {
	Api    string `json:"api"`
	Params string `json:"params"`
	Remark string `json:"remark"`
}

// Image data from the front-end
//
// Dstpath: the path to save dockerfile of user
// Osversion: for example, ubuntu:20.04
// Pythonversion: for example, python3.8-slim-buster
// Imagearray: user can select them at the front-end
// Imagename: the image name of which is built from user by using the dockerfile
type ImageData struct {
	Dstpath       string   `json:"dstpath"`
	Osversion     string   `json:"osversion"`
	Pythonversion string   `json:"pythonversion"`
	Imagearray    []string `json:"Imagearray"`
	Imagename     string   `json:"Imagename"`
}

// Dir data from the front-end
//
// Dir: the path from user to find a path that is empty or enough to use
// Depth: the arg from user to combine the command "du"
type DirData struct {
	Dir   string `json:"dir"`
	Depth string `json:"max-depth"`
}

// Pod data from the front-end
// Nowadays(2023.3.23) it can only support to build one pod with only one container
// In future, we'll update it to support multiple containers
//
// Podname: selected from user
// Container: selected from user
// Memory: this value is decided by your host machine, even though we'll check if it's valid,
// we still hope you can make sure the value you choose is less that the avaliable of your
// machine before creating pod. The unit is the same as Kubernetes.
// Cpu: this value's unit is core. For example, you choose CPU=4, means 4 CPU cores will be
// used in your work.
// Gpu: this value means which graphics card you want to use. For example, Gpu=0, means you will
// use /dev/nvidia0. We are currently working on how to use a certain graphics card, in fact,
// we have finished writing code, but for some reason, we haven't tested it yet. In future,
// we'll make it. TODO:
// XXXlim: there are three values to limit mem, GPU and CPU. In fact, Kubernetes's not forcing us
// to fill out them, but we think the most important thing of multi-model training is safety，
// so, we need you to do them. The regulation is the same as Kubernetes
// Mountname: the same as Kubernetes
// Mountpath: the path which admin give you
// Nodename: the node on which you want to create your pod. In future, we'll make it as the path boung
// with your account. TODO:
// Namespace: the namespace in which you want to create your pod. In future, we'll make it as the path boung
// with your account. The reasons why we want to bind it with account is that maybe in future, users want
// to use PV, PVC, SC, etc. It's easier to manage everything in the same namespace.  TODO:
type PodData struct {
	Podname   string `json:"podname"`
	Container string `json:"container_name"`
	Memory    string `json:"memory"`
	Cpu       string `json:"cpu"`
	Gpu       string `json:"gpu"`
	Memlim    string `json:"memlim"`
	Cpulim    string `json:"cpulim"`
	Gpulim    string `json:"gpulim"`
	Imagename string `json:"imagename"`
	Nodename  string `json:"nodename"`
	Namespace string `json:"namespace"`
}

// Used to read event file
type Reader struct {
	R   io.Reader
	Buf *bytes.Buffer
}

// Events data structure
//
// Just like tensorboard, we need logs' path to read data
type ModelLogData struct {
	Logdir string `json:"logdir"`
}

// User information
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Path     string `json:"path"`
}

type IP struct {
	Master string   `json:"masterip"`
	Node   []string `json:"nodeip"`
}

// JWT, when user login this platform, a token will be created and sent to platform
type LoginResult struct {
	Token string `json:"token"`
	User
}

// JWT load
type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Path     string `json:"path"`
	jwt.StandardClaims
}

type JWT struct {
	SigningKey []byte
}

type Monitor struct {
	Namespace string `json:"namespace"`
}

type NodeIP struct {
	Master string   `json:"MasterIP"`
	Node   []string `json:"nodeip"`
}

type DataJSON struct {
	JSON1 interface{} `json:"json1"`
	JSON2 interface{} `json:"json2"`
	JSON3 interface{} `json:"json3"`
}

type FileData struct {
	Dir  []string `json:"dir"`
	File []string `json:"file"`
}

type PodGPUData struct {
	Name      string
	Namespace string
	Device    string
	Node      string
	MemUse    string
	MemAll    string
	Req       string
}

type NodeGPU struct {
	NodeName string
	GPUCount int
}
