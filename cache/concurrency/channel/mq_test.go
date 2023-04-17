package channel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBroker_Send(t *testing.T) {
	b := &Broker{}

	// 模拟发送者
	go func() {
		for {
			err := b.Send(Msg{Content: time.Now().String()})
			if err != nil {
				t.Log(err)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("消费者 %d", i)
		go func() {
			defer wg.Done()
			msgs, err := b.Subscribe(100)
			if err != nil {
				t.Log(err)
				return
			}

			//select {
			//case m := <- msgs:
			//case <- ctx.Done():
			//}
			for msg := range msgs {
				fmt.Println(name, msg.Content)
			}
		}()
	}
	wg.Wait()
}
