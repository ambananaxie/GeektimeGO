package queue

import "github.com/google/uuid"

type Lock struct {
	ch chan string
	val string
}

// 就是要确保加锁和解锁是同一个人，有点像是 Redis 那个分布式锁
func NewLock() *Lock {
	val := uuid.New().String()
	res := &Lock{
		ch: make(chan string, 1),
		val: val,
	}

	res.ch <- val
	return res
}

func (l *Lock) Lock() string {
	<- l.ch
	return l.val
}

func (l *Lock) Unlock(id string) {
	if id != l.val {
		panic("不是你的锁")
	}
	val := uuid.New().String()
	select {
	case l.ch <- val:
		l.val = val
	default:
		panic("你没有加锁")
	}

	return
}

type Demo struct {
	ch chan any
}

//data := <- d.ch

// func (d Demo) Produce(data any) {
// 	select {
// 	case d.ch <- data:
// 	default:
//
// 	}
// }