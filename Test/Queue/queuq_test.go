package main_test

import (
	queue "GoSTL/Queue"
	"fmt"
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := queue.NewQueue[int]()
	if !q.Empty() {
		t.Error("New queue should be empty")
	}
	if q.Len() != 0 {
		t.Error("New queue length should be 0")
	}
	if q.Capacity() < 8 {
		t.Errorf("Expected capacity >= 8, got %d", q.Capacity())
	}
}

func TestQueueInit(t *testing.T) {
	q := queue.NewQueue[string]()
	q.Init(5) // Should use minimum capacity (8)
	if q.Capacity() < 8 {
		t.Errorf("Init(5) should set capacity >= 8, got %d", q.Capacity())
	}

	q.Init(20)
	if q.Capacity() != 20 {
		t.Errorf("Init(20) should set capacity to 20, got %d", q.Capacity())
	}
}

func TestQueuePushPop(t *testing.T) {
	q := queue.NewQueue[int]()

	// Test basic push/pop
	q.Push(1)
	q.Push(2)
	q.Push(3)

	if q.Len() != 3 {
		t.Errorf("Expected length 3, got %d", q.Len())
	}

	val, ok := q.Pop()
	if !ok || val != 1 {
		t.Errorf("Expected (1, true), got (%d, %v)", val, ok)
	}

	val, ok = q.Pop()
	if !ok || val != 2 {
		t.Errorf("Expected (2, true), got (%d, %v)", val, ok)
	}

	// Test empty queue
	q.Pop()
	q.Pop()
	_, ok = q.Pop()
	if ok {
		t.Error("Expected false when popping from empty queue")
	}
}

func TestQueueFront(t *testing.T) {
	q := queue.NewQueue[string]()

	// Test empty queue
	_, ok := q.Front()
	if ok {
		t.Error("Expected false when getting front from empty queue")
	}

	q.Push("first")
	q.Push("second")

	front, ok := q.Front()
	if !ok || front != "first" {
		t.Errorf("Expected ('first', true), got (%q, %v)", front, ok)
	}
}

func TestQueueCapacityGrowth(t *testing.T) {
	q := queue.NewQueue[int]()
	initialCap := q.Capacity()

	// Fill beyond initial capacity
	for i := 0; i < initialCap+1; i++ {
		q.Push(i)
	}

	if q.Capacity() <= initialCap {
		t.Error("Capacity should grow when needed")
	}
}

func TestQueueAtAndSet(t *testing.T) {
	q := queue.NewQueue[int]()
	for i := 0; i < 5; i++ {
		q.Push(i * 10)
	}

	// Test positive indices
	val, ok := q.At(0)
	if !ok || val != 0 {
		t.Errorf("At(0) failed: ok=%v, val=%d", ok, val)
	}

	val, ok = q.At(2)
	if !ok || val != 20 {
		t.Errorf("At(2) failed: ok=%v, val=%d", ok, val)
	}

	// Test negative indices
	val, ok = q.At(-1)
	if !ok || val != 40 {
		t.Errorf("At(-1) failed: ok=%v, val=%d", ok, val)
	}

	val, ok = q.At(-2)
	if !ok || val != 30 {
		t.Errorf("At(-2) failed: ok=%v, val=%d", ok, val)
	}

	// Test Set
	if !q.Set(1, 99) {
		t.Error("Set(1) failed")
	}
	val, ok = q.At(1)
	if !ok || val != 99 {
		t.Errorf("Set/At failed: ok=%v, val=%d", ok, val)
	}

	// Test out of bounds
	_, ok = q.At(10)
	if ok {
		t.Error("At(10) should fail")
	}
	if q.Set(10, 0) {
		t.Error("Set(10) should fail")
	}
}

func TestQueueSwap(t *testing.T) {
	q := queue.NewQueue[string]()
	q.Push("a")
	q.Push("b")
	q.Push("c")

	if !q.Swap(0, 2) {
		t.Error("Swap failed")
	}

	front, _ := q.Front()
	if front != "c" {
		t.Errorf("After swap, front should be 'c', got %q", front)
	}

	val, _ := q.At(2)
	if val != "a" {
		t.Errorf("After swap, index 2 should be 'a', got %q", val)
	}

	// Test invalid swap
	if q.Swap(0, 3) {
		t.Error("Swap(0,3) should fail")
	}
}

func TestQueueRotate(t *testing.T) {
	q := queue.NewQueue[int]()
	for i := 1; i <= 5; i++ {
		q.Push(i)
	}

	// Rotate right
	q.Rotate(2)
	front, _ := q.Front()
	if front != 4 {
		t.Errorf("After rotate right 2, front should be 4, got %d", front)
	}

	// Rotate left
	q.Rotate(-3)
	front, _ = q.Front()
	if front != 2 {
		t.Errorf("After rotate left 3, front should be 2, got %d", front)
	}
}

func TestQueueCopy(t *testing.T) {
	q := queue.NewQueue[int]()
	q.Push(1)
	q.Push(2)
	q.Push(3)

	qCopy := q.Copy()
	if q.Len() != qCopy.Len() {
		t.Error("Copy should have same length")
	}

	// Modify original
	q.Pop()
	if q.Len() == qCopy.Len() {
		t.Error("Modifying original should not affect copy")
	}

	// Verify copy contents
	val, _ := qCopy.At(0)
	if val != 1 {
		t.Errorf("Copy should have same elements, got %d", val)
	}
}

func TestQueueReverse(t *testing.T) {
	q := queue.NewQueue[int]()
	for i := 1; i <= 3; i++ {
		q.Push(i)
	}

	q.Reverse()

	front, _ := q.Front()
	if front != 3 {
		t.Errorf("After reverse, front should be 3, got %d", front)
	}

	back, _ := q.At(-1)
	if back != 1 {
		t.Errorf("After reverse, back should be 1, got %d", back)
	}
}

func TestQueueClear(t *testing.T) {
	q := queue.NewQueue[bool]()
	for i := 0; i < 5; i++ {
		q.Push(true)
	}

	q.Clear()
	if !q.Empty() {
		t.Error("After Clear, queue should be empty")
	}
	if q.Len() != 0 {
		t.Error("After Clear, length should be 0")
	}
}

func TestQueueShrinkToFit(t *testing.T) {
	q := queue.NewQueue[int]()
	initialCap := q.Capacity()

	// Fill and then empty the queue
	for i := 0; i < initialCap+1; i++ {
		q.Push(i)
	}
	for i := 0; i < initialCap+1; i++ {
		q.Pop()
	}

	q.ShrinkToFit()
	if q.Capacity() != initialCap {
		t.Errorf("After ShrinkToFit, capacity should be %d, got %d", initialCap, q.Capacity())
	}
}

func TestQueueString(t *testing.T) {
	q := queue.NewQueue[int]()
	q.Push(1)
	q.Push(2)
	q.Push(3)

	str := fmt.Sprint(q)
	expected := "[1 2 3]"
	if str != expected {
		t.Errorf("String() = %q, want %q", str, expected)
	}
}

func TestQueueEdgeCases(t *testing.T) {
	// Test zero capacity initialization
	q := queue.NewQueue[float64]()
	q.Init(0)
	if q.Capacity() < 8 {
		t.Errorf("Init(0) should set capacity >= 8, got %d", q.Capacity())
	}

	// Test empty queue operations
	_, ok := q.Front()
	if ok {
		t.Error("Front should fail on empty queue")
	}
	_, ok = q.Pop()
	if ok {
		t.Error("Pop should fail on empty queue")
	}

	// Test with different types
	strQueue := queue.NewQueue[string]()
	strQueue.Push("test")
	if strQueue.Len() != 1 {
		t.Error("String queue length mismatch")
	}

	// Test with struct type
	type point struct{ x, y int }
	pointQueue := queue.NewQueue[point]()
	pointQueue.Push(point{1, 2})
	if pointQueue.Len() != 1 {
		t.Error("Struct queue length mismatch")
	}
}
