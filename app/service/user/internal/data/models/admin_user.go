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
	Phone      string `gorm:"column:phone;size:20"`
	Department string `gorm:"column:department;size:64"`
	Position   string `gorm:"column:position;size:64"`
	Bio        string `gorm:"column:bio;size:512"`
	Avatar     string `gorm:"column:avatar;size:512"`
	Status     int32  `gorm:"column:status;default:1"`
	LastLogin  int64  `gorm:"column:last_login;default:0"`
}

// IAdminUser 管理员数据仓库接口
type IAdminUser interface {
	// GetByUsername
	// SELECT * FROM @@table WHERE username=@username LIMIT 1
	GetByUsername(username string) (*AdminUser, error)

	// GetByID
	// SELECT * FROM @@table WHERE id=@id LIMIT 1
	GetByID(id int) (*AdminUser, error)

	// CreateAdminUser
	// INSERT INTO @@table (username,password,name,role,email,status) VALUES (@admin.Username,@admin.Password,@admin.Name,@admin.Role,@admin.Email,@admin.Status)
	CreateAdminUser(admin AdminUser) (*AdminUser, error)

	// UpdateAdminUser
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

	// ListAdminUser
	// SELECT * FROM @@table
	//  {{where}}
	//    {{if keyword != ""}}
	//      username LIKE @keyword
	//    {{end}}
	//  {{end}}
	// ORDER BY id DESC
	ListAdminUser(keyword string) ([]*AdminUser, error)

	// ToggleStatus
	// UPDATE @@table SET status=@status WHERE id=@id
	ToggleStatus(id int, status int32) error
}
