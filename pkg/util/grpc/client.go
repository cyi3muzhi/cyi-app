package grpc

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	gg "google.golang.org/grpc"
)

// 创建 gRPC 客户端连接
func NewClientConn(opts []grpc.ClientOption) (*gg.ClientConn, error) {

	conn, err := grpc.DialInsecure(
		context.Background(),
		opts...,
	)
	if err != nil {
		fmt.Printf("Failed to connect to %v\n", err)
		return nil, err
	}
	return conn, nil
}
