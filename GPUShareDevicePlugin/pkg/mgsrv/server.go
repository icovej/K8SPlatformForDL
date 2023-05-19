package mgsrv

import (
	"context"
	"errors"
	"fmt"
	devicevpb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/gpumgr"
	devicevapi "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/mgsrv/deviceapi/device/v1"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
)

type VisibilityLevel string

type MetaGpuServer struct {
	gpuMgr                        *gpumgr.GpuMgr
	ContainerLevelVisibilityToken string
	DeviceLevelVisibilityToken    string
}

var (
	DeviceVisibility         VisibilityLevel = "l0"
	ContainerVisibility      VisibilityLevel = "l1"
	TokenVisibilityClaimName                 = "visibilityLevel"
)

func NewMetaGpuServer() *MetaGpuServer {
	return &MetaGpuServer{gpuMgr: gpumgr.NewGpuManager()}
}

func (s *MetaGpuServer) Start() {

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s", viper.GetString("serverAddr")))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Infof("metagpu gRPC management server listening on %s", viper.GetString("serverAddr"))

		opts := []grpc.ServerOption{
			grpc.UnaryInterceptor(s.unaryServerInterceptor()),
			grpc.StreamInterceptor(s.streamServerInterceptor()),
		}

		grpcServer := grpc.NewServer(opts...)

		dp := devicevapi.DeviceService{}
		devicevpb.RegisterDeviceServiceServer(grpcServer, &dp)
		reflection.Register(grpcServer)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *MetaGpuServer) GenerateAuthTokens(visibility VisibilityLevel) string {

	claims := jwt.MapClaims{"email": "metagpu@instance", TokenVisibilityClaimName: visibility}
	containerScopeToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := containerScopeToken.SignedString([]byte(viper.GetString("jwtSecret")))
	if err != nil {
		log.Error(err)
	}
	return tokenString
}

func (s *MetaGpuServer) IsMethodPublic(fullMethod string) bool {
	publicMethods := []string{
		"/device.v1.DeviceService/PingServer",
	}
	for _, method := range publicMethods {
		if method == fullMethod {
			return true
		}
	}
	return false

}

func authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "retrieving metadata is failed")
	}

	authHeader, ok := md["authorization"]

	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not supplied")
	}

	tokenString := authHeader[0]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, status.Errorf(codes.Unauthenticated, errors.New("error authenticate").Error())
		}
		return []byte(viper.GetString("jwtSecret")), nil
	})
	if err != nil {
		log.Error(err)
		return "", status.Errorf(codes.Unauthenticated, errors.New("error authenticate").Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if visibility, ok := claims[TokenVisibilityClaimName]; ok {
			visibility := visibility.(string)
			return visibility, nil
		}
	}
	return "", status.Errorf(codes.Unauthenticated, errors.New("error authenticate").Error())

}
