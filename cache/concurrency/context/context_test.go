package context

import (
	"context"
	"testing"
	"time"
)

type mykey struct {}
type mykeyv2 int
type mykeyv3 string

func TestContext(t *testing.T) {
	// 一般是链路起点，或者调用的起点
	ctx := context.Background()
	// 在你不确定 context 该用啥的时候，用 TODO()
	//ctx := context.TODO()

	ctx = context.WithValue(ctx, mykey{}, "my-value")
	//ctx = context.WithValue(ctx, "my-key", "my-value")
	val := ctx.Value(mykey{}).(string)
	t.Log(val)
	newVal := ctx.Value("不存在的key")
	val, ok := newVal.(string)
	if !ok {
		t.Log("类型不对")
		return
	}
	t.Log(val)
}

func TestContext_WithCancel(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	// 用完 ctx 再去调用
	//defer cancel()

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	// 用 ctx
	<- ctx.Done()
	t.Log("hello, cancel: ", ctx.Err())
}

func TestContext_WithDeadline(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second * 3))
	deadline, _ := ctx.Deadline()
	t.Log("deadline: ", deadline)
	defer cancel()
	<- ctx.Done()
	t.Log("hello, deadline: ", ctx.Err())
}

func TestContext_WithTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second * 3)
	deadline, _ := ctx.Deadline()
	t.Log("deadline: ", deadline)
	defer cancel()
	<- ctx.Done()
	t.Log("hello, timeout: ", ctx.Err())
}

func TestContext_Parent(t *testing.T) {
	ctx := context.Background()
	parent := context.WithValue(ctx, "my-key", "my value")
	child := context.WithValue(parent, "my-key", "my new value")

	t.Log("parent my-key: ", parent.Value("my-key"))
	t.Log("child my-key: ", child.Value("my-key"))

	child2, cancel := context.WithTimeout(parent, time.Second)
	defer cancel()
	t.Log("child2 my-key:", child2.Value("my-key"))

	child3 := context.WithValue(parent, "new-key", "child3 value")
	t.Log("parent new-key: ", parent.Value("new-key"))
	t.Log("child3 new-key: ", child3.Value("new-key"))

	// 逼不得已使用
	parent1 := context.WithValue(ctx, "map", map[string]string{})
	child4, cancel := context.WithTimeout(parent1, time.Second)
	defer cancel()
	m := child4.Value("map").(map[string]string)
	m["key1"]= "value1"
	nm := parent1.Value("map").(map[string]string)
	t.Log("parent1 key1: ", nm["key1"])
}

func TestContext_ParentTimeout(t *testing.T) {
	ctx := context.Background()
	parent, cancel := context.WithTimeout(ctx, time.Second)

	child, cancel1 := context.WithTimeout(parent, time.Second * 3)
	defer cancel1()

	go func() {
		<- child.Done()
		// 覆盖不成功，输出这句话
		t.Log("儿子收到了结束信号")
	}()

	time.Sleep(time.Second * 2)
	cancel()

	parent, cancel =  context.WithTimeout(ctx, time.Second * 3)
	defer cancel()
	child2, cancel2 := context.WithTimeout(parent, time.Second)
	defer cancel2()

	go func() {
		<- child2.Done()
		// 覆盖成功，输出这句话
		t.Log("2儿子收到了结束信号")
	}()
	time.Sleep(time.Second * 2)
}

func TestTimeoutExample(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 3)
	defer cancel()
	bsChan := make(chan struct{})
	go func() {
		slowBusiness()
		// 业务结束，发送信号
		bsChan <- struct{}{}
	}()

	select {
	case <- ctx.Done():
		t.Log("超时了")
	case <- bsChan:
		t.Log("业务正常结束")
	}
}

func slowBusiness() {
	time.Sleep(time.Second * 2)
}