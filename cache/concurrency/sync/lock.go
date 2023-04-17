package sync

import "sync/atomic"

const (
	UNLOCK int32 = 0
	LOCKED int32 =1
)

type Lock struct {
	state int32
}

func (l *Lock) Lock() {
	i := 0
	var locked = false
	for locked = atomic.CompareAndSwapInt32(&l.state, UNLOCK, LOCKED);  !locked && i < 10; i++{
	}

	if locked {
		return
	}

	// 加入队列
	// enqueue()

	// 到这里别人把你唤醒了
	// 再去竞争锁
}