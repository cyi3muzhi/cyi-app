package server

import (
	"cyi-app/app/cin/interface/internal/conf"
	consul "cyi-app/pkg/util/config"

	ts "cyi-app/pkg/util/server/transporter"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar, NewDiscovery, NewDB, NewGRPCServerConf, NewHTTPServerConf)

func NewHTTPServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	auth *conf.Auth) *ts.HttpServer {
	srv := http.NewServer(ts.NewHttpServerOpt(
		&ts.TransporterOpt{
			Network: c.Http.Network,
			Addr:    c.Http.Addr,
			Timeout: c.Http.Timeout,
		},
		&ts.MiddlewareOpt{
			ApiKey:         auth.ApiKey,
			WhiteList:      auth.WhiteList,
			Logger:         logger,
			TracerProvider: tp,
		})...)
	return &ts.HttpServer{HttpServer: srv}
}



func NewGRPCServerConf(
	c *conf.Server,
	tp trace.TracerProvider,
	logger log.Logger,
	ac *conf.Auth,
	) *ts.GrpcServer {
	srv := grpc.NewServer(ts.NewGrpcServerOpt(&ts.TransporterOpt{
		Network: c.Grpc.Network,
		Addr:    c.Grpc.Addr,
		Timeout: c.Grpc.Timeout,
	}, &ts.MiddlewareOpt{
		TracerProvider: tp,
		Logger:         logger,
		ApiKey:         ac.ApiKey,
	})...)

	return &ts.GrpcServer{GrpcServer: srv}
}



// 注册服务器
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	reg, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address:           conf.Consul.Address,
		Scheme:            conf.Consul.Scheme,
		EnableHealthCheck: conf.Consul.EnableHealthCheck,
		CheckInterval:     conf.Consul.CheckInterval,
	})
	if err != nil {
		panic(err)
	}
	return reg
}

// 发现服务器
func NewDiscovery(conf *conf.Registry) registry.Discovery {
	dis, err := consul.NewConsulRegistry(&consul.ConsulRegistryOptions{
		Address: conf.Consul.Address,
		Scheme:  conf.Consul.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return dis
}

func NewDB(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			// 表名不使用复数
			SingularTable: true,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}

	return db
}
