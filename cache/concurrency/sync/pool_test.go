package sync

import (
	"sync"
	"testing"
)

func TestPool(t *testing.T) {
	p := sync.Pool{
		New: func() any {
			t.Log("创建资源了")
			return "hello"
			// 最好永远不要返回 nil
		},
	}

	str := p.Get().(string)
	t.Log(str)
	p.Put(str)
	str = p.Get().(string)
	t.Log(str)
	p.Put(str)
}
