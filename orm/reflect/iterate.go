package reflect

import "reflect"

func IterateArrayOrSlice(entity any) ([]any, error) {
	val := reflect.ValueOf(entity)
	res := make([]any, 0, val.Len())
	for i := 0; i< val.Len(); i ++ {
		ele := val.Index(i)
		res = append(res, ele.Interface())
	}
	return res, nil
}

// 返回值是 keys, values, error
func IterateMap(entity any) ([]any, []any, error) {
	val := reflect.ValueOf(entity)
	resKeys := make([]any, 0, val.Len())
	resValues := make([]any, 0, val.Len())

	itr := val.MapRange()
	for itr.Next() {
		resKeys = append(resKeys, itr.Key().Interface())
		resValues = append(resValues, itr.Value().Interface())
	}

	// keys := val.MapKeys()
	// for _, key := range keys {
	// 	v := val.MapIndex(key)
	// 	resKeys = append(resKeys, key.Interface())
	// 	resValues = append(resValues, v.Interface())
	// }
	return resKeys, resValues, nil
}