package server

import (
	v1 "cyi-app/api/service/chat/v1"
	"cyi-app/app/service/chat/internal/service"

	ts "cyi-app/pkg/kratos/server/transporter"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewServer 注册 gRPC 服务
func NewServer(s *ts.GrpcServer, chatService *service.ChatService) *grpc.Server {
	v1.RegisterChatServer(s.GrpcServer, chatService)
	return s.GrpcServer
}
