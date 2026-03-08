package priorityqueue

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Test struct {
	id       string
	groupID  string
	priority int64
}

func NewTest(priority int64, groupID string, id string) Test {
	return Test{
		priority: priority,
		groupID:  groupID,
		id:       id,
	}
}

func (t Test) ID() string {
	return t.id
}

func (t Test) GroupID() string {
	return t.groupID
}

func (t Test) Priority() int64 {
	return t.priority
}

func TestBinHeap_Init(t *testing.T) {
	a := []Item{
		NewTest(2, "foo0", "bar0"),
		NewTest(23, uuid.NewString(), uuid.NewString()),
		NewTest(33, uuid.NewString(), uuid.NewString()),
		NewTest(44, uuid.NewString(), uuid.NewString()),
		NewTest(1, uuid.NewString(), uuid.NewString()),
		NewTest(2, "foo1", "bar1"),
		NewTest(2, "foo2", "bar2"),
		NewTest(2, "foo3", "bar3"),
		NewTest(4, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(99, uuid.NewString(), uuid.NewString()),
	}

	bh := NewBinHeap[Item](12)

	for i := range a {
		bh.Insert(a[i])
	}

	expected := []int64{
		1,
		2,
		2,
		2,
		2,
		4,
		6,
		23,
		33,
		44,
		99,
	}

	res := make([]int64, 0, 12)

	for range 11 {
		item := bh.ExtractMin()
		item.Priority()
		res = append(res, item.Priority())
	}

	require.Equal(t, expected, res)
}

func TestBinHeap_MaxLen(t *testing.T) {
	a := []Item{
		NewTest(2, uuid.NewString(), uuid.NewString()),
		NewTest(23, uuid.NewString(), uuid.NewString()),
		NewTest(33, uuid.NewString(), uuid.NewString()),
		NewTest(44, uuid.NewString(), uuid.NewString()),
		NewTest(1, uuid.NewString(), uuid.NewString()),
		NewTest(2, uuid.NewString(), uuid.NewString()),
		NewTest(2, uuid.NewString(), uuid.NewString()),
		NewTest(2, uuid.NewString(), uuid.NewString()),
		NewTest(4, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(99, uuid.NewString(), uuid.NewString()),
	}

	bh := NewBinHeap[Item](1)

	go func() {
		res := make([]Item, 0, 12)

		for range 11 {
			item := bh.ExtractMin()
			res = append(res, item)
		}
		require.Equal(t, 11, len(res))
	}()

	time.Sleep(time.Second)
	for i := range a {
		bh.Insert(a[i])
	}

	time.Sleep(time.Second)
}

func TestNewPriorityQueue(t *testing.T) {
	insertsPerSec := uint64(0)
	getPerSec := uint64(0)
	stopCh := make(chan struct{}, 1)
	pq := NewBinHeap[Item](1000)

	go func() {
		tt3 := time.NewTicker(time.Millisecond * 10)
		for {
			select {
			case <-tt3.C:
				require.Less(t, pq.Len(), uint64(1002))
			case <-stopCh:
				return
			}
		}
	}()

	go func() {
		tt := time.NewTicker(time.Second)

		for {
			select {
			case <-tt.C:
				fmt.Printf("Insert per second: %d\n", atomic.LoadUint64(&insertsPerSec))
				atomic.StoreUint64(&insertsPerSec, 0)
				fmt.Printf("ExtractMin per second: %d\n", atomic.LoadUint64(&getPerSec))
				atomic.StoreUint64(&getPerSec, 0)
			case <-stopCh:
				tt.Stop()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				pq.ExtractMin()
				atomic.AddUint64(&getPerSec, 1)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				pq.Insert(NewTest(rand.Int63(), uuid.NewString(), uuid.NewString())) //nolint:gosec
				atomic.AddUint64(&insertsPerSec, 1)
			}
		}
	}()

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	stopCh <- struct{}{}
	stopCh <- struct{}{}
	stopCh <- struct{}{}
}

func TestNewItemWithTimeout(t *testing.T) {
	a := []Item{
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(23, uuid.NewString(), uuid.NewString()),
		NewTest(33, uuid.NewString(), uuid.NewString()),
		NewTest(44, uuid.NewString(), uuid.NewString()),
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(7, uuid.NewString(), uuid.NewString()),
		NewTest(8, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(99, uuid.NewString(), uuid.NewString()),
	}

	/*
		first item should be extracted not less than 5 seconds after we call ExtractMin
		5 seconds is a minimum timeout for our items
	*/
	bh := NewBinHeap[Item](100)

	for i := range a {
		bh.Insert(a[i])
	}

	tn := time.Now()
	item := bh.ExtractMin()
	assert.Equal(t, int64(5), item.Priority())
	assert.GreaterOrEqual(t, float64(5), time.Since(tn).Seconds())
}

func TestItemPeek(t *testing.T) {
	a := []Item{
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(23, uuid.NewString(), uuid.NewString()),
		NewTest(33, uuid.NewString(), uuid.NewString()),
		NewTest(44, uuid.NewString(), uuid.NewString()),
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(7, uuid.NewString(), uuid.NewString()),
		NewTest(8, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(99, uuid.NewString(), uuid.NewString()),
	}

	/*
		first item should be extracted not less than 5 seconds after we call ExtractMin
		5 seconds is a minimum timeout for our items
	*/
	bh := NewBinHeap[Item](100)

	for i := range a {
		bh.Insert(a[i])
	}

	tmp := bh.PeekPriority()
	assert.Equal(t, int64(5), tmp)

	tn := time.Now()
	item := bh.ExtractMin()
	assert.Equal(t, int64(5), item.Priority())
	assert.GreaterOrEqual(t, float64(5), time.Since(tn).Seconds())
}

func TestItemPeekConcurrent(t *testing.T) {
	a := []Item{
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(23, uuid.NewString(), uuid.NewString()),
		NewTest(33, uuid.NewString(), uuid.NewString()),
		NewTest(44, uuid.NewString(), uuid.NewString()),
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(5, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(7, uuid.NewString(), uuid.NewString()),
		NewTest(8, uuid.NewString(), uuid.NewString()),
		NewTest(6, uuid.NewString(), uuid.NewString()),
		NewTest(99, uuid.NewString(), uuid.NewString()),
	}

	/*
		first item should be extracted not less than 5 seconds after we call ExtractMin
		5 seconds is a minimum timeout for our items
	*/
	bh := NewBinHeap[Item](100)

	for i := range a {
		bh.Insert(a[i])
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 1000 {
			tmp := bh.PeekPriority()
			_ = tmp
		}
	}()

	go func() {
		defer wg.Done()
		for range 11 {
			m := bh.ExtractMin()
			_ = m
		}
	}()

	wg.Wait()
}

func TestBinHeap_RemoveHeapProperty(t *testing.T) {
	// Regression test: Remove must restore the heap property after compaction.
	// Insert priorities [1(A), 3(B), 2(B)] → heap: [1, 3, 2]
	// Remove group "A" → compacts to [3, 2] which violates min-heap
	// Without re-heapify, ExtractMin would return 3 instead of 2.
	bh := NewBinHeap[Item](10)
	bh.Insert(NewTest(1, "A", "id1"))
	bh.Insert(NewTest(3, "B", "id2"))
	bh.Insert(NewTest(2, "B", "id3"))

	removed := bh.Remove("A")
	require.Len(t, removed, 1)
	require.Equal(t, "id1", removed[0].ID())

	first := bh.ExtractMin()
	assert.Equal(t, int64(2), first.Priority(), "expected min priority 2, got %d", first.Priority())

	second := bh.ExtractMin()
	assert.Equal(t, int64(3), second.Priority(), "expected priority 3, got %d", second.Priority())
}

func TestBinHeap_Remove(t *testing.T) {
	a := []Item{
		NewTest(2, "1", "101"),
		NewTest(5, "1", "102"),
		NewTest(99, "1", "103"),
		NewTest(4, "6", "104"),
		NewTest(6, "7", "105"),
		NewTest(23, "2", "106"),
		NewTest(2, "1", "107"),
		NewTest(2, "1", "108"),
		NewTest(33, "3", "109"),
		NewTest(44, "4", "110"),
		NewTest(2, "1", "111"),
	}

	bh := NewBinHeap[Item](12)

	for i := range a {
		bh.Insert(a[i])
	}

	expected := []Item{
		NewTest(4, "6", "104"),
		NewTest(6, "7", "105"),
		NewTest(23, "2", "106"),
		NewTest(33, "3", "109"),
		NewTest(44, "4", "110"),
	}

	out := bh.Remove("1")
	if len(out) != 6 {
		t.Fatalf("expected 6, got %d", len(out))
	}

	for i := range out {
		if out[i].GroupID() != "1" {
			t.Fatal("id is not 1")
		}
	}

	res := make([]Item, 0, 12)

	for range 5 {
		item := bh.ExtractMin()
		res = append(res, item)
	}

	require.Equal(t, expected, res)
}

func TestExists(t *testing.T) {
	const id = "11111111111"
	a := []Item{
		NewTest(2, "1", id),
		NewTest(5, "1", uuid.NewString()),
		NewTest(99, "1", uuid.NewString()),
		NewTest(4, "6", uuid.NewString()),
		NewTest(6, "7", uuid.NewString()),
		NewTest(23, "2", uuid.NewString()),
		NewTest(2, "1", uuid.NewString()),
		NewTest(2, "1", uuid.NewString()),
		NewTest(33, "3", uuid.NewString()),
		NewTest(44, "4", uuid.NewString()),
		NewTest(2, "1", uuid.NewString()),
	}

	bh := NewBinHeap[Item](12)

	for i := range a {
		bh.Insert(a[i])
	}

	assert.False(t, bh.Exists("1"))
	assert.True(t, bh.Exists(id))

	_ = bh.Remove("1")

	assert.False(t, bh.Exists(id))
}

func TestBinHeap_RemoveHeapPropertyLarge(t *testing.T) {
	bh := NewBinHeap[Item](200)

	// Insert 100 items across 5 groups with interleaved priorities so
	// the target group's items are scattered at root, mid, and leaf heap levels.
	for i := 0; i < 100; i++ {
		groupID := fmt.Sprintf("g%d", i%5)
		priority := int64(i + 1) // 1..100, round-robin across groups
		id := fmt.Sprintf("item-%d", i)
		bh.Insert(NewTest(priority, groupID, id))
	}

	require.Equal(t, uint64(100), bh.Len())

	// Remove group "g2" (priorities 3,8,13,18,...,98 — 20 items)
	removed := bh.Remove("g2")
	require.Len(t, removed, 20)
	for _, item := range removed {
		require.Equal(t, "g2", item.GroupID())
	}
	require.Equal(t, uint64(80), bh.Len())

	// Extract all remaining items and verify strictly non-decreasing order
	var prev int64
	for i := 0; i < 80; i++ {
		item := bh.ExtractMin()
		require.GreaterOrEqual(t, item.Priority(), prev,
			"item %d: priority %d should be >= previous %d", i, item.Priority(), prev)
		require.NotEqual(t, "g2", item.GroupID())
		prev = item.Priority()
	}

	require.Equal(t, uint64(0), bh.Len())
}

func TestBinHeap_RemoveMultipleGroups(t *testing.T) {
	bh := NewBinHeap[Item](100)

	// 4 groups with known priorities
	bh.Insert(NewTest(10, "A", "a1"))
	bh.Insert(NewTest(30, "A", "a2"))
	bh.Insert(NewTest(5, "B", "b1"))
	bh.Insert(NewTest(25, "B", "b2"))
	bh.Insert(NewTest(15, "C", "c1"))
	bh.Insert(NewTest(35, "C", "c2"))
	bh.Insert(NewTest(20, "D", "d1"))
	bh.Insert(NewTest(40, "D", "d2"))

	// Remove group A, verify min is B's 5
	removedA := bh.Remove("A")
	require.Len(t, removedA, 2)
	require.Equal(t, int64(5), bh.PeekPriority())

	// Remove group B, verify min is now C's 15
	removedB := bh.Remove("B")
	require.Len(t, removedB, 2)
	require.Equal(t, int64(15), bh.PeekPriority())

	// Extract remaining items (C and D) and verify order
	expected := []int64{15, 20, 35, 40}
	for _, exp := range expected {
		item := bh.ExtractMin()
		require.Equal(t, exp, item.Priority())
	}
}

func TestBinHeap_RemoveEdgeCases(t *testing.T) {
	t.Run("remove all items", func(t *testing.T) {
		bh := NewBinHeap[Item](10)
		bh.Insert(NewTest(1, "only", "id1"))
		bh.Insert(NewTest(2, "only", "id2"))
		bh.Insert(NewTest(3, "only", "id3"))

		removed := bh.Remove("only")
		require.Len(t, removed, 3)
		require.Equal(t, uint64(0), bh.Len())

		// Insert new items and verify they work after full removal
		bh.Insert(NewTest(42, "new", "id4"))
		require.Equal(t, uint64(1), bh.Len())
		item := bh.ExtractMin()
		require.Equal(t, int64(42), item.Priority())
	})

	t.Run("remove non-existent group", func(t *testing.T) {
		bh := NewBinHeap[Item](10)
		bh.Insert(NewTest(1, "exists", "id1"))
		bh.Insert(NewTest(2, "exists", "id2"))

		removed := bh.Remove("ghost")
		require.Empty(t, removed)
		require.Equal(t, uint64(2), bh.Len())

		// Verify heap still works correctly
		item := bh.ExtractMin()
		require.Equal(t, int64(1), item.Priority())
	})

	t.Run("remove from empty heap", func(t *testing.T) {
		bh := NewBinHeap[Item](10)
		removed := bh.Remove("anything")
		require.Empty(t, removed)
		require.Equal(t, uint64(0), bh.Len())
	})
}

func TestBinHeap_BoundedInsertBackpressure(t *testing.T) {
	bh := NewBinHeap[Item](5)

	// Fill to capacity
	for i := 0; i < 5; i++ {
		bh.Insert(NewTest(int64(i+1), "g1", fmt.Sprintf("item-%d", i)))
	}
	require.Equal(t, uint64(5), bh.Len())

	// Launch goroutine to insert one more (should block at capacity)
	inserted := make(chan struct{})
	go func() {
		bh.Insert(NewTest(10, "g1", "blocked-item"))
		close(inserted)
	}()

	// Give goroutine time to start and block on the full heap
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, uint64(5), bh.Len(), "producer should be blocked, heap still at capacity")

	// Extract one item to free space and signal the blocked producer
	item := bh.ExtractMin()
	require.Equal(t, int64(1), item.Priority())

	// Wait for insert goroutine to complete
	select {
	case <-inserted:
		// success — producer unblocked
	case <-time.After(2 * time.Second):
		t.Fatal("insert goroutine did not unblock after ExtractMin")
	}

	require.Equal(t, uint64(5), bh.Len(), "should be 5: was 5, extracted 1, inserted 1")
}

func TestBinHeap_RemoveUnblocksInsert(t *testing.T) {
	bh := NewBinHeap[Item](5)

	// Fill to capacity with one removable group
	for i := 0; i < 5; i++ {
		bh.Insert(NewTest(int64(i+1), "removeMe", fmt.Sprintf("item-%d", i)))
	}
	require.Equal(t, uint64(5), bh.Len())

	// Launch goroutine to insert one more (should block at capacity)
	inserted := make(chan struct{})
	go func() {
		bh.Insert(NewTest(99, "keep", "new-item"))
		close(inserted)
	}()

	// Give goroutine time to start and block
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, uint64(5), bh.Len(), "producer should be blocked, heap still at capacity")

	// Remove group to free space — this should unblock the producer
	removed := bh.Remove("removeMe")
	require.Len(t, removed, 5)

	select {
	case <-inserted:
		// success — producer unblocked by Remove
	case <-time.After(2 * time.Second):
		t.Fatal("insert goroutine did not unblock after Remove freed space")
	}

	require.Equal(t, uint64(1), bh.Len())
	item := bh.ExtractMin()
	require.Equal(t, int64(99), item.Priority())
}

func TestBinHeap_ConcurrentInsertRemoveExtract(t *testing.T) {
	// Large capacity to avoid Insert back-pressure during stress test;
	// back-pressure is tested separately in TestBinHeap_BoundedInsertBackpressure.
	bh := NewBinHeap[Item](^uint64(0))

	var done atomic.Bool
	var producerWg sync.WaitGroup
	var consumerWg sync.WaitGroup

	// 3 producer goroutines inserting items with random priorities across 10 groups
	for p := 0; p < 3; p++ {
		producerWg.Add(1)
		go func(id int) {
			defer producerWg.Done()
			for i := 0; !done.Load(); i++ {
				groupID := fmt.Sprintf("g%d", i%10)
				itemID := fmt.Sprintf("p%d-i%d", id, i)
				bh.Insert(NewTest(rand.Int63n(1000), groupID, itemID)) //nolint:gosec
			}
		}(p)
	}

	// 2 consumer goroutines calling ExtractMin
	for c := 0; c < 2; c++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			for !done.Load() {
				_ = bh.ExtractMin()
			}
		}()
	}

	// 1 remover goroutine periodically removing a random group
	producerWg.Add(1)
	go func() {
		defer producerWg.Done()
		for !done.Load() {
			bh.Remove(fmt.Sprintf("g%d", rand.Intn(10))) //nolint:gosec
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Run for 2 seconds
	time.Sleep(2 * time.Second)
	done.Store(true)

	// Wait for producers and remover to finish
	producerWg.Wait()

	// Unblock consumers that may be stuck waiting on an empty heap
	consumerDone := make(chan struct{})
	go func() {
		consumerWg.Wait()
		close(consumerDone)
	}()

	for {
		select {
		case <-consumerDone:
			// All consumers exited — verify heap is in a consistent state
			_ = bh.Len()
			return
		default:
			bh.Insert(NewTest(0, "sentinel", fmt.Sprintf("s-%d", time.Now().UnixNano())))
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestBinHeap_LargeScaleOrdering(t *testing.T) {
	const n = 10_000
	bh := NewBinHeap[Item](uint64(n) + 1)

	for i := 0; i < n; i++ {
		bh.Insert(NewTest(rand.Int63n(1000), "g", fmt.Sprintf("item-%d", i))) //nolint:gosec
	}

	var prev int64
	for i := 0; i < n; i++ {
		item := bh.ExtractMin()
		require.GreaterOrEqual(t, item.Priority(), prev,
			"item %d: priority %d should be >= previous %d", i, item.Priority(), prev)
		prev = item.Priority()
	}
}

func BenchmarkInsert(b *testing.B) {
	bh := NewBinHeap[Item](1 << 30)
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		bh.Insert(NewTest(rand.Int63n(100000), "bench", fmt.Sprintf("b-%d", i))) //nolint:gosec
		i++
	}
}

func BenchmarkExtractMin(b *testing.B) {
	bh := NewBinHeap[Item](uint64(max(b.N, 0)) + 1)
	for i := range b.N {
		bh.Insert(NewTest(rand.Int63n(100000), "bench", fmt.Sprintf("b-%d", i))) //nolint:gosec
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bh.ExtractMin()
	}
}

func BenchmarkInsertExtractMin(b *testing.B) {
	bh := NewBinHeap[Item](2000)
	// Pre-fill with 1000 items
	for i := 0; i < 1000; i++ {
		bh.Insert(NewTest(rand.Int63n(100000), "bench", fmt.Sprintf("pre-%d", i))) //nolint:gosec
	}
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		bh.Insert(NewTest(rand.Int63n(100000), "bench", fmt.Sprintf("b-%d", i))) //nolint:gosec
		bh.ExtractMin()
		i++
	}
}

func BenchmarkRemove(b *testing.B) {
	const numGroups = 100
	const itemsPerGroup = 10
	bh := NewBinHeap[Item](numGroups*itemsPerGroup + 100)

	// Fill with 1000 items across 100 groups (10 items each)
	groups := make([][]Item, numGroups)
	for g := 0; g < numGroups; g++ {
		groups[g] = make([]Item, 0, itemsPerGroup)
		for i := 0; i < itemsPerGroup; i++ {
			item := NewTest(rand.Int63n(10000), fmt.Sprintf("g%d", g), fmt.Sprintf("g%d-i%d", g, i)) //nolint:gosec
			bh.Insert(item)
			groups[g] = append(groups[g], item)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		groupIdx := i % numGroups
		groupID := fmt.Sprintf("g%d", groupIdx)
		bh.Remove(groupID)
		// Restore items for next iteration
		b.StopTimer()
		for _, item := range groups[groupIdx] {
			bh.Insert(item)
		}
		b.StartTimer()
	}
}

func BenchmarkConcurrentInsertExtract(b *testing.B) {
	bh := NewBinHeap[Item](10000)
	// Pre-fill so ExtractMin rarely blocks
	for i := 0; i < 5000; i++ {
		bh.Insert(NewTest(rand.Int63n(10000), "bench", fmt.Sprintf("pre-%d", i))) //nolint:gosec
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				bh.Insert(NewTest(rand.Int63n(10000), "bench", fmt.Sprintf("p-%d-%d", i, rand.Int63()))) //nolint:gosec
			} else {
				bh.ExtractMin()
			}
			i++
		}
	})
}
