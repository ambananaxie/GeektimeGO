package queue

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// func TestConcurrentBlockingQueue_EnQueue(t *testing.T) {
// 	testCases := []struct{
// 		name string
//
// 		q *ConcurrentBlockingQueue[int]
//
// 		timeout time.Duration
// 		value int
//
// 		data []int
//
// 		wantErr error
// 	}{
// 		{
// 			name: "enqueue",
// 			q: NewConcurrentBlockingQueue[int](10),
// 			value: 1,
// 			timeout: time.Minute,
// 			data: []int{1},
// 		},
// 		{
// 			name: "blocking and timeout",
// 			q: func() *ConcurrentBlockingQueue[int]{
// 				res := NewConcurrentBlockingQueue[int](2)
// 				ctx, cancel :=context.WithTimeout(context.Background(), time.Second)
// 				defer cancel()
// 				err := res.EnQueue(ctx, 1)
// 				require.NoError(t, err)
// 				err = res.EnQueue(ctx, 2)
// 				require.NoError(t, err)
// 				return res
// 			}(),
// 			value: 3,
// 			timeout: time.Second,
// 			data: []int{1, 2},
// 			wantErr: context.DeadlineExceeded,
// 		},
// 	}
//
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
// 			defer cancel()
// 			err := tc.q.EnQueue(ctx, tc.value)
// 			assert.Equal(t, tc.wantErr, err)
// 			assert.Equal(t, tc.data, tc.q.data)
// 		})
// 	}
// }

func TestConcurrentBlockingQueue(t *testing.T) {
	// 只能确保没有死锁
	q := NewConcurrentBlockingQueue[int](10000)
	// data := make(chan int, 10000000000000000000000)

	// 并发的问题都落在 m 上
	// var m sync.Mutex
	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				// 你没有办法校验这里面的中间结果
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				// m.Lock()
				val := rand.Int()
				_ = q.EnQueue(ctx, val)
				// 怎么断言 error
				// data <- val
				// m.Unlock()
				cancel()
			}
			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, _ = q.DeQueue(ctx)
				// 怎么断言 error
				cancel()
			}
			wg.Done()
		}()
	}

	// 怎么校验 q 对还是不对
	wg.Wait()
}

// 切片实现
// BenchmarkConcurrentQueue-12       100000               306.7 ns/op           223 B/op          4 allocs/op

// BenchmarkConcurrentQueue-12       100000               280.8 ns/op           208 B/op          4 allocs/op
func BenchmarkConcurrentQueue(b *testing.B) {
	var wg sync.WaitGroup
	q := NewConcurrentBlockingQueue[int](100)
	wg.Add(2)
	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			_ = q.EnQueue(context.Background(), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			_, _ = q.DeQueue(context.Background())
		}
		wg.Done()
	}()
	wg.Wait()
}


