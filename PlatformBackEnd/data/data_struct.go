package data

import (
	"bytes"
	"errors"
	"hash/crc32"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	v1 "k8s.io/api/core/v1"
)

const (
	// user data file
	UserFile      = "User.json"
	PodFile       = "Pod.json"
	NamespaceFile = "Ns.json"

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

	Socketconfig = &engineio.Options{
		PingTimeout:  7 * time.Second,
		PingInterval: 5 * time.Second,
		Transports: []transport.Transport{
			&polling.Transport{
				Client: &http.Client{
					Timeout: time.Minute,
				},
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		},
	}
)

// Api information structure
type Operation struct {
	Api    string `json:"api"`
	Params string `json:"params"`
	Remark string `json:"remark"`
}

// Image data from the front-end
type ImageData struct {
	Dockerfile string `json:"dockerfile"`
	Dstpath    string `json:"dstpath"`
	Imagename  string `json:"Imagename"`
}

// Dir data from the front-end
type DirData struct {
	Dir   string `json:"dir"`
	Depth string `json:"max-depth"`
}

// Pod data from the front-end
type PodData struct {
	Podname   string  `json:"podname"`
	Container string  `json:"container"`
	Memory    string  `json:"memory"`
	Cpu       string  `json:"cpu"`
	Gpu       string  `json:"gpu"`
	Memlim    string  `json:"memlim"`
	Cpulim    string  `json:"cpulim"`
	Gpulim    string  `json:"gpulim"`
	Imagename string  `json:"imagename"`
	Namespace string  `json:"namespace"`
	CPort     []int32 `json:"cport"`
	HPort     []int32 `json:"hport"`
}

type PodInfo struct {
	Name      string
	AgeInDays int
	Status    v1.PodPhase
}

type NsData struct {
	Namespace string `json:"namespace"`
	Days      int    `json:"days"`
}

type PodTimeDelete struct {
	NsData
	Time int
}

// Used to read event file
type Reader struct {
	R   io.Reader
	Buf *bytes.Buffer
}

// Events data structure
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
	Username  string
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

type PodUser struct {
	PodName  string
	UserName string
}
type ClusterNodeData struct {
	NodeName     string
	NodeCPUAll   float64
	NodeCPUUse   float64
	NodeMemAllGB float64
	NodeMemUseGB float64
	NodeMemAllMB float64
	NodeMemUseMB float64
}
