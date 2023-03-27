package synccollections

import (
	"fmt"
	"testing"
)

func TestMakeSyncSlice(t *testing.T) {
	a := MakeSlice[int](0, 20)

	a.Append(1)
	a.Append(2)
	a.Append(3)
	a.Append(4)
	a.Append(5)
	a.Append(6)

	fmt.Println(a.Copy())

	a = a.Append(122)

	a.Replace(nil)

	fmt.Println(a.Copy())

	a.Append(1)
	a.Append(2)
	fmt.Println(a.Copy())

	fmt.Println(a.Get(1))
	a.Set(1, 111)
	fmt.Println(a.Get(1))
	a.Append(3)
	a.Append(4)
	//fmt.Println(a.CopySlice(0, 10)) // panic

	fmt.Println(a.CopySlice(1, 4))

	a.Range(func(i int, v int) bool {
		fmt.Println(i, v)
		if i == 2 { // i=2就退出
			return false
		}
		return true
	})

	a.Replace(nil)

	for i := 0; i < 10; i++ {
		a.Append(i)
	}
	a.DeleteAllIf(func(i int) bool {
		return i%2 == 0
	})
	fmt.Println(a.Copy())

	a.DeleteAllIf(func(i int) bool {
		return i > 3
	})
	fmt.Println(a.Copy())
}
