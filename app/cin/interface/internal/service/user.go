package service

// 应用层
// 业务编排 DTO
import (
	"context"
	"net"
	"strings"

	v1 "cyi-app/api/cin/interface/v1"
	com "cyi-app/api/common"
	"cyi-app/app/cin/interface/internal/biz"

	service "cyi-app/pkg/kratos/service"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/types/known/emptypb"
)

// getClientIP 从 HTTP 请求中获取客户端真实 IP
// 依次从 X-Forwarded-For / X-Real-IP 获取（代理场景），最后回退 RemoteAddr
func getClientIP(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if ht, ok := tr.(*khttp.Transport); ok && ht.Request() != nil {
			req := ht.Request()
			// X-Forwarded-For 可能为逗号分隔的多个 IP，取第一个
			if ip := strings.TrimSpace(req.Header.Get("X-Forwarded-For")); ip != "" {
				if idx := strings.Index(ip, ","); idx > 0 {
					return strings.TrimSpace(ip[:idx])
				}
				return ip
			}
			if ip := strings.TrimSpace(req.Header.Get("X-Real-IP")); ip != "" {
				return ip
			}
			if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
				return host
			}
			return req.RemoteAddr
		}
	}
	return ""
}

func (s *CinInterface) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	// 客户端 IP 由服务端从请求头获取（前端无需传递）
	if req.ClientIp == "" {
		req.ClientIp = getClientIP(ctx)
	}
	err := s.uc.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.RegisterReply{
		Success: true,
	}, nil
}

func (s *CinInterface) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	// 客户端 IP 由服务端从请求头获取（前端无需传递）
	if req.ClientIp == "" {
		req.ClientIp = getClientIP(ctx)
	}
	return s.uc.Login(ctx, req)
}

// 获取用户信息
func (s *CinInterface) GetMyInfo(ctx context.Context, _ *emptypb.Empty) (*v1.GetMyInfoReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, com.ErrorUnauthorized("用户未登录")
	}

	// 获取用户信息
	user, err := s.uc.UserInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	reply := &v1.GetMyInfoReply{}
	if err := convert.Convert(user, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// 更新用户信息
func (s *CinInterface) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.User{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	in.ID = id

	// 更新用户信息
	err = s.uc.UpdateUserInfo(ctx, in)
	if err != nil {
		return &v1.UpdateUserInfoReply{
			Success: false,
		}, err
	}
	reply := &v1.UpdateUserInfoReply{
		Success: true,
	}
	return reply, nil
}

// 修改密码
func (s *CinInterface) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	err = s.uc.ChangePassword(ctx, id, req.OldPassword, req.NewPassword)
	if err != nil {
		return &v1.ChangePasswordReply{Success: false}, err
	}
	return &v1.ChangePasswordReply{Success: true}, nil
}

// 个人中心概览：用户信息 + 订单/优惠券统计
func (s *CinInterface) GetUserCenter(ctx context.Context, _ *emptypb.Empty) (*v1.GetUserCenterReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	center, err := s.uc.GetUserCenter(ctx, id)
	if err != nil {
		return nil, err
	}
	userReply := &v1.GetMyInfoReply{}
	if err := convert.Convert(center.User, userReply); err != nil {
		return nil, err
	}
	return &v1.GetUserCenterReply{
		User: userReply,
		Stats: &v1.UserCenterStats{
			TotalOrders:    center.Stats.TotalOrders,
			PendingPay:     center.Stats.PendingPay,
			PendingShip:    center.Stats.PendingShip,
			PendingReceive: center.Stats.PendingReceive,
			PendingReview:  center.Stats.PendingReview,
			RefundCount:    center.Stats.RefundCount,
			CouponCount:    center.Stats.CouponCount,
			TotalSpent:     center.Stats.TotalSpent,
		},
	}, nil
}

// 收货地址列表
func (s *CinInterface) ListAddress(ctx context.Context, _ *emptypb.Empty) (*v1.ListAddressReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.uc.ListAddress(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AddressItem, 0, len(list))
	for _, a := range list {
		items = append(items, &v1.AddressItem{
			Id:         a.ID,
			Receiver:   a.Receiver,
			Phone:      a.Phone,
			Province:   a.Province,
			City:       a.City,
			District:   a.District,
			Detail:     a.Detail,
			IsDefault:  a.IsDefault,
			CreateTime: a.CreateTime,
		})
	}
	return &v1.ListAddressReply{Items: items}, nil
}

// 新增收货地址
func (s *CinInterface) CreateAddress(ctx context.Context, req *v1.CreateAddressRequest) (*v1.CreateAddressReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.Address{
		UserID:    id,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}
	created, err := s.uc.CreateAddress(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CreateAddressReply{Id: created.ID}, nil
}

// 更新收货地址
func (s *CinInterface) UpdateAddress(ctx context.Context, req *v1.UpdateAddressRequest) (*v1.UpdateAddressReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.Address{
		ID:        req.Id,
		UserID:    id,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}
	err = s.uc.UpdateAddress(ctx, in)
	if err != nil {
		return &v1.UpdateAddressReply{Success: false}, err
	}
	return &v1.UpdateAddressReply{Success: true}, nil
}

// 删除收货地址
func (s *CinInterface) DeleteAddress(ctx context.Context, req *v1.DeleteAddressRequest) (*v1.DeleteAddressReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	err = s.uc.DeleteAddress(ctx, req.Id, id)
	if err != nil {
		return &v1.DeleteAddressReply{Success: false}, err
	}
	return &v1.DeleteAddressReply{Success: true}, nil
}

// 设置默认收货地址
func (s *CinInterface) SetDefaultAddress(ctx context.Context, req *v1.SetDefaultAddressRequest) (*v1.SetDefaultAddressReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	err = s.uc.SetDefaultAddress(ctx, req.Id, id)
	if err != nil {
		return &v1.SetDefaultAddressReply{Success: false}, err
	}
	return &v1.SetDefaultAddressReply{Success: true}, nil
}