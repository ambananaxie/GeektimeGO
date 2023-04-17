package channel

import "context"

type Task func()

type TaskPool struct {
	tasks chan Task
	//close *atomic.Bool

	close chan struct{}

	//closeOnce sync.Once
}

// numG 是 goroutine 数量，就是你要控制住的
// capacity 是缓存的容量
func NewTaskPool(numG int, capacity int) *TaskPool {
	res := &TaskPool{
		tasks: make(chan Task, capacity),
		//close: atomic.NewBool(false),
		close: make(chan struct{}),
	}

	// 这个东西，要是没有退出 goroutine 的机制，那就是妥妥的 goroutine 泄露
	for i := 0; i < numG; i++ {
		go func() {
			for {
				select {
				case <- res.close:
					return
				case t := <- res.tasks:
					t()
				}
			}
			//for t := range res.tasks {
				//if res.close.Load() {
				//	return
				//}
				//t()
			//}
		}()
	}
	return res
}

// Submit 提交任务
func (p *TaskPool) Submit(ctx context.Context, t Task) error {
	select {
	case p.tasks <- t:
	case <- ctx.Done():
		return ctx.Err()
	}
	return nil
}

// Close 方法会释放资源
func (p *TaskPool) Close() error {
	//p.close.Store(true)

	// 这种写法不行
	//p.close <- struct{}{}

	// 这种实现又有一种缺陷
	// 重复调用 Close 方法，会 panic
	close(p.close)

	//p.closeOnce.Do(func() {
	//	close(p.close)
	//})
	return nil
}
