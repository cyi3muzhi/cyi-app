package models

import "cyi-app/pkg/util/gorm"

// AdminUser 管理员用户模型
type AdminUser struct {
	gorm.Model
	Username   string `gorm:"column:username;uniqueIndex;size:32"`
	Password   string `gorm:"column:password;size:128"`
	Name       string `gorm:"column:name;size:64"`
	Role       string `gorm:"column:role;size:32"`
	Email      string `gorm:"column:email;size:128"`
	Status     int32  `gorm:"column:status;default:1"`
	LastLogin  int64  `gorm:"column:last_login;default:0"`
}

// IAdminUser 管理员数据仓库接口
type IAdminUser interface {
	// GetByUsername 根据用户名查询
	// SELECT * FROM @@table WHERE username=@username LIMIT 1
	GetByUsername(username string) (*AdminUser, error)

	// GetByID 根据ID查询
	// SELECT * FROM @@table WHERE id=@id LIMIT 1
	GetByID(id int) (*AdminUser, error)

	// CreateAdminUser 创建管理员
	// INSERT INTO @@table (username,password,name,role,email,status) VALUES (@admin.Username,@admin.Password,@admin.Name,@admin.Role,@admin.Email,@admin.Status)
	CreateAdminUser(admin AdminUser) (*AdminUser, error)

	// UpdateAdminUser 更新管理员
	// UPDATE @@table
	//  {{set}}
	//    {{if admin.Name != ""}} name=@admin.Name, {{end}}
	//    {{if admin.Role != ""}} role=@admin.Role, {{end}}
	//    {{if admin.Email != ""}} email=@admin.Email, {{end}}
	//    {{if admin.Status > 0}} status=@admin.Status, {{end}}
	//    {{if admin.Password != ""}} password=@admin.Password, {{end}}
	//  {{end}}
	// WHERE id=@id
	UpdateAdminUser(id int, admin AdminUser) error

	// ListAdminUser 查询管理员列表
	// SELECT * FROM @@table
	//  {{where}}
	//    {{if keyword != ""}}
	//      username LIKE @keyword
	//    {{end}}
	//  {{end}}
	// ORDER BY id DESC
	ListAdminUser(keyword string) ([]*AdminUser, error)

	// ToggleStatus 切换状态
	// UPDATE @@table SET status=@status WHERE id=@id
	ToggleStatus(id int, status int32) error
}
