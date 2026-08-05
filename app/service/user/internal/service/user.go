package service

import (
	"context"
	"errors"
	"fmt"

	pb "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UserService 用户服务实现
type UserService struct {
	pb.UnimplementedUserServer
	uc        *biz.UserUseCase
	addressUC *biz.AddressUseCase
	log       *log.Helper
}

// NewUserService 创建 UserService
func NewUserService(uc *biz.UserUseCase, addressUC *biz.AddressUseCase, logger log.Logger) *UserService {
	return &UserService{
		uc:        uc,
		addressUC: addressUC,
		log:       log.NewHelper(logger),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, convertBizError(err)
	}
	user, err := s.uc.CreateUser(ctx, u)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.CreateUserReply{Id: user.ID}, nil
}

// UpdateUser 更新用户资料
func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, convertBizError(err)
	}
	u.ID = req.Id
	if err := s.uc.UpdateUser(ctx, u); err != nil {
		return &pb.UpdateUserReply{Success: false}, convertBizError(err)
	}
	return &pb.UpdateUserReply{Success: true}, nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	if err := s.uc.DeleteUser(ctx, req.Id); err != nil {
		return nil, convertBizError(err)
	}
	return &pb.DeleteUserReply{Success: true}, nil
}

// GetUser 获取用户详情
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	user, err := s.uc.GetUser(ctx, &biz.User{ID: req.Id})
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.GetUserReply{User: userBizToProto(user)}, nil
}

// ListUser 获取用户列表（支持分页/搜索/筛选）
func (s *UserService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
	reply, err := s.uc.ListUser(ctx, &biz.UserListRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Level:    req.Level,
		Status:   req.Status,
	})
	if err != nil {
		return nil, convertBizError(err)
	}
	items := make([]*pb.UserInfo, 0, len(reply.Items))
	for _, u := range reply.Items {
		items = append(items, userBizToProto(u))
	}
	return &pb.ListUserReply{Items: items, Total: reply.Total}, nil
}

// ToggleUserStatus 冻结/解冻用户
func (s *UserService) ToggleUserStatus(ctx context.Context, req *pb.ToggleUserStatusRequest) (*pb.ToggleUserStatusReply, error) {
	newStatus, err := s.uc.ToggleUserStatus(ctx, req.Id)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.ToggleUserStatusReply{Success: true, NewStatus: newStatus}, nil
}

// GetUserStats 用户统计数据
func (s *UserService) GetUserStats(ctx context.Context, _ *emptypb.Empty) (*pb.GetUserStatsReply, error) {
	stats, err := s.uc.GetUserStats(ctx)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.GetUserStatsReply{
		TotalUsers:     stats.TotalUsers,
		TodayNew:       stats.TodayNew,
		TodayNewGrowth: stats.TodayGrowth,
		ActiveUsers:    stats.ActiveUsers,
		MemberUsers:    stats.MemberUsers,
		MemberRatio:    stats.MemberRatio,
	}, nil
}

// UpdateUserAvatar 单独更新用户头像
func (s *UserService) UpdateUserAvatar(ctx context.Context, req *pb.UpdateUserAvatarRequest) (*pb.UpdateUserAvatarReply, error) {
	if err := s.uc.UpdateUserAvatar(ctx, req.UserId, req.AvatarUrl); err != nil {
		return nil, convertBizError(err)
	}
	return &pb.UpdateUserAvatarReply{Success: true}, nil
}

// CheckUserExist 校验用户是否存在
func (s *UserService) CheckUserExist(ctx context.Context, req *pb.CheckUserExistRequest) (*pb.CheckUserExistReply, error) {
	fmt.Println("phone:", req.Phone)
	// 查询用户是否存在
	user, err := s.uc.GetUser(ctx, &biz.User{Phone: req.Phone})
	if err != nil {
		return nil, convertBizError(err)
	}

	if user.ID != 0 {
		return &pb.CheckUserExistReply{Exist: true}, nil
	}
	// 查询用户是否存在
	user, err = s.uc.GetUser(ctx, &biz.User{Username: req.Username})
	if err != nil {
		return nil, convertBizError(err)
	}

	if user.ID != 0 {
		return &pb.CheckUserExistReply{Exist: true}, nil
	}

	return &pb.CheckUserExistReply{Exist: false}, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, in *pb.GetUserByUsernameReq) (*pb.GetUserByUsernameReply, error) {
	user, err := s.uc.GetUserByUsername(ctx, in)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.GetUserByUsernameReply{
		Id:       user.ID,
		Username: user.Username,
	}, nil
}

// VerifyPassword 校验密码
func (s *UserService) VerifyPassword(ctx context.Context, in *pb.VerifyPasswordRequest) (*pb.VerifyPasswordReply, error) {
	u := &biz.User{}
	if err := convert.Convert(in, u); err != nil {
		return nil, convertBizError(err)
	}

	if err := s.uc.VerifyPassword(ctx, u); err != nil {
		return &pb.VerifyPasswordReply{Valid: false}, convertBizError(err)
	}
	return &pb.VerifyPasswordReply{Valid: true}, nil
}

// UserCount 统计用户数量
func (s *UserService) UserCount(ctx context.Context, _ *emptypb.Empty) (*pb.UserCountReply, error) {
	count, err := s.uc.UserCount(ctx)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.UserCountReply{Count: count}, nil
}

// Register 注册用户
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, convertBizError(err)
	}
	// 客户端 IP（注册即登录时记录）
	u.LastLoginIP = req.ClientIp
	if err := s.uc.Register(ctx, u); err != nil {
		return nil, convertBizError(err)
	}
	return &pb.RegisterReply{Success: true}, nil
}

// userBizToProto biz.User -> pb.UserInfo 转换
// 手动转换所有字段（包含嵌入字段 CreatedAt/UpdatedAt）
func userBizToProto(u *biz.User) *pb.UserInfo {
	if u == nil {
		return nil
	}
	return &pb.UserInfo{
		Id:             u.ID,
		Username:       u.Username,
		Phone:          u.Phone,
		Email:          u.Email,
		Nickname:       u.Nickname,
		Avatar:         u.Avatar,
		Gender:         u.Gender,
		Birthday:       u.Birthday,
		Status:         u.Status,
		Level:          u.Level,
		GrowthValue:    u.GrowthValue,
		LastLoginAt:    u.LastLoginAt,
		LastLoginIp:    u.LastLoginIP,
		RegisterSource: u.RegisterSource,
		CreatedAt:      u.CreatedAt.Unix(),
		UpdatedAt:      u.UpdatedAt.Unix(),
	}
}

// convertBizError 将 biz 层错误转换为 proto 错误
func convertBizError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, biz.ErrUserNotFound) {
		return pb.ErrorUserNotFound("%s", err.Error())
	}
	if errors.Is(err, biz.ErrPasswordError) {
		return pb.ErrorPasswordError("%s", err.Error())
	}
	if errors.Is(err, biz.ErrUserAlreadyRegistered) {
		return pb.ErrorUserAlreadyRegistered("%s", err.Error())
	}
	if errors.Is(err, biz.ErrUserRegisterFailed) {
		return pb.ErrorUserRegisterFailed("%s", err.Error())
	}
	if errors.Is(err, biz.ErrMessageNotFound) {
		return pb.ErrorMessageNotFound("%s", err.Error())
	}
	if errors.Is(err, biz.ErrMessageCreateFailed) {
		return pb.ErrorMessageCreateFailed("%s", err.Error())
	}
	return err
}

// Login 登录
func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, convertBizError(err)
	}
	// 客户端 IP（登录成功后更新 last_login_at / last_login_ip）
	u.LastLoginIP = req.ClientIp
	reply, err := s.uc.Login(ctx, u, req.ApiKey)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.LoginReply{
		Token: reply.Token,
	}, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordReply, error) {
	if err := s.uc.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword); err != nil {
		return &pb.ChangePasswordReply{Success: false}, convertBizError(err)
	}
	return &pb.ChangePasswordReply{Success: true}, nil
}

// ListAddress 查询用户收货地址列表
func (s *UserService) ListAddress(ctx context.Context, req *pb.ListAddressRequest) (*pb.ListAddressReply, error) {
	list, err := s.addressUC.ListAddress(ctx, req.UserId)
	if err != nil {
		return nil, convertBizError(err)
	}
	items := make([]*pb.AddressInfo, 0, len(list))
	for _, a := range list {
		items = append(items, addressBizToProto(a))
	}
	return &pb.ListAddressReply{Items: items}, nil
}

// CreateAddress 新增收货地址
func (s *UserService) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.CreateAddressReply, error) {
	a := &biz.Address{
		UserID:    req.UserId,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}
	created, err := s.addressUC.CreateAddress(ctx, a)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.CreateAddressReply{Id: created.ID}, nil
}

// UpdateAddress 更新收货地址
func (s *UserService) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.UpdateAddressReply, error) {
	a := &biz.Address{
		ID:        req.Id,
		UserID:    req.UserId,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	}
	if err := s.addressUC.UpdateAddress(ctx, a); err != nil {
		return &pb.UpdateAddressReply{Success: false}, convertBizError(err)
	}
	return &pb.UpdateAddressReply{Success: true}, nil
}

// DeleteAddress 删除收货地址
func (s *UserService) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.DeleteAddressReply, error) {
	if err := s.addressUC.DeleteAddress(ctx, req.Id, req.UserId); err != nil {
		return &pb.DeleteAddressReply{Success: false}, convertBizError(err)
	}
	return &pb.DeleteAddressReply{Success: true}, nil
}

// SetDefaultAddress 设置默认收货地址
func (s *UserService) SetDefaultAddress(ctx context.Context, req *pb.SetDefaultAddressRequest) (*pb.SetDefaultAddressReply, error) {
	if err := s.addressUC.SetDefaultAddress(ctx, req.Id, req.UserId); err != nil {
		return &pb.SetDefaultAddressReply{Success: false}, convertBizError(err)
	}
	return &pb.SetDefaultAddressReply{Success: true}, nil
}

// addressBizToProto biz.Address -> pb.AddressInfo 转换
func addressBizToProto(a *biz.Address) *pb.AddressInfo {
	if a == nil {
		return nil
	}
	return &pb.AddressInfo{
		Id:         a.ID,
		Receiver:   a.Receiver,
		Phone:      a.Phone,
		Province:   a.Province,
		City:       a.City,
		District:   a.District,
		Detail:     a.Detail,
		IsDefault:  a.IsDefault,
		CreateTime: a.CreateTime,
		UpdateTime: a.UpdateTime,
	}
}
