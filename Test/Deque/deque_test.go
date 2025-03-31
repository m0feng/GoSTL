package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"GoSTL/Deque"
)

func TestNewDeque(t *testing.T) {
	q := Deque.NewDeque[int]()
	if q == nil {
		t.Fatal("NewDeque returned nil")
	}
	if q.Len() != 0 {
		t.Errorf("New deque should be empty, got length %d", q.Len())
	}

	q = Deque.NewDeque[int](16)
	if q.Capacity() < 16 {
		t.Errorf("Expected capacity >= 16, got %d", q.Capacity())
	}
}

func TestPushPopFrontBack(t *testing.T) {
	q := Deque.NewDeque[int]()

	// Push back and pop front
	for i := 0; i < 100; i++ {
		q.PushBack(i)
	}
	for i := 0; i < 100; i++ {
		val, ok := q.PopFront()
		if !ok || val != i {
			t.Errorf("Expected %d, got %d (ok: %v)", i, val, ok)
		}
	}

	// Push front and pop back
	for i := 0; i < 100; i++ {
		q.PushFront(i)
	}
	for i := 0; i < 100; i++ {
		val, ok := q.PopBack()
		if !ok || val != i {
			t.Errorf("Expected %d, got %d (ok: %v)", i, val, ok)
		}
	}

	// Mixed operations
	q.PushBack(1)
	q.PushBack(2)
	q.PushFront(0)
	q.PushBack(3)
	q.PushFront(-1)

	expected := []int{-1, 0, 1, 2, 3}
	for _, exp := range expected {
		val, ok := q.PopFront()
		if !ok || val != exp {
			t.Errorf("Expected %d, got %d (ok: %v)", exp, val, ok)
		}
	}

	// Test empty pops
	if _, ok := q.PopFront(); ok {
		t.Error("PopFront on empty deque should return false")
	}
	if _, ok := q.PopBack(); ok {
		t.Error("PopBack on empty deque should return false")
	}
}

func TestFrontBack(t *testing.T) {
	q := Deque.NewDeque[string]()

	// Test empty
	if _, ok := q.Front(); ok {
		t.Error("Front on empty deque should return false")
	}
	if _, ok := q.Back(); ok {
		t.Error("Back on empty deque should return false")
	}

	q.PushBack("first")
	q.PushBack("second")

	// Test front
	if val, ok := q.Front(); !ok || val != "first" {
		t.Errorf("Expected 'first', got '%s' (ok: %v)", val, ok)
	}

	// Test back
	if val, ok := q.Back(); !ok || val != "second" {
		t.Errorf("Expected 'second', got '%s' (ok: %v)", val, ok)
	}

	// Test after pop
	q.PopFront()
	if val, ok := q.Front(); !ok || val != "second" {
		t.Errorf("Expected 'second', got '%s' (ok: %v)", val, ok)
	}
}

func TestLenEmpty(t *testing.T) {
	q := Deque.NewDeque[float64]()

	if !q.Empty() {
		t.Error("New deque should be empty")
	}

	for i := 0; i < 1000; i++ {
		q.PushBack(float64(i))
		if q.Len() != i+1 {
			t.Errorf("Expected length %d, got %d", i+1, q.Len())
		}
	}

	for i := 0; i < 1000; i++ {
		q.PopFront()
		if q.Len() != 999-i {
			t.Errorf("Expected length %d, got %d", 999-i, q.Len())
		}
	}

	if !q.Empty() {
		t.Error("Deque should be empty after all pops")
	}
}

func TestResize(t *testing.T) {
	q := Deque.NewDeque[int](4) // Small initial capacity to force resizes

	// Fill and check capacity growth
	initialCap := q.Capacity()
	for i := 0; i < 100; i++ {
		q.PushBack(i)
		if q.Len() != i+1 {
			t.Errorf("Length mismatch, expected %d got %d", i+1, q.Len())
		}
	}
	if q.Capacity() <= initialCap {
		t.Error("Capacity should have increased")
	}

	// Verify all elements
	for i := 0; i < 100; i++ {
		val, ok := q.PopFront()
		if !ok || val != i {
			t.Errorf("Expected %d, got %d (ok: %v)", i, val, ok)
		}
	}

	// Test front resizing
	for i := 0; i < 100; i++ {
		q.PushFront(i)
	}
	for i := 99; i >= 0; i-- {
		val, ok := q.PopFront()
		if !ok || val != i {
			t.Errorf("Expected %d, got %d (ok: %v)", i, val, ok)
		}
	}
}

func TestAt(t *testing.T) {
	q := Deque.NewDeque[rune]()

	// Fill with letters 'a' to 'z'
	for i := 0; i < 26; i++ {
		q.PushBack(rune('a' + i))
	}

	// Test positive indices
	for i := 0; i < 26; i++ {
		val, ok := q.At(i)
		expected := rune('a' + i)
		if !ok || val != expected {
			t.Errorf("At(%d) expected %c, got %c (ok: %v)", i, expected, val, ok)
		}
	}

	// Test negative indices
	for i := -1; i >= -26; i-- {
		val, ok := q.At(i)
		expected := rune('z' + i + 1)
		if !ok || val != expected {
			t.Errorf("At(%d) expected %c, got %c (ok: %v)", i, expected, val, ok)
		}
	}

	// Test out of bounds
	if _, ok := q.At(26); ok {
		t.Error("At(26) should be out of bounds")
	}
	if _, ok := q.At(-27); ok {
		t.Error("At(-27) should be out of bounds")
	}
}

func TestSet(t *testing.T) {
	q := Deque.NewDeque[int]()

	// Fill with 0..9
	for i := 0; i < 10; i++ {
		q.PushBack(i)
	}

	// Set some values
	if !q.Set(2, 20) {
		t.Error("Set(2, 20) failed")
	}
	if !q.Set(-1, 90) { // last element
		t.Error("Set(-1, 90) failed")
	}

	// Verify
	if val, _ := q.At(2); val != 20 {
		t.Errorf("Expected 20 at index 2, got %d", val)
	}
	if val, _ := q.At(9); val != 90 {
		t.Errorf("Expected 90 at index 9, got %d", val)
	}

	// Test out of bounds
	if q.Set(10, 100) {
		t.Error("Set(10, 100) should fail")
	}
	if q.Set(-11, 100) {
		t.Error("Set(-11, 100) should fail")
	}
}

func TestSwap(t *testing.T) {
	q := Deque.NewDeque[string]()

	// Fill with some values
	values := []string{"a", "b", "c", "d", "e"}
	for _, v := range values {
		q.PushBack(v)
	}

	// Swap some elements
	if !q.Swap(0, 4) {
		t.Error("Swap(0, 4) failed")
	}
	if !q.Swap(1, 3) {
		t.Error("Swap(1, 3) failed")
	}
	if !q.Swap(-1, -2) { // last and second last
		t.Error("Swap(-1, -2) failed")
	}

	// Verify
	expected := []string{"e", "d", "c", "a", "b"}
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("At(%d) expected %s, got %s", i, exp, val)
		}
	}

	// Test invalid swaps
	if q.Swap(0, 5) {
		t.Error("Swap(0, 5) should fail")
	}
	if q.Swap(-1, -6) {
		t.Error("Swap(-1, -6) should fail")
	}

	if q.Swap(2, 2) { // swapping with self
		t.Error("Swap(2, 2) should fail or be no-op")
	}
}

func TestRotate(t *testing.T) {
	q := Deque.NewDeque[int]()

	// Fill with 0..9
	for i := 0; i < 10; i++ {
		q.PushBack(i)
	}

	// Rotate right by 3
	q.Rotate(3)
	expected := []int{7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("After rotate(3), At(%d) expected %d, got %d", i, exp, val)
		}
	}

	// Rotate left by 5 (same as right by -5)
	q.Rotate(-5)
	expected = []int{2, 3, 4, 5, 6, 7, 8, 9, 0, 1}
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("After rotate(-5), At(%d) expected %d, got %d", i, exp, val)
		}
	}

	// Rotate by length (should be no-op)
	q.Rotate(10)
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("After rotate(10), At(%d) expected %d, got %d", i, exp, val)
		}
	}

	// Rotate by more than length
	q.Rotate(23) // 23 % 10 = 3
	expected = []int{9, 0, 1, 2, 3, 4, 5, 6, 7, 8}
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("After rotate(23), At(%d) expected %d, got %d", i, exp, val)
		}
	}
}

func TestReverse(t *testing.T) {
	q := Deque.NewDeque[int]()

	// Test empty and single element
	q.Reverse() // should do nothing
	q.PushBack(1)
	q.Reverse()
	if val, _ := q.Front(); val != 1 {
		t.Error("Reverse on single element should not change it")
	}

	// Test even number of elements
	q.Clear()
	for i := 0; i < 6; i++ {
		q.PushBack(i)
	}
	q.Reverse()
	expected := []int{5, 4, 3, 2, 1, 0}
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("After reverse, At(%d) expected %d, got %d", i, exp, val)
		}
	}

	// Test odd number of elements
	q.Clear()
	for i := 0; i < 5; i++ {
		q.PushBack(i)
	}
	q.Reverse()
	expected = []int{4, 3, 2, 1, 0}
	for i, exp := range expected {
		if val, _ := q.At(i); val != exp {
			t.Errorf("After reverse, At(%d) expected %d, got %d", i, exp, val)
		}
	}
}

func TestShrinkToFit(t *testing.T) {
	q := Deque.NewDeque[int](64)
	initCap := 64
	for i := 0; i < 32; i++ {
		q.PushBack(i)
	}

	for i := 0; i < 16; i++ {
		q.PopFront()
	}

	capBefore := q.Capacity()
	q.ShrinkToFit()
	capAfter := q.Capacity()

	if capAfter > capBefore {
		t.Errorf("Expected capacity to shrink, before: %d, after: %d", capBefore, capAfter)
	}
	if q.Len() != 16 {
		t.Errorf("Expected length 16 after shrink, got %d", q.Len())
	}

	// Verify elements
	for i := 0; i < 16; i++ {
		val, _ := q.At(i)
		if val != i+16 {
			t.Errorf("At(%d) expected %d, got %d", i, i+16, val)
		}
	}

	// Test with empty deque
	q.Clear()
	q.ShrinkToFit()
	if q.Capacity() != initCap {
		t.Errorf("Empty deque should shrink to initCap (%d), got %d", initCap, q.Capacity())
	}
}

func TestClear(t *testing.T) {
	q := Deque.NewDeque[int]()

	// Fill with data
	for i := 0; i < 100; i++ {
		q.PushBack(i)
	}

	q.Clear()
	if q.Len() != 0 {
		t.Errorf("After clear, length should be 0, got %d", q.Len())
	}
	if !q.Empty() {
		t.Error("After clear, deque should be empty")
	}

	// Should be able to reuse
	for i := 0; i < 10; i++ {
		q.PushBack(i)
	}
	if q.Len() != 10 {
		t.Errorf("After reuse, expected length 10, got %d", q.Len())
	}
}

func TestCopy(t *testing.T) {
	q := Deque.NewDeque[int]()

	// Fill with data
	for i := 0; i < 100; i++ {
		q.PushBack(i)
	}

	// Make a copy
	copyQ := q.Copy()

	// Modify original
	for i := 0; i < 50; i++ {
		q.PopFront()
	}
	for i := 100; i < 150; i++ {
		q.PushBack(i)
	}

	// Verify copy is unchanged
	if copyQ.Len() != 100 {
		t.Errorf("Copy length should be 100, got %d", copyQ.Len())
	}
	for i := 0; i < 100; i++ {
		val, _ := copyQ.At(i)
		if val != i {
			t.Errorf("Copy At(%d) expected %d, got %d", i, i, val)
		}
	}

	// Verify original is modified as expected
	if q.Len() != 100 {
		t.Errorf("Original length should be 100, got %d", q.Len())
	}
	for i := 0; i < 100; i++ {
		val, _ := q.At(i)
		expected := 50 + i
		if val != expected {
			t.Errorf("Original At(%d) expected %d, got %d", i, expected, val)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	q := Deque.NewDeque[int]()
	var wg sync.WaitGroup
	count := 1000

	// Concurrent writers
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			q.PushBack(i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			q.PushFront(-i)
		}
	}()

	// Concurrent reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < count*2; i++ {
			q.Len()
			runtime.Gosched()
		}
	}()

	wg.Wait()

	// Verify final length
	if q.Len() != count*2 {
		t.Errorf("Expected length %d, got %d", count*2, q.Len())
	}
}

func TestFormat(t *testing.T) {
	q := Deque.NewDeque[int]()
	for i := 0; i < 5; i++ {
		q.PushBack(i)
	}

	// Test default format
	expected := "[0 1 2 3 4]"
	if s := fmt.Sprintf("%v", q); s != expected {
		t.Errorf("Format %v expected %q, got %q", "%v", expected, s)
	}

	// Test width limit
	expected = "[0 1 2 ...+2]"
	if s := fmt.Sprintf("%5v", q); s != expected {
		t.Errorf("Format %v expected %q, got %q", "%5v", expected, s)
	}

	// Test precision limit
	expected = "[0 1 ...+3]"
	if s := fmt.Sprintf("%.2v", q); s != expected {
		t.Errorf("Format %v expected %q, got %q", "%.2v", expected, s)
	}

	// Test empty deque
	q.Clear()
	expected = "[]"
	if s := fmt.Sprintf("%v", q); s != expected {
		t.Errorf("Format %v expected %q, got %q", "%v", expected, s)
	}
}

func BenchmarkPushPop(b *testing.B) {
	q := Deque.NewDeque[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.PushBack(i)
		q.PopFront()
	}
}

func BenchmarkConcurrentPushPop(b *testing.B) {
	q := Deque.NewDeque[int]()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			if r.Intn(2) == 0 {
				q.PushBack(1)
			} else {
				q.PopFront()
			}
		}
	})
}

func BenchmarkRotate(b *testing.B) {
	q := Deque.NewDeque[int]()
	for i := 0; i < 1000; i++ {
		q.PushBack(i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Rotate(1)
	}
}
