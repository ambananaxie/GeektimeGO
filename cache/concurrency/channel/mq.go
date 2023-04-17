package channel

import (
	"errors"
	"sync"
)

type Broker struct {
	mutex sync.RWMutex
	chans []chan Msg
	//chans map[string][]chan Msg
}

func (b *Broker) Send(m Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, ch := range b.chans {
		select {
		case ch <- m:
		default:
			return errors.New("消息队列已满")
		}
	}
	return nil
}

func (b *Broker) Subscribe(capacity int) (<- chan Msg, error) {
	res := make(chan Msg, capacity)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.chans = append(b.chans, res)
	return res, nil
}

func (b *Broker) Close() error {
	b.mutex.Lock()
	chans := b.chans
	b.chans = nil
	b.mutex.Unlock()

	// 避免了重复 close chan 的问题
	for _, ch := range chans {
		close(ch)
	}
	return nil
}

type Msg struct {
	//Topic string
	Content string
}

//type Listener func(msg Msg)

type BrokerV2 struct {
	mutex sync.RWMutex
	consumers []func(msg Msg)
	//listeners []Listener
	// map[string][]func(msg Msg)
}

func (b *BrokerV2) Send(m Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, c := range b.consumers {
		c(m)
	}
	return nil
}

func (b *BrokerV2) Subscribe(cb func(s Msg)) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.consumers = append(b.consumers, cb)
	return nil
}