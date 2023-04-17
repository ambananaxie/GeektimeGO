package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	q := NewConcurrentBlockingQueue[int](10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := q.DeQueue(ctx)
	assert.Equal(t, context.DeadlineExceeded, err)
}
