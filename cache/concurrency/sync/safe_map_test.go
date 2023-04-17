package sync

import (
	"testing"
	"time"
)

func TestSafeMap_LoadOrStore(t *testing.T) {
	s := &SafeMap[string, string]{
		data: make(map[string]string),
	}

	go func() {
		val, ok := s.LoadOrStore("key1", "value1")
		t.Log("goroutine1 value: ", val, ok)
	}()

	go func() {
		val, ok := s.LoadOrStore("key1", "value2")
		t.Log("goroutine2 value: ", val, ok)
	}()
	time.Sleep(time.Second)
}
