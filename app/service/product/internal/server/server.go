package server

import (
	"context"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/conf"
	productNats "cyi-app/app/service/product/internal/nats"
	consul "cyi-app/pkg/kratos/config"
	"os"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/google/wire"
	"github.com/nats-io/nats.go"

	ts "cyi-app/pkg/kratos/server/transporter"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServer, NewRegistrar, NewGRPCServerConf, NewNATSConn, NewNATSServer)

func NewGRPCServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	metric *mid.Meter,
	ac *conf.Auth) *ts.GrpcServer {
	srv := grpc.NewServer(ts.NewGrpcServerOpt(
		&ts.TransporterOpt{
			Network: c.Grpc.Network,
			Addr:    c.Grpc.Addr,
			Timeout: c.Grpc.Timeout,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			Logger:         logger,
			ApiKey:         ac.ApiKey,
			Meter:          metric,
		})...)
	return &ts.GrpcServer{GrpcServer: srv}
}

// 注册服务器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return reg
}

// NewNATSConn 创建NATS连接
// 优先从环境变量 NATS_URL 读取地址，默认使用 localhost（本地开发）
func NewNATSConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	return conn
}

// NATSServer NATS订阅服务（作为kratos server管理生命周期）
type NATSServer struct {
	handler *productNats.StockHandler
	conn    *nats.Conn
}

// NewNATSServer 创建NATS订阅服务
func NewNATSServer(handler *productNats.StockHandler, conn *nats.Conn) *NATSServer {
	return &NATSServer{handler: handler, conn: conn}
}

// Start 启动NATS订阅
func (s *NATSServer) Start(ctx context.Context) error {
	if err := s.handler.Register(s.conn); err != nil {
		return err
	}
	return nil
}

// Stop 停止NATS订阅
func (s *NATSServer) Stop(ctx context.Context) error {
	s.conn.Close()
	return nil
}

// NewNATSHandler 创建NATS库存处理器（给wire注入用）
func NewNATSHandler(uc *biz.StockUseCase, logger log.Logger) *productNats.StockHandler {
	return productNats.NewStockHandler(uc, logger)
}
