package ringbuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRing_WalkFirstN(t *testing.T) {
	ring := New[int](5)
	for i := 0; i < 7; i++ {
		ring.PushFront(i)
	}

	actual := make([]int, 0)
	ring.WalkFirstN(7, func(v int) {
		actual = append(actual, v)
	})
	assert.Equal(t, []int{6, 5, 4, 3, 2, 6, 5}, actual)
}

func TestRing_SetN(t *testing.T) {
	ring := New[int](5)
	for i := 0; i < 7; i++ {
		ring.PushFront(i)
	}

	assert.Equal(t, 6, ring.GetN(0))
	assert.Equal(t, 2, ring.GetN(-1))
}
