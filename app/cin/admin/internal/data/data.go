package data

import (
	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDashboardRepo,
	NewProductRepo,
	NewProductServiceClient,
)

// Data 包含所有下游微服务的 gRPC 客户端连接
type Data struct {
	pc productRpc.ProductClient
}

// NewData .
func NewData(pc productRpc.ProductClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the admin data resources")
	}
	return &Data{pc: pc}, cleanup, nil
}
