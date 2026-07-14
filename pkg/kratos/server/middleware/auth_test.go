package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGetTokenFromContext(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantToken string
		wantOK    bool
	}{
		{
			name:      "valid token",
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantOK:    true,
		},
		{
			name:      "empty token",
			token:     "",
			wantToken: "",
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := SetTokenToContext(context.Background(), tt.token)
			got, ok := GetTokenFromContext(ctx)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantToken, got)
		})
	}
}

func TestGetTokenFromContext_NilContext(t *testing.T) {
	_, ok := GetTokenFromContext(nil)
	assert.False(t, ok)
}

func TestGetTokenFromContext_EmptyContext(t *testing.T) {
	_, ok := GetTokenFromContext(context.Background())
	assert.False(t, ok)
}

func TestNewWhiteListMatcher(t *testing.T) {
	tests := []struct {
		name      string
		whiteList []string
		operation string
		want      bool
	}{
		{
			name:      "in white list",
			whiteList: []string{"/api.cin.admin.v1.AdminAuth/Login"},
			operation: "/api.cin.admin.v1.AdminAuth/Login",
			want:      false,
		},
		{
			name:      "not in white list",
			whiteList: []string{"/api.cin.admin.v1.AdminAuth/Login"},
			operation: "/api.cin.admin.v1.AdminAuth/Logout",
			want:      true,
		},
		{
			name:      "nil white list",
			whiteList: nil,
			operation: "/api.cin.admin.v1.AdminAuth/Login",
			want:      false,
		},
		{
			name:      "empty white list",
			whiteList: []string{},
			operation: "/api.cin.admin.v1.AdminAuth/Login",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewWhiteListMatcher(tt.whiteList)
			got := matcher(context.Background(), tt.operation)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserClaims(t *testing.T) {
	claims := &UserClaims{
		ID:       1,
		Username: "admin",
		RegisteredClaims: jwt2.RegisteredClaims{
			ExpiresAt: jwt2.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "cin",
		},
	}

	token := jwt2.NewWithClaims(jwt2.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)
	assert.NotEmpty(t, signedToken)
}

func TestAuthServiceMiddleware_WhiteList(t *testing.T) {
	auth := Auth{
		ApiKey:    "test-secret",
		WhiteList: []string{"/api.cin.admin.v1.AdminAuth/Login"},
	}

	mw := AuthServiceMiddleware(auth)
	assert.NotNil(t, mw)
}

func TestAuthServiceMiddleware_EmptyApiKey(t *testing.T) {
	auth := Auth{
		ApiKey: "",
	}

	assert.Panics(t, func() {
		AuthServiceMiddleware(auth)
	})
}

func TestTokenValidator_Success(t *testing.T) {
	validatorCalled := false
	testValidator := TokenValidator(func(ctx context.Context, rawToken string, claims *UserClaims) error {
		validatorCalled = true
		assert.Equal(t, "test-token", rawToken)
		assert.Equal(t, int64(1), claims.ID)
		assert.Equal(t, "admin", claims.Username)
		return nil
	})

	err := testValidator(context.Background(), "test-token", &UserClaims{ID: 1, Username: "admin"})

	assert.NoError(t, err)
	assert.True(t, validatorCalled)
}

func TestTokenValidator_Failure(t *testing.T) {
	expectedErr := errors.New("token is blacklisted")
	testValidator := TokenValidator(func(ctx context.Context, rawToken string, claims *UserClaims) error {
		return expectedErr
	})

	err := testValidator(context.Background(), "test-token", &UserClaims{ID: 1, Username: "admin"})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAuthServiceMiddleware_WithTokenValidator_Config(t *testing.T) {
	auth := Auth{
		ApiKey: "test-secret",
		TokenValidator: func(ctx context.Context, rawToken string, claims *UserClaims) error {
			return nil
		},
	}

	mw := AuthServiceMiddleware(auth)
	assert.NotNil(t, mw)
}

func TestAuthServiceMiddleware_WithoutTokenValidator_Config(t *testing.T) {
	auth := Auth{
		ApiKey: "test-secret",
	}

	mw := AuthServiceMiddleware(auth)
	assert.NotNil(t, mw)
}

func TestAuthClientMiddleware(t *testing.T) {
	auth := Auth{
		ServiceKey: "test-service-key",
	}

	mw := AuthClientMiddleware(auth)
	assert.NotNil(t, mw)
}