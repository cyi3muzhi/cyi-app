package encryption

import (
	"strings"
	"testing"
)

// TestEncryptPassword 测试密码加密
func TestEncryptPassword(t *testing.T) {
	password := "testPassword123"

	hashed, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("EncryptPassword failed: %v", err)
	}

	// 加密后的密码不应该等于原始密码
	if hashed == password {
		t.Error("hashed password should not equal original password")
	}

	// 加密后的密码应该以$2a$开头（bcrypt格式）
	if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") {
		t.Errorf("hashed password should start with $2a$ or $2b$, got: %s", hashed)
	}
}

// TestEncryptPassword_DifferentResults 测试相同密码每次加密结果不同（因为有盐值）
func TestEncryptPassword_DifferentResults(t *testing.T) {
	password := "testPassword123"

	hashed1, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("first EncryptPassword failed: %v", err)
	}

	hashed2, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("second EncryptPassword failed: %v", err)
	}

	// 由于盐值不同，两次加密结果应该不同
	if hashed1 == hashed2 {
		t.Error("two hashes of same password should be different due to salt")
	}
}

// TestCheckPassword 测试密码验证
func TestCheckPassword(t *testing.T) {
	password := "testPassword123"

	hashed, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("EncryptPassword failed: %v", err)
	}

	// 正确的密码应该验证通过
	if !CheckPassword(password, hashed) {
		t.Error("CheckPassword should return true for correct password")
	}

	// 错误的密码应该验证失败
	if CheckPassword("wrongPassword", hashed) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

// TestCheckPassword_EmptyPassword 测试空密码
func TestCheckPassword_EmptyPassword(t *testing.T) {
	// 空密码也应该能加密和验证
	hashed, err := EncryptPassword("")
	if err != nil {
		t.Fatalf("EncryptPassword with empty password failed: %v", err)
	}

	if !CheckPassword("", hashed) {
		t.Error("CheckPassword should return true for empty password that matches")
	}

	if CheckPassword("non-empty", hashed) {
		t.Error("CheckPassword should return false for non-empty password against empty hash")
	}
}

// TestCheckPassword_VariousPasswords 测试各种密码组合
func TestCheckPassword_VariousPasswords(t *testing.T) {
	passwords := []string{
		"short",
		"this is a much longer password with spaces",
		"password123!@#",
		"中文密码测试",
		"😀😃😄😁",
		"1234567890",
		"a",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hashed, err := EncryptPassword(password)
			if err != nil {
				t.Fatalf("EncryptPassword failed: %v", err)
			}

			if !CheckPassword(password, hashed) {
				t.Error("CheckPassword should return true for matching password")
			}

			if CheckPassword(password+"wrong", hashed) {
				t.Error("CheckPassword should return false for non-matching password")
			}
		})
	}
}

// TestEncryptPassword_LongPassword 测试长密码（bcrypt有72字节限制）
func TestEncryptPassword_LongPassword(t *testing.T) {
	// bcrypt不支持超过72字节的密码，会返回错误
	longPassword := strings.Repeat("a", 100)

	_, err := EncryptPassword(longPassword)
	if err == nil {
		t.Error("expected error for password longer than 72 bytes, got nil")
	}

	// 72字节及以下的密码应该能正常加密
	validPassword := strings.Repeat("a", 72)
	hashed, err := EncryptPassword(validPassword)
	if err != nil {
		t.Fatalf("EncryptPassword with 72-byte password failed: %v", err)
	}

	if !CheckPassword(validPassword, hashed) {
		t.Error("CheckPassword should return true for matching 72-byte password")
	}
}

// TestCheckPassword_InvalidHash 测试无效的哈希值
func TestCheckPassword_InvalidHash(t *testing.T) {
	// 无效的哈希值应该返回false
	result := CheckPassword("password", "invalid-hash")
	if result {
		t.Error("CheckPassword should return false for invalid hash")
	}
}
