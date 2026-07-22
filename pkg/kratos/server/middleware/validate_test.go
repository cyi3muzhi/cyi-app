package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockValidatable struct {
	valid bool
}

func (m *mockValidatable) Validate() error {
	if !m.valid {
		return assert.AnError
	}
	return nil
}

type mockNonValidatable struct{}

func TestValidateMiddleware_Valid(t *testing.T) {
	mw := Validate()
	req := &mockValidatable{valid: true}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	_, err := mw(handler)(context.Background(), req)
	assert.NoError(t, err)
}

func TestValidateMiddleware_Invalid(t *testing.T) {
	mw := Validate()
	req := &mockValidatable{valid: false}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	_, err := mw(handler)(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "参数校验失败")
}

func TestValidateMiddleware_NoValidateMethod(t *testing.T) {
	mw := Validate()
	req := &mockNonValidatable{}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	_, err := mw(handler)(context.Background(), req)
	assert.NoError(t, err)
}