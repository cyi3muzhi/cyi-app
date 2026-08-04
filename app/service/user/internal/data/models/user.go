package models

import (
	"cyi-app/app/service/user/internal/biz"
	"cyi-app/pkg/util/gorm"
)

// User 用户端用户模型
// 用户端用户与管理员语义不同：无 role/department/position 等管理属性，
// 但需要支持管理后台的状态管理、会员等级（预留）等扩展字段。
type User struct {
	gorm.Model
	// 登录与基础认证
	Username       string `gorm:"column:username;type:varchar(32);uniqueIndex;not null;comment:登录账号"`
	Phone          string `gorm:"column:phone;type:varchar(20);index;comment:手机号"`
	Email          string `gorm:"column:email;type:varchar(128);index;comment:邮箱"`
	Password       string `gorm:"column:password;type:varchar(128);not null;comment:加密后的密码"`

	// 个人资料
	Nickname string `gorm:"column:nickname;type:varchar(64);comment:展示名"`
	Avatar   string `gorm:"column:avatar;type:varchar(512);comment:头像URL"`
	Gender   int32  `gorm:"column:gender;type:tinyint;default:0;comment:性别 0=未知 1=男 2=女"`
	Birthday string `gorm:"column:birthday;type:varchar(10);comment:生日YYYY-MM-DD"`

	// 状态与登录信息
	Status         int32  `gorm:"column:status;type:tinyint;default:1;index;comment:状态 1=正常 2=冻结"`
	LastLoginAt    int64  `gorm:"column:last_login_at;type:bigint;default:0;comment:最近登录时间戳"`
	LastLoginIP    string `gorm:"column:last_login_ip;type:varchar(45);comment:最近登录IP"`
	RegisterSource string `gorm:"column:register_source;type:varchar(32);default:web;comment:注册来源"`

	// 会员预留扩展（user 表只存等级标记，完整会员体系后续单独建表）
	Level       string `gorm:"column:level;type:varchar(32);default:regular;index;comment:会员等级"`
	GrowthValue int64  `gorm:"column:growth_value;type:bigint;default:0;comment:成长值"`
}

// TableName 显式指定表名（与 SingularTable:true 配合，表名为 user）
func (User) TableName() string { return "user" }

// Address 用户收货地址
type Address struct {
	gorm.Model
	UserID    int64  `gorm:"column:user_id;type:int;index;not null;comment:关联用户ID"`
	Receiver  string `gorm:"column:receiver;type:varchar(64);not null;comment:收货人姓名"`
	Phone     string `gorm:"column:phone;type:varchar(20);not null;comment:收货人手机号"`
	Province  string `gorm:"column:province;type:varchar(32);comment:省份"`
	City      string `gorm:"column:city;type:varchar(32);comment:城市"`
	District  string `gorm:"column:district;type:varchar(32);comment:区县"`
	Detail    string `gorm:"column:detail;type:varchar(255);not null;comment:详细地址"`
	IsDefault int32  `gorm:"column:is_default;type:tinyint;default:0;index;comment:是否默认地址 0=否 1=是"`
}

// TableName 显式指定表名
func (Address) TableName() string { return "user_address" }

// IUser 用户数据仓库接口（gorm gen 模板方法）
type IUser interface {
	// GetByID
	// SELECT * FROM @@table WHERE id=@id
	GetByID(id int) (*biz.User, error)

	// FilterWithColumn
	// SELECT * FROM @@table WHERE @@column=@value
	FilterWithColumn(column string, value string) (User, error)

	// QueryWith
	// SELECT * FROM @@table
	//   {{if user.ID > 0}}
	//       WHERE id=@user.ID
	//   {{else if user.Username != ""}}
	//       WHERE username=@user.Username
	//   {{else if user.Phone != ""}}
	//       WHERE phone=@user.Phone
	//   {{end}}
	QueryWith(user User) (User, error)

	// CreateUser
	// INSERT INTO @@table (username, phone, email, password, nickname, register_source, level, status)
	// VALUES (@user.Username, @user.Phone, @user.Email, @user.Password, @user.Nickname, @user.RegisterSource, @user.Level, @user.Status)
	CreateUser(user User) (User, error)
}

type IAddress interface {
	// QueryAddress
	// SELECT * FROM @@table WHERE user_id=@userID
	QueryAddress(userID int) ([]Address, error)
}
