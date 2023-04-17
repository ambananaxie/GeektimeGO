package cache

//func TestReadThroughCacheV1_Get(t *testing.T) {
//	var c1 ReadThroughCache = ReadThroughCache{
//		LoadFunc: func(ctx context.Context, key string) (any, error) {
//			if strings.HasPrefix(key, "user_") {
//				// 加载 user
//			} else if strings.HasPrefix(key, "order_") {
//				// 加载 order
//			} else {
//				return nil, errors.New("不知道怎么加载")
//			}
//		},
//	}
//	var c2 ReadThroughCacheV1[User]
//
//	val, _ := c1.Get(context.Background(), "user_1")
//	u := val.(User)
//	val, _ = c1.Get(context.Background(), "order_1")
//	o := val.(Order)
//	t.Log(o)
//
//	u,_ = c2.Get(context.Background(), "user_1")
//	//o, _ = c2.Get(context.Background(), "order_1")
//	t.Log(u.Name)
//}
