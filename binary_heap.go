package priorityqueue

import (
	"sync"
)

// Item represents a binary heap item
type Item interface {
	// ID represents a unique ID of the item
	ID() string
	// Priority returns the Item's priority to sort
	Priority() int64
	// GroupID represents the Item's group, used to delete all Items with the same GroupID
	GroupID() string
}

type BinHeap[T Item] struct {
	items []T
	// exists used as a shadow structure to check if the item exists in the BinHeap
	exists map[string]struct{}
	st     *stack
	// find a way to use a pointer to the raw data
	maxLen uint64
	cond   sync.Cond
}

func NewBinHeap[T Item](maxLen uint64) *BinHeap[T] {
	return &BinHeap[T]{
		items:  make([]T, 0, 1000),
		exists: make(map[string]struct{}, 1000),
		st:     newStack(),
		maxLen: maxLen,
		cond:   sync.Cond{L: &sync.Mutex{}},
	}
}

func (bh *BinHeap[T]) fixUp() {
	k := uint64(len(bh.items)) - 1
	p := (k - 1) >> 1 // k-1 / 2

	for k > 0 {
		cur, par := (bh.items)[k], (bh.items)[p]

		if cur.Priority() < par.Priority() {
			bh.swap(k, p)
			k = p
			p = (k - 1) >> 1
		} else {
			return
		}
	}
}

func (bh *BinHeap[T]) swap(i, j uint64) {
	(bh.items)[i], (bh.items)[j] = (bh.items)[j], (bh.items)[i]
}

func (bh *BinHeap[T]) fixDown(curr, end int) {
	cOneIdx := (curr << 1) + 1
	for cOneIdx <= end {
		cTwoIdx := -1
		if (curr<<1)+2 <= end {
			cTwoIdx = (curr << 1) + 2
		}

		idxToSwap := cOneIdx
		if cTwoIdx > -1 && (bh.items)[cTwoIdx].Priority() < (bh.items)[cOneIdx].Priority() {
			idxToSwap = cTwoIdx
		}
		if (bh.items)[idxToSwap].Priority() < (bh.items)[curr].Priority() {
			bh.swap(uint64(curr), uint64(idxToSwap)) //nolint:gosec
			curr = idxToSwap
			cOneIdx = (curr << 1) + 1
		} else {
			return
		}
	}
}

func (bh *BinHeap[T]) Exists(id string) bool {
	bh.cond.L.Lock()
	defer bh.cond.L.Unlock()

	_, ok := bh.exists[id]
	return ok
}

// Remove removes all elements with the provided ID and returns the slice with them
func (bh *BinHeap[T]) Remove(groupID string) []T {
	bh.cond.L.Lock()
	defer bh.cond.L.Unlock()

	out := make([]T, 0, 10)

	for i := range bh.items {
		if bh.items[i].GroupID() == groupID {
			// delete element
			delete(bh.exists, bh.items[i].ID())
			out = append(out, bh.items[i])
			bh.st.Add(i)
		}
	}

	ids := bh.st.Indices()
	adjustment := 0
	for i := range ids {
		start := ids[i][0] - adjustment
		end := ids[i][1] - adjustment

		bh.items = append(bh.items[:start], bh.items[end+1:]...)
		adjustment += end - start + 1
	}

	// re-heapify after compaction (Floyd's algorithm)
	n := len(bh.items)
	for i := n/2 - 1; i >= 0; i-- {
		bh.fixDown(i, n-1)
	}

	bh.st.clear()
	bh.cond.Broadcast()

	return out
}

// PeekPriority returns the highest priority
func (bh *BinHeap[T]) PeekPriority() int64 {
	bh.cond.L.Lock()
	defer bh.cond.L.Unlock()

	if len(bh.items) > 0 {
		return bh.items[0].Priority()
	}

	return -1
}

func (bh *BinHeap[T]) Len() uint64 {
	bh.cond.L.Lock()
	defer bh.cond.L.Unlock()
	return uint64(len(bh.items))
}

func (bh *BinHeap[T]) Insert(item T) {
	bh.cond.L.Lock()

	for uint64(len(bh.items)) >= bh.maxLen {
		bh.cond.Wait()
	}

	bh.items = append(bh.items, item)

	// fix binary heap up
	bh.fixUp()

	// add item
	bh.exists[item.ID()] = struct{}{}

	bh.cond.L.Unlock()

	// signal the goroutine on wait
	bh.cond.Signal()
}

func (bh *BinHeap[T]) ExtractMin() T {
	bh.cond.L.Lock()

	// if len == 0, wait for the signal
	for len(bh.items) == 0 {
		bh.cond.Wait()
	}

	n := uint64(len(bh.items))
	bh.swap(0, n-1)

	item := bh.items[n-1]
	bh.items = bh.items[:n-1]
	bh.fixDown(0, int(n)-2) //nolint:gosec

	// remove item
	delete(bh.exists, item.ID())

	bh.cond.L.Unlock()

	// signal blocked producers waiting for space
	bh.cond.Signal()
	return item
}
