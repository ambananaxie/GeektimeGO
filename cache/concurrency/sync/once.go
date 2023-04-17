package sync

import "sync"

type MyBiz struct {
	once sync.Once
}

func (m *MyBiz) Init() {
	m.once.Do(func() {

	})
}

//func (m MyBiz) InitV1() {
//	m.once.Do(func() {
//
//	})
//}

type MyBizV1 struct {
	once *sync.Once
}

func (m MyBizV1) Init() {
	m.once.Do(func() {

	})
}


type MyBusiness interface {
	DoSomething()
}

type singleton struct {

}

func (s singleton) DoSomething() {
	panic("implement me")
}

var s *singleton
var singletonOnce sync.Once

// 懒加载
func GetSingleton() MyBusiness {
	singletonOnce.Do(func() {
		s = &singleton{}
	})
	return s
}

// 饥饿
func init() {
	// 用包初始化函数取代 once
	s = &singleton{}
}