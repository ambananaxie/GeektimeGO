package sync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestWaitGroup(t *testing.T) {
	wg := sync.WaitGroup{}
	var result int64 = 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(delta int) {
			defer wg.Done()
			atomic.AddInt64(&result, int64(delta))
		}(i)
	}
	wg.Wait()
	t.Log(result)
}
