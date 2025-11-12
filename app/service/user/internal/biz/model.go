package biz

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
	Password string `json:"password"`
	Username string `json:"username"`
}
