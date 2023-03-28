package synccollections

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	var a = new(RWMutexMap[int, int])
	a.Store(1, 1)

	a.Range(func(i, j int) bool {
		fmt.Println(i, j)
		return true
	})

	swapped := a.CompareAndSwap(1, 2, 2)
	fmt.Println(swapped)
	a.Range(func(i, j int) bool {
		fmt.Println(i, j)
		return true
	})

}

func TestMap(t *testing.T) {
	var m Map[int, string]

	//actual, loaded := m.LoadOrStore(1, "10")
	//fmt.Println(actual)
	//fmt.Println(loaded)
	//
	fmt.Println(m.Swap(1, "200"))
}
