package priorityqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonotonicStackInit(t *testing.T) {
	c := newStack()
	c.Add(1)
	c.Add(2)
	c.Add(3)
	c.Add(5)
	c.Add(6)
	c.Add(10)
	c.Add(14)

	assert.Equal(t, [][2]int{{1, 3}, {5, 6}, {10, 10}, {14, 14}}, c.Indices())
}

func TestMonotonicStackLastStartEl(t *testing.T) {
	c := newStack()
	c.Add(1)
	c.Add(3)
	c.Add(5)
	c.Add(7)
	c.Add(9)
	c.Add(11)
	c.Add(13)

	assert.Equal(t, [][2]int{{1, 1}, {3, 3}, {5, 5}, {7, 7}, {9, 9}, {11, 11}, {13, 13}}, c.Indices())
}
