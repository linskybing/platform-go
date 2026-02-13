package k8s

import (
	"context"
	"time"
)

const defaultRequestTimeout = 10 * time.Second

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultRequestTimeout)
}
