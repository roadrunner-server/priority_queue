/*
binary heap (min-heap) algorithm used as a core for the priority queue
*/

package priorityqueue

import (
	"context"
	"sync"
	"sync/atomic"
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
	len     uint64
	maxLen  uint64
	cond    sync.Cond
	minCh   chan T
	stopCf  context.CancelFunc
	stopCtx context.Context
}

func NewBinHeap[T Item](maxLen uint64) *BinHeap[T] {
	bh := &BinHeap[T]{
		items:  make([]T, 0, 1000),
		exists: make(map[string]struct{}, 1000),
		st:     newStack(),
		len:    0,
		maxLen: maxLen,
		cond:   sync.Cond{L: &sync.Mutex{}},
		minCh:  make(chan T),
	}

	bh.stopCtx, bh.stopCf = context.WithCancel(context.Background())

	go bh.extractMin()

	return bh
}

func (bh *BinHeap[T]) fixUp() {
	k := bh.len - 1
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

	if _, ok := bh.exists[id]; ok {
		return true
	}

	return false
}

func (bh *BinHeap[T]) Stop() {
	bh.stopCf()
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

	atomic.StoreUint64(&bh.len, uint64(len(bh.items)))
	bh.st.clear()

	return out
}

// PeekPriority returns the highest priority
func (bh *BinHeap[T]) PeekPriority() int64 {
	bh.cond.L.Lock()
	defer bh.cond.L.Unlock()

	if bh.Len() > 0 {
		return bh.items[0].Priority()
	}

	return -1
}

func (bh *BinHeap[T]) Len() uint64 {
	return atomic.LoadUint64(&bh.len)
}

func (bh *BinHeap[T]) Insert(item T) {
	bh.cond.L.Lock()

	// check the binary heap len before insertion
	if bh.Len() > bh.maxLen {
		// unlock the mutex to proceed to get-max
		bh.cond.L.Unlock()

		// signal waiting goroutines
		for bh.Len() > 0 {
			// signal waiting goroutines
			bh.cond.Signal()
		}
		// lock mutex to proceed inserting into the empty slice
		bh.cond.L.Lock()
	}

	bh.items = append(bh.items, item)

	// add len to the slice
	atomic.AddUint64(&bh.len, 1)

	// fix binary heap up
	bh.fixUp()

	// add item
	bh.exists[item.ID()] = struct{}{}

	bh.cond.L.Unlock()

	// signal the goroutine on wait
	bh.cond.Signal()
}

// ExtractMinCh returns a channel to extract the minimum item
// We need this function to be able to use select statement and avoid blocking
// because ExtractMin is a blocking operation (on bh.cond.Wait())
func (bh *BinHeap[T]) ExtractMinCh() <-chan T {
	return bh.minCh
}

func (bh *BinHeap[T]) extractMin() {
	for {
		bh.cond.L.Lock()

		// if len == 0, wait for the signal
		for bh.Len() == 0 {
			bh.cond.Wait()
		}

		bh.swap(0, bh.len-1)

		item := (bh.items)[int(bh.len)-1]        //nolint:gosec
		bh.items = (bh).items[0 : int(bh.len)-1] //nolint:gosec
		bh.fixDown(0, int(bh.len-2))             //nolint:gosec

		// reduce len
		atomic.AddUint64(&bh.len, ^uint64(0))

		// remove item
		delete(bh.exists, item.ID())
		bh.cond.L.Unlock()
		// send item to the channel
		select {
		case bh.minCh <- item:
		case <-bh.stopCtx.Done():
			close(bh.minCh)
			return
		}
	}
}
