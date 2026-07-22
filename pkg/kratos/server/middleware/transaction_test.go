package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransactionListMatcher(t *testing.T) {
	matcher := NewTransactionListMatcher([]string{"CreateProduct", "UpdateOrder"})

	assert.True(t, matcher(context.Background(), "CreateProduct"))
	assert.True(t, matcher(context.Background(), "UpdateOrder"))
	assert.False(t, matcher(context.Background(), "GetProduct"))
	assert.False(t, matcher(context.Background(), "ListOrder"))
}

func TestNewTransactionListMatcher_Nil(t *testing.T) {
	matcher := NewTransactionListMatcher(nil)

	assert.False(t, matcher(context.Background(), "CreateProduct"))
	assert.False(t, matcher(context.Background(), "UpdateOrder"))
}

func TestNewTransactionListMatcher_Empty(t *testing.T) {
	matcher := NewTransactionListMatcher([]string{})

	assert.False(t, matcher(context.Background(), "CreateProduct"))
}