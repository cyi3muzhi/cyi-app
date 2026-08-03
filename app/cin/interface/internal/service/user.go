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