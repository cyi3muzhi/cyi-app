package grpc

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	gg "google.golang.org/grpc"
)

func ConnClient(serviceName string, r registry.Discovery, opts ...grpc.ClientOption) (*gg.ClientConn, error) {
	// 合并用户传入的选项和基础选项
	allOpts := append([]grpc.ClientOption{
		grpc.WithEndpoint(fmt.Sprintf("discovery:///%v.%v", os.Getenv("SERVICE_ENV"), serviceName)),
		grpc.WithDiscovery(r),
	}, opts...)

	conn, err := grpc.DialInsecure(
		context.Background(),
		allOpts...,
	)
	if err != nil {
		fmt.Printf("Failed to connect to %v: %v\n", serviceName, err)
		return nil, err
	}
	return conn, nil
}
