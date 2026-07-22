package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceIDHeader(t *testing.T) {
	mw := TraceIDHeader()

	req := struct{}{}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	_, err := mw(handler)(context.Background(), req)
	assert.NoError(t, err)
}

func TestGetTraceID(t *testing.T) {
	traceID := getTraceID(context.Background())
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 32)
}