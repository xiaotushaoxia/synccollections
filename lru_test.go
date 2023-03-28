package synccollections

import (
	"fmt"
	"testing"
)

func TestNewLRU(t *testing.T) {
	a := MakeLRU[int, int](3)

	a.Set(1, 1)
	a.Set(2, 2)

	a.Set(3, 3)
	a.Set(1, 4)
	a.Set(2, 5)
	a.Set(3, 5)
	a.Get(1)
	a.Set(4, 5)
	a.Set(2, 5)

	fmt.Println(a.Get(1))
	fmt.Println(a.Get(2))
	fmt.Println(a.Get(3))
	fmt.Println(a.Get(4))

	fmt.Println(a.GetSize())
}

func TestNewLRU2(t *testing.T) {
	var a LRU[int, int]

	for i := 0; i < 103; i++ {
		a.Set(i, i)
	}

	a.SetCap(40)

	fmt.Println(a.GetSize())
}

func TestNewLRU3(t *testing.T) {
	var a LRU[int, int]

	for i := 0; i < 103; i++ {
		a.Set(i, i)
	}
	a.Get(3)

	a.Range(func(key int, value int) (shouldContinue bool) {
		fmt.Println(key, value)
		return true
	})

	var ts = a.Keys()
	fmt.Println(ts)

	a.Peek(19)

}
