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
}
