package encryption

import (
	"golang.org/x/crypto/bcrypt"
)

// 加密密码
func EncryptPassword(password string) (string, error) {
	// 成本因子，值越大计算越慢，安全性越高，推荐值：10-12
	cost := 12

	// 自动生成盐值并加密
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// 验证密码
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
