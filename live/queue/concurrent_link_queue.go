package queue

import (
	"context"
	"fmt"
	"sync/atomic"
	"unsafe"
)

type ConcurrentLinkedQueue[T any] struct {
	// 就是指针，任意类型的指针
	head unsafe.Pointer
	tail unsafe.Pointer
	count uint64
}

func NewConcurrentLinkedQueue[T any]() *ConcurrentLinkedQueue[T] {
	head := &node[T]{}
	ptr := unsafe.Pointer(head)
	return &ConcurrentLinkedQueue[T]{
		head: ptr,
		tail: ptr,
	}
}

func (c *ConcurrentLinkedQueue[T]) EnQueue(ctx context.Context, data T) error {

	// for {
	// 	select {
	// 	case <- ctx.Done():
	// 		return
		// case <- time.After(10 * time.Second):
	// 	default:
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }

	newNode := &node[T]{
		val: data,
	}
	newNodePtr := unsafe.Pointer(newNode)

	// 先改 tail
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// select tail; => tail = 4
		tail := atomic.LoadPointer(&c.tail)
		// 为什么不能这样写？
		// tail = c.tail // 这种是非线程安全

		// Update Set tail = 3 WHERE tail = 4
		if atomic.CompareAndSwapPointer(&c.tail,
			tail, newNodePtr) {
			// 在这一步，就要讲 tail.next 指向 c.tail
			// tail.next = c.tail
			tailNode := (*node[T])(tail)
			// 你在这一步，c.tail 被人修改了

			atomic.StorePointer(&tailNode.next, newNodePtr)
			atomic.AddUint64(&c.count, 1)
			return nil
		}
	}

	// 先改 tail.next
	// newNode := &node[T]{val: t}
	// newPtr := unsafe.Pointer(newNode)
	// for {
	// 	tailPtr := atomic.LoadPointer(&c.tail)
	// 	tail := (*node[T])(tailPtr)
	// 	tailNext := atomic.LoadPointer(&tail.next)
	// 	if tailNext != nil {
	// 		// 已经被人修改了，我们不需要修复，因为预期中修改的那个人会把 c.tail 指过去
	// 		continue
	// 	}
	// 	if atomic.CompareAndSwapPointer(&tail.next, tailNext, newPtr) {
	// 		// 如果失败也不用担心，说明有人抢先一步了
	// 		atomic.CompareAndSwapPointer(&c.tail, tailPtr, newPtr)
	// 		return nil
	// 	}
	// }

	// 先改 tail next
	// for {
	// 	if atomic.CompareAndSwapPointer(&c.tail.next,
	// 		tailNext, unsafe.Pointer(newNode)) {
	//
	// 	}
	// }

	// return nil
}

func (c *ConcurrentLinkedQueue[T]) DeQueue(ctx context.Context) (T, error) {
	for {
		if ctx.Err() != nil {
			var t T
			return t, ctx.Err()
		}
		// select {
		// case <- ctx.Done():
		//
		// default:
		//
		// }
		headPtr := atomic.LoadPointer(&c.head)
		head := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&c.tail)
		tail := (*node[T])(tailPtr)
		if head == tail {
			// 不需要做更多检测，在当下这一刻，我们就认为没有元素，即便这时候正好有人入队
			// 但是并不妨碍我们在它彻底入队完成——即所有的指针都调整好——之前，
			// 认为其实还是没有元素
			var t T
			return t, ErrEmptyQueue
		}
		headNextPtr := atomic.LoadPointer(&head.next)
		// 如果到这里为空了，CAS 操作不会成功。因为原本的数据，被人拿走了
		if atomic.CompareAndSwapPointer(&c.head, headPtr, headNextPtr) {
			headNext := (*node[T])(headNextPtr)
			return headNext.val, nil
		}
	}
}

func (c *ConcurrentLinkedQueue[T]) IsFull() bool {
	fmt.Println()
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) IsEmpty() bool {
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) Len() uint64 {
	// 在你读的过程中，就被人改了
	return atomic.LoadUint64(&c.count)
}

type node[T any] struct {
	next unsafe.Pointer
	val T
}