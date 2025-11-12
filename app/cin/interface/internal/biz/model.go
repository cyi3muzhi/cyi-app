package biz

type User struct {
	ID int64 `json:"id"`
	// 姓名
	Name string `json:"name"`
	// 手机号
	Phone string `json:"phone"`
	// 密码
	Password string `json:"password"`
	// 用户名
	Username string `json:"username"`
}
